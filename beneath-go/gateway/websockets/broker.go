package websockets

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
		// remove client from all instance subscriptions
		for instanceID := range c.InstancesToIDs {
			index := getIndex(b.instanceSubscriptions[instanceID], c)
			if index != -1 {
				b.instanceSubscriptions[instanceID] = remove(b.instanceSubscriptions[instanceID], index)
			}
		}

		// clear the client's mappings
		c.IDsToInstances = make(map[string]uuid.UUID)
		c.InstancesToIDs = make(map[uuid.UUID]string)

		// delete client from the broker's client list
		delete(b.clients, c)
	}
}

// process a client request
func (b *Broker) processRequest(r Request) {
	// use the request's message type to process it accordingly
	switch r.Message.Type {
	case connectionInitMsgType:
		// TODO
	case connectionTerminateMsgType:
		// TODO
	case startMsgType:
		// get instanceID that the user would like to subscribe to
		instanceID := uuid.FromStringOrNil(r.Message.Payload["instance_id"].(string))

		// get instanceID info
		stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
		if stream == nil {
			r.Client.SendError("", "Error! That instance_id doesn't exist.")
			return
		}

		// check if client has already used the same ID
		if _, ok := r.Client.IDsToInstances[r.Message.ID]; ok {
			r.Client.SendError("", "Error! You are already using that ID.")
			return
		}

		// check auth
		if !r.Client.Key.ReadsProject(stream.ProjectID) {
			r.Client.SendError("", "Error! You don't have permission to read that stream.")
			return
		}

		// map the instanceID and the  1) instanceID and 2) the user's  to client subscription mappings
		r.Client.IDsToInstances[r.Message.ID] = instanceID
		r.Client.InstancesToIDs[instanceID] = r.Message.ID

		// add client to instance subscriptions
		if b.instanceSubscriptions[instanceID] == nil {
			b.instanceSubscriptions[instanceID] = []*Client{r.Client}
		} else {
			b.instanceSubscriptions[instanceID] = append(b.instanceSubscriptions[instanceID], r.Client)
		}
		return
	case stopMsgType:
		instanceID := r.Client.IDsToInstances[r.Message.ID]
		index := getIndex(b.instanceSubscriptions[instanceID], r.Client)

		// check if client is subscribed to the stream
		if index == -1 {
			r.Client.SendError("", "Error! The ID you supplied is not subscribed to any stream.")
			return
		}

		// remove all subscriptions
		b.instanceSubscriptions[instanceID] = remove(b.instanceSubscriptions[instanceID], index)
		delete(r.Client.IDsToInstances, r.Message.ID)
		delete(r.Client.InstancesToIDs, instanceID)
		return

	default:
		log.Panic("unrecognized message type")
	}
}

// process messages from pubsub (route to relevant clients)
func (b *Broker) processMessage(rep *pb.WriteRecordsReport) {
	// metrics to track
	startTime := time.Now()

	// lookup stream
	instanceID := uuid.FromBytesOrNil(rep.InstanceId)
	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		log.Panicf("cached stream is null for instanceid %s", instanceID.String())
	}

	// check for clients subscribed to the instanceID
	subscribedClients := b.instanceSubscriptions[instanceID]

	// only proceed if there are subscribers
	if len(subscribedClients) > 0 {
		// initialize the records object that will be sent to clients
		records := make([]map[string]interface{}, len(rep.Keys))

		// read records from the Tables driver, decode the values, and assign the keys to values
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

		// loop through all the subscribed clients
		for _, client := range subscribedClients {
			id := client.InstancesToIDs[instanceID]
			// loop through records and send one at a time
			for _, record := range records {
				client.SendData(id, record)
			}
		}

		// finalize metrics
		elapsed := time.Since(startTime)

		// log: number of clients sent to, number of rows sent, time elapsed, instance id
		log.Printf("Message sent to client(s). Number of clients: %d, Number of rows: %d, Time elapsed: %s, InstanceID: %s",
			len(subscribedClients), len(records), elapsed, instanceID.String())
	}
	return
}

// get index of element in Client list
func getIndex(slice []*Client, client *Client) int {
	for i, v := range slice {
		if v == client {
			return i
		}
	}
	return -1
}

// remove client from Client list
func remove(s []*Client, i int) []*Client {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
