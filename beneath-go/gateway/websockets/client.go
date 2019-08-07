package websockets

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/beneath-core/beneath-go/control/model"
	uuid "github.com/satori/go.uuid"

	"github.com/gorilla/websocket"
)

// SubscriptionID is an identifer passed by a user to distinguish between data when they
// have multiple subscriptions
type SubscriptionID string

// SubscriptionFilter identifies the instance ID subscribed to by a user
// FUTURE: Add info to filter by key prefix
type SubscriptionFilter uuid.UUID

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// The broker that the client belongs to
	Broker *Broker

	// The websocket connection
	WS *websocket.Conn

	// Key used by client to authenticate
	Key *model.Key

	// Buffered channel of outbound messages to users
	Outbound chan WebsocketMessage

	// Tracks the client's subscriptions to instanceIDs (instanceID -> subscription ID)
	Subscriptions map[SubscriptionFilter]SubscriptionID

	// Lock
	mu sync.Mutex
}

// WebsocketMessage represents a message passed over the wire
type WebsocketMessage struct {
	Type    string                 `json:"type"`
	ID      SubscriptionID         `json:"id,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"` // json.RawMessage?
}

// List of potential values for WebsocketMessage.Type
// See: https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md
const (
	connectionInitMsgType      = "connection_init"      // Client -> Server
	connectionTerminateMsgType = "connection_terminate" // Client -> Server
	startMsgType               = "start"                // Client -> Server
	stopMsgType                = "stop"                 // Client -> Server
	connectionAckMsgType       = "connection_ack"       // Server -> Client
	connectionErrorMsgType     = "connection_error"     // Server -> Client
	dataMsgType                = "data"                 // Server -> Client
	errorMsgType               = "error"                // Server -> Client
	completeMsgType            = "complete"             // Server -> Client
	connectionKeepAliveMsgType = "ka"                   // Server -> Client
)

// Configuration settings for websocket
const (
	// Maximum message size allowed from client
	maxMessageSize = 256
)

// NewClient initializes a new client
func NewClient(broker *Broker, ws *websocket.Conn, key *model.Key) *Client {
	// set WS connection configuration
	ws.SetReadLimit(maxMessageSize)

	client := &Client{
		Broker:        broker,
		WS:            ws,
		Key:           key,
		Outbound:      make(chan WebsocketMessage, 16),
		Subscriptions: make(map[SubscriptionFilter]SubscriptionID),
	}

	// start background workers
	go client.beginReading()
	go client.beginWriting()

	log.Printf("New client connected. Client IP: %v", client.WS.RemoteAddr())

	return client
}

// SendConnectionError sends a GQL_CONNECTION_ERROR WebsocketMessage
func (c *Client) SendConnectionError(msg string) {
	c.Outbound <- WebsocketMessage{
		Type: connectionErrorMsgType,
		Payload: map[string]interface{}{
			"message": msg,
		},
	}
}

// SendConnectionAck sends a GQL_CONNECTION_ACK WebsocketMessage
func (c *Client) SendConnectionAck() {
	c.Outbound <- WebsocketMessage{
		Type: connectionAckMsgType,
	}
}

// SendData sends a GQL_DATA WebsocketMessage
func (c *Client) SendData(id SubscriptionID, data map[string]interface{}) {
	c.Outbound <- WebsocketMessage{
		Type: dataMsgType,
		ID:   id,
		Payload: map[string]interface{}{
			"data": data,
		},
	}
}

// SendError sends a GQL_ERROR WebsocketMessage
func (c *Client) SendError(id SubscriptionID, msg string) {
	c.Outbound <- WebsocketMessage{
		Type: errorMsgType,
		ID:   id,
		Payload: map[string]interface{}{
			"message": msg,
		},
	}
}

// SendComplete sends a GQL_COMPLETE WebsocketMessage
func (c *Client) SendComplete(id SubscriptionID) {
	c.Outbound <- WebsocketMessage{
		Type: completeMsgType,
		ID:   id,
	}
}

// SendKeepAlive sends a GQL_CONNECTION_KEEP_ALIVE WebsocketMessage
func (c *Client) SendKeepAlive() {
	c.Outbound <- WebsocketMessage{
		Type: connectionKeepAliveMsgType,
	}
}

// TrackSubscription checks the subscription ID and tracks it in the client's bookkeeping
func (c *Client) TrackSubscription(filter SubscriptionFilter, subID SubscriptionID) error {
	c.mu.Lock()
	for _, id := range c.Subscriptions {
		if id == subID {
			return fmt.Errorf("Error! You are already using that ID")
		}
	}
	c.Subscriptions[filter] = subID
	c.mu.Unlock()
	return nil
}

// UntrackSubscription removes the subscription form the client's bookkeeping and returns
// the filter that the subscription was indexed under
func (c *Client) UntrackSubscription(subID SubscriptionID) (filter SubscriptionFilter, ok bool) {
	c.mu.Lock()

	// look for subID
	for f, id := range c.Subscriptions {
		if id == subID {
			filter = f
			ok = true
		}
	}

	if ok {
		delete(c.Subscriptions, filter)
	}

	c.mu.Unlock()

	return filter, ok
}

// GetSubscriptionID safely accesses c.Subscriptions
func (c *Client) GetSubscriptionID(filter SubscriptionFilter) SubscriptionID {
	c.mu.Lock()
	subID := c.Subscriptions[filter]
	c.mu.Unlock()
	return subID
}

// Terminate gracefully closes the connection
func (c *Client) Terminate(closeCode int, msg string) {
	c.mu.Lock()
	_ = c.WS.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, msg))
	c.mu.Unlock()
}

// Close decomissions the client and should only be called by the broker during unregister
func (c *Client) Close() {
	close(c.Outbound)
	_ = c.WS.Close()
	log.Printf("Closed connection. Client IP: %v", c.WS.RemoteAddr())
}

// beginReading relays requests from the websocket connection to the broker
func (c *Client) beginReading() {
	// close the client if the function stops for some reason
	defer c.Broker.CloseClient(c)

	for {
		// read a message from the client's websocket
		_, data, err := c.WS.ReadMessage()
		if err != nil {
			// break out of the run loop (triggering CloseClient) if the connection closed
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseProtocolError, websocket.CloseNoStatusReceived) {
				log.Printf("Client closed unexpectedly: %v", err)
			}
			return
		}

		// parse message as WebsocketMessage
		var msg WebsocketMessage
		err = json.Unmarshal(data, &msg)
		if err != nil {
			c.SendConnectionError("couldn't parse message as json")
			c.Terminate(websocket.CloseProtocolError, "couldn't parse message as json")
			return
		}

		// log information about client message
		log.Printf("Message received from client. Client IP: %v, Message Type: %v, Message Size: %d bytes", c.WS.RemoteAddr(), msg.Type, len(data))

		// send user request to broker
		c.Broker.requests <- Request{
			Client:  c,
			Message: msg,
		}
	}
}

// beginWriting relays messages from the broker to the websocket
func (c *Client) beginWriting() {
	// close the client if the function stops for some reason
	defer c.Broker.CloseClient(c)

	for {
		select {
		// process a message from the broker
		case msg, ok := <-c.Outbound:
			if !ok {
				// the channel was closed (client is gone)
				return
			}

			// open a writer for the next message to send
			c.mu.Lock()
			w, err := c.WS.NextWriter(websocket.TextMessage)
			if err != nil {
				// breaking out of run loop -- closing the client
				c.mu.Unlock()
				return
			}

			// writes a to writer
			fn := func(msg WebsocketMessage) {
				// encode message as json
				data, err := json.Marshal(msg)
				if err != nil {
					c.mu.Unlock()
					log.Panicf("couldn't marshal WebsocketMessage: %v", msg)
					return
				}

				// write message
				w.Write(data)
				w.Write([]byte("\n"))
			}

			// efficiency gain: empty the Send channel now that we have an open writer
			fn(msg)
			for i := 0; i < len(c.Outbound); i++ {
				fn(<-c.Outbound)
			}

			// close the writer; flush the complete message to the network
			if err := w.Close(); err != nil {
				// breaking out of run loop -- closing the client
				c.mu.Unlock()
				return
			}

			c.mu.Unlock()
		}
	}
}
