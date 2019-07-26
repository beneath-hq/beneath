package websockets

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	uuid "github.com/satori/go.uuid"
)

// Request bundles the client's message with the clientID
type Request struct {
	Client  *Client
	Message WebsocketMessage
}

// Broker maintains the set of active clients and routes messages to the clients
type Broker struct {
	// use to register a new client with the broker
	Register chan *Client

	// use to remove a client from the broker
	Unregister chan *Client

	// use to submit a new request from the client for processing
	Requests chan Request

	// use to distribute records to subscrined clients
	Dispatch chan *Message

	// used to manage websocket
	upgrader *websocket.Upgrader

	// Registered clients.
	clients map[*Client]bool

	// Client subscriptions. Mapping: stream instance ID --> list of subscribed clients
	// TODO
	clientSubscriptions map[uuid.UUID][]*Client
}

// NewBroker initializes a new Broker
func NewBroker() *Broker {
	// create upgrader
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// create broker
	broker := &Broker{
		upgrader:            upgrader,
		clients:             make(map[*Client]bool),
		Register:            make(chan *Client),
		Unregister:          make(chan *Client),
		Requests:            make(chan Request),
		Dispatch:            make(chan *Message),
		clientSubscriptions: make(map[uuid.UUID][]*Client),
	}

	// initialize run loop
	go broker.runForever()

	// done
	return broker
}

// HTTPHandler upgrades incoming requests to websocket connections
func (b *Broker) HTTPHandler(w http.ResponseWriter, r *http.Request) error {
	// upgrade to a websocket connection
	ws, err := b.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	// create client and register it with the hub
	b.Register <- newClient(b, ws)

	return nil
}

func (b *Broker) runForever() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case client := <-b.Register:
			b.registerClient(client)
		case client := <-b.Unregister:
			b.unregisterClient(client)
		case req := <-b.Requests:
			b.processRequest(req)
		case msg := <-b.Dispatch:
			b.processMessage(msg)
		case <-ticker.C:
			b.broadcast(fmt.Sprint(rand.Intn(100), rand.Intn(100)))
		}
	}
}

// handle a new client
func (b *Broker) registerClient(c *Client) {
	b.clients[c] = true
}

// remove a client
func (b *Broker) unregisterClient(c *Client) {
	if _, ok := b.clients[c]; ok {
		delete(b.clients, c)
	}
}

// process a client request
func (b *Broker) processRequest(r Request) {
	log.Printf("Read new request: %v", r.Message)
	// TODO: Update broker state (dispatch tables) based on client's request
}

// process messages from pubsub (route to relevant clients)
func (b *Broker) processMessage(msg *Message) {
	// TODO: Dispatch message to relevant clients
}

// publish a string message to all clients
func (b *Broker) broadcast(msg string) {
	for client := range b.clients {
		client.sendData("", msg)
	}
}
