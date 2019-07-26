package gateway

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages to users.
	send chan []byte

	// Buffered channel of inbound messages from users.
	receive chan []byte
}

// Packet bundles the client's message with the clientID
type Packet struct {
	message  []byte
	clientID *Client
}

// readPump relays messages from the websocket connection to the hub.
func (c *Client) readPump() {
	// close the client if the function stops for some reason
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// optional settings
	// c.conn.SetReadLimit(maxMessageSize)
	// c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		// read a message from the client's websocket
		_, message, err := c.conn.ReadMessage()
		log.Printf("reading message from client...")

		// handle error
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// bundle the client's message with a client identifier
		packet := Packet{message: message, clientID: c}

		// send the message + clientID to the hub
		c.hub.receive <- packet
	}
}

// writePump relays messages from the hub to the websocket connection.
func (c *Client) writePump() {
	// close the client if the function stops for some reason
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		// send a message from the client to the websocket
		case message, ok := <-c.send:
			// c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Printf("the hub closed the channel")
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// open a writer for the next message to send
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			// close the writer; flush the complete message to the network
			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the client
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// upgrade to a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// create a client and register it with the hub
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// launch the client's read and write go routines
	go client.writePump()
	go client.readPump()
}

// EXTRA CODE

// const (
// 	// Time allowed to write a message to the peer.
// 	writeWait = 10 * time.Second

// 	// Time allowed to read the next pong message from the peer.
// 	pongWait = 60 * time.Second

// 	// Send pings to peer with this period. Must be less than pongWait.

// 	// Maximum message size allowed from peer.
// 	maxMessageSize = 512
// )
