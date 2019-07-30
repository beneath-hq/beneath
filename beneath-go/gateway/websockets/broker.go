package websockets

import (
	"fmt"
	"log"
	"net/http"

	"github.com/beneath-core/beneath-go/control/auth"

	"github.com/beneath-core/beneath-go/control/model"
	"github.com/beneath-core/beneath-go/engine"
	pb "github.com/beneath-core/beneath-go/proto"

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

	// use to distribute records to subscribed clients
	Dispatch chan *pb.WriteRecordsReport

	// Reference to the engine that supplies new data
	engine *engine.Engine

	// used to manage websocket
	upgrader *websocket.Upgrader

	// Registered clients.
	clients map[*Client]bool

	// Client subscriptions. Mapping: stream instance ID --> list of subscribed clients
	instanceSubscriptions map[uuid.UUID][]*Client
}

// NewBroker initializes a new Broker
func NewBroker(engine *engine.Engine) *Broker {
	// create upgrader
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: true,
	}

	// create broker
	broker := &Broker{
		engine:                engine,
		upgrader:              upgrader,
		clients:               make(map[*Client]bool),
		Register:              make(chan *Client),
		Unregister:            make(chan *Client),
		Requests:              make(chan Request),
		Dispatch:              make(chan *pb.WriteRecordsReport),
		instanceSubscriptions: make(map[uuid.UUID][]*Client),
	}

	// initialize run loop
	go broker.runForever()

	// initialize reading data from engine (puts data on Dispatch, which is read in runForever)
	go func() {
		err := engine.Streams.ReadWriteReports(func(rep *pb.WriteRecordsReport) error {
			broker.Dispatch <- rep
			return nil
		})
		if err != nil {
			log.Panicf("ReadWriteReports crashed: %v", err.Error())
		}
	}()

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

	// get auth
	key := auth.GetKey(r.Context())

	// create client and register it with the hub
	b.Register <- NewClient(b, ws, key)

	return nil
}

func (b *Broker) runForever() {
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
		// TODO: remove all subscriptions
		delete(b.clients, c)
	}
}

// process a client request
func (b *Broker) processRequest(r Request) {
	/*
		// PSEUDOCODE

		// get id, payload (parse instanceID), type from r

		// type is either startMsgType or stopMsgType

		// IF startMsgType

		stream := ... from instanceID

		if r.Client.Subscriptions[id] != nil {
			// TODO IN FUTURE: What if IDs are different?
			return
		}

		// check auth
		if !r.Client.Key.ReadsProject(stream.ProjectID) {
			// ERROR
		}

		// add to client subscriptions
		r.Client.Subscriptions[id] = instanceID

		// add to instanceSubscriptions
		subscribed := b.instanceSubscriptions[instanceID]
		if b.instanceSubscriptions[instanceID] == nil {
			b.instanceSubscriptions[instanceID] = []*Client{r.Client}
		} else {
			b.instanceSubscriptions[instanceID] = append(b.instanceSubscriptions[instanceID], r.Client)
		}

		// ELSE IF stopMsgType

		Use Client.Subscriptions to find instance ID, then use that to remove from broker.instanceSubscriptions
	*/
	log.Printf("Read new request: %v", r.Message)
	// TODO: Update broker state (dispatch tables) based on client's request
	// TODO: Use client.Key to check if allowed to read instance requested
}

// process messages from pubsub (route to relevant clients)
func (b *Broker) processMessage(rep *pb.WriteRecordsReport) {
	// lookup stream
	instanceID := uuid.FromBytesOrNil(rep.InstanceId)
	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		log.Panicf("cached stream is null for instanceid %s", instanceID.String())
	}

	/* TODO: Dispatch message to only the subscribed clients
	// check for clients subscribed to the instance_id
	subscribedClients := b.instanceSubscriptions[instanceID]

	// if there are subscribers to the rep packet, pull the data
	if len(subscribedClients > 0) {
		// get values for each key
	}
	else return
	*/

	// initialize records object
	records := make([]map[string]interface{}, len(rep.Keys))

	// make new ReadRecords function
	err := b.engine.Tables.ReadRecords(instanceID, rep.Keys, func(idx uint, avroData []byte, sequenceNumber int64) error {
		// decode the avro data
		dataT, err := stream.AvroCodec.Unmarshal(avroData, false)
		if err != nil {
			return fmt.Errorf("unable to decode avro data")
		}

		// assert that the decoded data is a map
		data, ok := dataT.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected decoded data to be a map, got %T", dataT)
		}

		// assign key to value
		records[idx] = data

		return nil
	})
	if err != nil {
		log.Panicf("error reading Tables for instance ID '%s': %v", instanceID.String(), err.Error())
	}

	// TODO: for i, client := range subscribedClients {
	for client := range b.clients {
		// loop through records and send one at a time
		for _, record := range records {
			client.SendData("gotten from client", record)
		}
	}
}
