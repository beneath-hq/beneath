package gateway

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	writePeriod = 100 * time.Millisecond
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	receive chan Packet // should I make this a slice []Packet?

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Client subscriptions. Mapping: stream instance ID --> list of subscribed clients
	clientSubscriptions map[uuid.UUID][]*Client
}

func newHub() *Hub {
	return &Hub{
		clients:             make(map[*Client]bool),
		receive:             make(chan Packet),
		register:            make(chan *Client),
		unregister:          make(chan *Client),
		clientSubscriptions: make(map[uuid.UUID][]*Client),
	}
}

func (h *Hub) run() {
	ticker := time.NewTicker(writePeriod)
	for {
		select {
		// handle a request to register a new client
		case clientToCreate := <-h.register:
			log.Printf("creating new client...")
			h.clients[clientToCreate] = true

		// handle a request to unregister a client
		case clientToDelete := <-h.unregister:
			if _, ok := h.clients[clientToDelete]; ok {
				log.Printf("removing client...")
				// remove the client from the client list
				delete(h.clients, clientToDelete)
				// close the client's outbound message channel
				close(clientToDelete.send)
			}

		// read from clients
		case packet := <-h.receive:
			log.Printf("receiving packet from client %v...", packet.clientID)
			err := readFromClient(packet)
			if err != nil {
				log.Printf(err.Error())
			}

		// write to clients
		case <-ticker.C:
			log.Printf("sending data to %v clients...", len(h.clients))
			err := writeToClients(h)
			if err != nil {
				log.Printf(err.Error())
			}
		}
	}
}

func readFromClient(packet Packet) error {
	// hub performs a function on the data from the client
	// client will query a certain subset of data
	// keep state of the client's request
	log.Print(packet.message)
	return nil
}

func writeToClients(h *Hub) error {
	// generate fake data
	fakeData := fmt.Sprint(rand.Intn(100), rand.Intn(100))
	for client := range h.clients {
		client.send <- []byte(fakeData)
	}
	return nil
}
