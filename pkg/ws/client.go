package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

// QueryID is an identifer passed by a user to distinguish between data when they
// have multiple subscriptions
type QueryID string

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// Context for the client
	Context context.Context

	// MessagesRead tracks the number of messages sent by the client
	MessagesRead uint

	// BytesRead tracks the number of bytes sent by the client
	BytesRead uint

	// MessagesWritten tracks the number of messages sent to the client
	MessagesWritten uint

	// BytesWritten tracks the number of bytes sent to the client
	BytesWritten uint

	// StartTime tracks the server time the client connected
	StartTime time.Time

	// Initialized tracks if the client has sent a connectionInitMsgType message
	Initialized bool

	// Err is set if the client was closed unexpectedly
	Err error

	// broker that the client belongs to
	broker *Broker

	// ws is the websocket connection
	ws *websocket.Conn

	// state on the client
	clientState interface{}

	// state for each query
	queryState map[QueryID]interface{}

	// outbound is a buffered channel of messages that should be sent to the client
	outbound chan websocketMessage

	// terminate is a channel that can be used to manually terminate the connection
	terminate chan terminateMessage
}

// websocketMessage represents a message passed over the wire
type websocketMessage struct {
	Type    string                 `json:"type"`
	ID      QueryID                `json:"id,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"` // json.RawMessage?
}

// terminateMessage represents a manual termination of the connection
type terminateMessage struct {
	Code    int
	Message string
}

// List of potential values for websocketMessage.Type
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

// newClient initializes a new client
func newClient(broker *Broker, ws *websocket.Conn) *Client {
	// set WS connection configuration
	ws.SetReadLimit(maxMessageSize)

	client := &Client{
		Context:    context.Background(),
		StartTime:  time.Now(),
		broker:     broker,
		ws:         ws,
		queryState: make(map[QueryID]interface{}),
		outbound:   make(chan websocketMessage, 16),
		terminate:  make(chan terminateMessage),
	}

	// start background workers
	go client.beginReading()
	go client.beginWriting()

	return client
}

// SendData sends data for the given query encoded as {"data": data}
func (c *Client) SendData(id QueryID, data interface{}) {
	c.sendData(id, data)
}

// GetRemoteAddr returns the remote address of the client
func (c *Client) GetRemoteAddr() net.Addr {
	return c.ws.RemoteAddr()
}

// SetState saves arbitrary state that's persisted for the duration of the connection
func (c *Client) SetState(state interface{}) {
	c.clientState = state
}

// GetState returns the latest value assigned with SetState
func (c *Client) GetState() interface{} {
	return c.clientState
}

// SetQueryState saves arbitrary state that's persisted for the duration of the
func (c *Client) SetQueryState(id QueryID, state interface{}) {
	c.queryState[id] = state
}

// GetQueryState returns the latest value assigned with SetQueryState
func (c *Client) GetQueryState(id QueryID) interface{} {
	return c.queryState[id]
}

func (c *Client) startQuery(id QueryID, payload map[string]interface{}) {
	if _, ok := c.queryState[id]; ok {
		c.sendError(id, "message ID already in use")
		return
	}

	c.queryState[id] = nil
	err := c.broker.server.StartQuery(c, id, payload)
	if err != nil {
		delete(c.queryState, id)
		c.sendError(id, err.Error())
	}
}

func (c *Client) stopQuery(id QueryID) {
	if _, ok := c.queryState[id]; !ok {
		c.sendError(id, "message ID not initiated")
		return
	}

	err := c.broker.server.StopQuery(c, id)
	if err != nil {
		c.sendError(id, err.Error())
	} else {
		// send complete (as per spec)
		c.sendComplete(id)
	}

	delete(c.queryState, id)
}

// beginClose initiates closing of the client
func (c *Client) beginClose() {
	c.broker.unregister <- c
}

// finalizeClose cleans up the client
func (c *Client) finalizeClose() {
	for id := range c.queryState {
		c.stopQuery(id)
	}
	if c.Initialized {
		c.broker.server.CloseClient(c)
	}
	close(c.outbound)
	_ = c.ws.Close()
}

// sendConnectionError sends a GQL_CONNECTION_ERROR websocketMessage
func (c *Client) sendConnectionError(msg string) {
	c.outbound <- websocketMessage{
		Type: connectionErrorMsgType,
		Payload: map[string]interface{}{
			"message": msg,
		},
	}
}

// sendConnectionAck sends a GQL_CONNECTION_ACK websocketMessage
func (c *Client) sendConnectionAck() {
	c.outbound <- websocketMessage{
		Type: connectionAckMsgType,
	}
}

// sendData sends a GQL_DATA websocketMessage
func (c *Client) sendData(id QueryID, data interface{}) {
	c.outbound <- websocketMessage{
		Type: dataMsgType,
		ID:   id,
		Payload: map[string]interface{}{
			"data": data,
		},
	}
}

// sendError sends a GQL_ERROR websocketMessage
func (c *Client) sendError(id QueryID, msg string) {
	c.outbound <- websocketMessage{
		Type: errorMsgType,
		ID:   id,
		Payload: map[string]interface{}{
			"message": msg,
		},
	}
}

// sendComplete sends a GQL_COMPLETE websocketMessage
func (c *Client) sendComplete(id QueryID) {
	c.outbound <- websocketMessage{
		Type: completeMsgType,
		ID:   id,
	}
}

// sendKeepAlive sends a GQL_CONNECTION_KEEP_ALIVE websocketMessage
func (c *Client) sendKeepAlive() {
	c.outbound <- websocketMessage{
		Type: connectionKeepAliveMsgType,
	}
}

// terminateConnection gracefully closes the connection
func (c *Client) terminateConnection(closeCode int, msg string) {
	c.terminate <- terminateMessage{
		Code:    closeCode,
		Message: msg,
	}
}

// beginReading relays requests from the websocket connection to the broker
func (c *Client) beginReading() {
	// close the client if the function stops for some reason
	defer c.beginClose()

	for {
		// read a message from the client's websocket
		_, data, err := c.ws.ReadMessage()
		if err != nil {
			// break out of the run loop (triggering beginClose) if the connection closed
			c.Err = err
			return
		}

		// record metrics
		c.MessagesRead++
		c.BytesRead += uint(len(data))

		// parse message as WebsocketMessage
		var msg websocketMessage
		err = json.Unmarshal(data, &msg)
		if err != nil {
			c.Err = err
			c.sendConnectionError("couldn't parse message as json")
			c.terminateConnection(websocket.CloseProtocolError, "couldn't parse message as json")
			return
		}

		// send user request to broker
		c.broker.requests <- clientRequest{
			Client:  c,
			Message: msg,
		}
	}
}

// beginWriting relays messages from the broker to the websocket
func (c *Client) beginWriting() {
	// close the client if the function stops for some reason
	defer c.beginClose()

	for {
		select {
		// process a termination message
		case msg, ok := <-c.terminate:
			if !ok {
				// channel is gone
				return
			}
			_ = c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(msg.Code, msg.Message))
			// break the loop
			return
		// process a message from the broker
		case msg, ok := <-c.outbound:
			if !ok {
				// the channel was closed (client is gone)
				return
			}

			// encode message as json
			data, err := json.Marshal(msg)
			if err != nil {
				panic(fmt.Errorf("couldn't marshal WebsocketMessage: %v", msg))
			}

			// record metrics
			c.MessagesWritten++
			c.BytesWritten += uint(len(data))

			// write
			err = c.ws.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				// breaking out of run loop -- closing the client
				return
			}
		}
	}
}
