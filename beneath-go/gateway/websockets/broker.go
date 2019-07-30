package websockets

import (
	"fmt"
	"log"
	"net/http"

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
	// Reference to the engine that supplies new data
	Engine *engine.Engine

	// use to register a new client with the broker
	Register chan *Client

	// use to remove a client from the broker
	Unregister chan *Client

	// use to submit a new request from the client for processing
	Requests chan Request

	// use to distribute records to subscribed clients
	Dispatch chan *pb.StreamMetricsPacket

	// used to manage websocket
	upgrader *websocket.Upgrader

	// Registered clients.
	clients map[*Client]bool

	// Client subscriptions. Mapping: stream instance ID --> list of subscribed clients
	// TODO: modify this to accept instanceID+key as input
	clientSubscriptions map[uuid.UUID][]*Client
}

// NewBroker initializes a new Broker
func NewBroker(engine *engine.Engine) *Broker {
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
		Engine:              engine,
		upgrader:            upgrader,
		clients:             make(map[*Client]bool),
		Register:            make(chan *Client),
		Unregister:          make(chan *Client),
		Requests:            make(chan Request),
		Dispatch:            make(chan *pb.StreamMetricsPacket),
		clientSubscriptions: make(map[uuid.UUID][]*Client),
	}

	// initialize run loop
	go broker.runForever()

	// initialize read loop
	// TODO: Call subscribe to keys function on broker.Engine
	go func() {
		err := engine.Streams.ReadMetricsMessage(func(msg *pb.StreamMetricsPacket) error {
			broker.Dispatch <- msg
			return nil
		})
		if err != nil {
			log.Panicf("ReadMetricsMessage crashed: %v", err.Error())
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

	// create client and register it with the hub
	log.Printf("registering new client...")
	b.Register <- newClient(b, ws)

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
		delete(b.clients, c)
	}
}

// process a client request
func (b *Broker) processRequest(r Request) {
	log.Printf("Read new request: %v", r.Message)
	// TODO: Update broker state (dispatch tables) based on client's request
}

// process messages from pubsub (route to relevant clients)
func (b *Broker) processMessage(metrics *pb.StreamMetricsPacket) {
	instanceID := uuid.FromBytesOrNil(metrics.InstanceId)

	// lookup stream
	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		log.Printf("cached stream is null for instanceid %s", instanceID.String())
	}

	/* TODO: Dispatch message to only the subscribed clients
	// check for clients subscribed to the instance_id
	subscribedClients := b.clientSubscriptions[instanceID]

	// if there are subscribers to the metrics packet, pull the data
	if len(subscribedClients > 0) {
		// get values for each key
	}
	else return
	*/

	// initialize records object
	records := make([]map[string]interface{}, len(metrics.Keys))

	// make new ReadRecords function
	err := b.Engine.Tables.ReadRecords(instanceID, metrics.Keys, func(idx uint, avroData []byte, sequenceNumber int64) error {
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

	// handle ReadRecords error
	if err != nil {
		log.Print("Could not read records from BigTable: ", err.Error())
	}

	// TODO: for i, client := range subscribedClients {
	for client := range b.clients {
		// loop through records and send one at a time
		for _, record := range records {
			client.sendData(fmt.Sprint(client), record)
		}
	}
}
