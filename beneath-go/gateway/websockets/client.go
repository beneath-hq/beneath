package websockets

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// The broker that the client belongs to
	Broker *Broker

	// The websocket connection
	WS *websocket.Conn

	// Buffered channel of outbound messages to users
	Outbound chan WebsocketMessage
}

// WebsocketMessage represents a message passed over the wire
type WebsocketMessage struct {
	Type    string `json:"type"`
	ID      string `json:"id,omitempty"`
	Payload string `json:"payload,omitempty"` // json.RawMessage?
}

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

// newClient initializes a new client
func newClient(broker *Broker, ws *websocket.Conn) *Client {
	// TODO: revisit WS connection configuration
	// ws.SetReadLimit(maxMessageSize)
	// ws.SetReadDeadline(time.Now().Add(pongWait))
	// ws.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	// ws.SetWriteDeadline(time.Now().Add(writeWait))

	client := &Client{
		Broker:   broker,
		WS:       ws,
		Outbound: make(chan WebsocketMessage, 256),
	}

	// start background workers
	go client.beginReading()
	go client.beginWriting()

	return client
}

// close decomissions the client
func (c *Client) close() {
	c.Broker.Unregister <- c
	c.WS.Close()
	close(c.Outbound)
	// TODO: Log closed connection
}

// beginReading relays requests from the websocket connection to the broker
func (c *Client) beginReading() {
	// close the client if the function stops for some reason
	defer c.close()

	for {
		// read a message from the client's websocket
		_, data, err := c.WS.ReadMessage()
		if err != nil {
			// break out of the run loop (triggering c.close) if the connection closed
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Client closed unexpectedly: %v", err)
			}
			return
		}

		// parse message as WebsocketMessage
		var msg WebsocketMessage
		err = json.Unmarshal(data, &msg)
		if err != nil {
			c.sendError("", "couldn't parse message as json")
			continue
		}

		// send user request to broker
		c.Broker.Requests <- Request{
			Client:  c,
			Message: msg,
		}
	}
}

// beginWriting relays messages from the broker to the websocket
func (c *Client) beginWriting() {
	// close the client if the function stops for some reason
	defer c.close()

	for {
		select {
		// process a message from the broker
		case msg, ok := <-c.Outbound:
			if !ok {
				// the channel was closed (client is gone)
				return
			}

			// open a writer for the next message to send
			w, err := c.WS.NextWriter(websocket.TextMessage)
			if err != nil {
				// breaking out of run loop -- closing the client
				return
			}

			// writes a to writer
			fn := func(msg WebsocketMessage) {
				// encode message as json
				data, err := json.Marshal(msg)
				if err != nil {
					log.Panicf("couldn't marshal WebsocketMessage: %v", msg)
					return
				}

				// write message
				w.Write(data)
			}

			// efficiency gain: empty the Send channel now that we have an open writer
			fn(msg)
			for i := 0; i < len(c.Outbound); i++ {
				fn(<-c.Outbound)
			}

			// close the writer; flush the complete message to the network
			if err := w.Close(); err != nil {
				// breaking out of run loop -- closing the client
				return
			}
		}
	}
}

// sends a data message type to the user
func (c *Client) sendData(id string, description string) {
	c.Outbound <- WebsocketMessage{
		Type:    dataMsgType,
		ID:      id,
		Payload: description,
	}
}

// sendError sends an error WebsocketMessage to the user
func (c *Client) sendError(id string, description string) {
	c.Outbound <- WebsocketMessage{
		Type:    errorMsgType,
		ID:      id,
		Payload: description,
	}
}
