package ws

import (
	"context"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
)

// Server implements handlers for websocket connections and queries
type Server interface {
	KeepAlive(numClients int, elapsed time.Duration)
	InitClient(client *Client, payload map[string]interface{}) error
	CloseClient(client *Client)
	StartQuery(client *Client, id QueryID, payload map[string]interface{}) error
	StopQuery(client *Client, id QueryID) error
}

// Broker maintains the set of active clients and routes messages to the clients
type Broker struct {
	// context used by the broker
	ctx context.Context

	// server implementation
	server Server

	// used to manage websocket
	upgrader *websocket.Upgrader

	// use to register a new client with the broker
	register chan *Client

	// use to remove a client from the broker
	unregister chan *Client

	// use to submit a new request from the client for processing
	requests chan clientRequest

	// used to send keep alive messages regularly
	keepAliveTicker *time.Ticker

	// Registered clients
	clients map[*Client]bool
}

// clientRequest bundles a client and a WebsocketMessage
type clientRequest struct {
	Client  *Client
	Message websocketMessage
}

const (
	connectionKeepAliveInterval = 10 * time.Second
)

// NewBroker initializes a new Broker
func NewBroker(server Server) *Broker {
	// create upgrader
	upgrader := &websocket.Upgrader{
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// create broker
	broker := &Broker{
		ctx:             context.Background(),
		server:          server,
		upgrader:        upgrader,
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		requests:        make(chan clientRequest),
		keepAliveTicker: time.NewTicker(connectionKeepAliveInterval),
		clients:         make(map[*Client]bool),
	}

	// initialize run loop
	go broker.runForever()

	// done
	return broker
}

// HTTPHandler upgrades incoming requests to websocket connections
func (b *Broker) HTTPHandler(w http.ResponseWriter, r *http.Request) error {
	// chimiddleware.WrapResponseWriter combined with chimiddleware.DefaultCompress
	// interferes with Upgrade causing an error. Hack: unwrap chi wrapper (side effect:
	// disables response size tracking -- doesn't matter here)
	if ww, ok := w.(chimiddleware.WrapResponseWriter); ok {
		w = ww.Unwrap()
	}

	// upgrade to a websocket connection
	ws, err := b.upgrader.Upgrade(w, r, http.Header{
		"Sec-Websocket-Protocol": []string{"graphql-ws"},
	})
	if err != nil {
		return err
	}

	// create client and register it with the hub
	b.register <- newClient(b, ws)

	return nil
}

// runForever is the main run loop of the broker
func (b *Broker) runForever() {
	for {
		select {
		case client := <-b.register:
			b.processRegister(client)
		case client := <-b.unregister:
			b.processUnregister(client)
		case req := <-b.requests:
			b.processRequest(req)
		case <-b.keepAliveTicker.C:
			b.sendKeepAlive()
		}
	}
}

// sendKeepAlive broadcasts keepalive message to all clients
func (b *Broker) sendKeepAlive() {
	if len(b.clients) == 0 {
		return
	}
	startTime := time.Now()
	for client := range b.clients {
		client.sendKeepAlive()
	}
	elapsed := time.Since(startTime)
	b.server.KeepAlive(len(b.clients), elapsed)
}

// handle a new client
func (b *Broker) processRegister(c *Client) {
	b.clients[c] = true
}

// remove a client and its subscriptions
func (b *Broker) processUnregister(c *Client) {
	if !b.clients[c] {
		return
	}
	delete(b.clients, c)
	c.finalizeClose()
}

// process a client request
func (b *Broker) processRequest(r clientRequest) {
	// use the request's message type to process it accordingly
	switch r.Message.Type {
	case connectionInitMsgType:
		b.processInitRequest(r)
	case connectionTerminateMsgType:
		r.Client.terminateConnection(websocket.CloseNormalClosure, "terminated")
	case startMsgType:
		b.processStartRequest(r)
	case stopMsgType:
		b.processStopRequest(r)
	default:
		r.Client.sendConnectionError("unrecognized message type")
		r.Client.terminateConnection(websocket.CloseProtocolError, "unrecognized message type")
	}
}

// process a connectionInitMsgType
func (b *Broker) processInitRequest(r clientRequest) {
	err := b.server.InitClient(r.Client, r.Message.Payload)
	if err != nil {
		r.Client.sendConnectionError(err.Error())
	} else {
		r.Client.sendConnectionAck()
	}
}

// process a startMsgType
func (b *Broker) processStartRequest(r clientRequest) {
	r.Client.startQuery(r.Message.ID, r.Message.Payload)
}

// process a stopMsgType
func (b *Broker) processStopRequest(r clientRequest) {
	r.Client.stopQuery(r.Message.ID)
}
