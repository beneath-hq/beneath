package websockets

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/beneath-core/beneath-go/core/timeutil"

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

// Dispatch bundles a filter and several records to sent to subscribers
type Dispatch struct {
	Filter  SubscriptionFilter
	Records []map[string]interface{}
}

// Broker maintains the set of active clients and routes messages to the clients
type Broker struct {
	// use to register a new client with the broker
	register chan *Client

	// use to remove a client from the broker
	unregister chan *Client

	// use to submit a new request from the client for processing
	requests chan Request

	// use to distribute records to subscribed clients
	dispatch chan Dispatch

	// Reference to the engine that supplies new data
	engine *engine.Engine

	// used to manage websocket
	upgrader *websocket.Upgrader

	// used to send keep alive messages regularly
	keepAliveTicker *time.Ticker

	// Registered clients
	clients map[*Client]bool

	// All open subscriptions (map of instanceIDs to subscribed clients)
	subscriptions map[SubscriptionFilter]map[*Client]bool

	// lock on subscriptions
	subscriptionsMu sync.RWMutex
}

const (
	connectionKeepAliveInterval = 10 * time.Second
)

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
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		requests:        make(chan Request),
		dispatch:        make(chan Dispatch),
		engine:          engine,
		upgrader:        upgrader,
		keepAliveTicker: time.NewTicker(connectionKeepAliveInterval),
		clients:         make(map[*Client]bool),
		subscriptions:   make(map[SubscriptionFilter]map[*Client]bool),
	}

	// initialize run loop
	go broker.runForever()

	// initialize reading data from engine (puts data on Dispatch, which is read in runForever)
	go func() {
		err := engine.Streams.ReadWriteReports(broker.handleWriteReport)
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
	ws, err := b.upgrader.Upgrade(w, r, http.Header{
		"Sec-Websocket-Protocol": []string{"graphql-ws"},
	})
	if err != nil {
		return err
	}

	// get auth
	secret := auth.GetSecret(r.Context())

	// create client and register it with the hub
	b.register <- NewClient(b, ws, secret)

	return nil
}

func (b *Broker) runForever() {
	for {
		select {
		case client := <-b.register:
			b.processRegister(client)
		case client := <-b.unregister:
			b.processUnregister(client)
		case req := <-b.requests:
			b.processRequest(req)
		case msg := <-b.dispatch:
			b.processMessage(msg)
		case <-b.keepAliveTicker.C:
			b.sendKeepAlive()
		}
	}
}

// CloseClient safely closes the client and can be called multiple times
func (b *Broker) CloseClient(c *Client) {
	b.unregister <- c
}

// broadcasts keep alive message to all subscribers
func (b *Broker) sendKeepAlive() {
	if len(b.clients) == 0 {
		return
	}
	startTime := time.Now()
	for client := range b.clients {
		client.SendKeepAlive()
	}
	elapsed := time.Since(startTime)
	log.Printf("Sent keep alive to %d client(s) in %s", len(b.clients), elapsed)
}

// Handles incoming write report from pubsub and puts them on the dispatch channel.
// It checks if there are any subscribers for this report before dispatching it.
// If there are subscribers, it gets the full records from bigtable first.
func (b *Broker) handleWriteReport(rep *pb.WriteRecordsReport) error {
	// metrics to track
	startTime := time.Now()

	// get instance and filter
	instanceID := uuid.FromBytesOrNil(rep.InstanceId)
	filter := SubscriptionFilter(instanceID)

	// check if worth continuing
	b.subscriptionsMu.RLock()
	n := len(b.subscriptions[filter])
	b.subscriptionsMu.RUnlock()
	if n == 0 {
		return nil
	}

	// get stream
	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		log.Panicf("cached stream is null for instanceid %s", instanceID.String())
	}

	// read and decode records matchin rep.Keys from Tables
	records := make([]map[string]interface{}, len(rep.Keys))
	err := b.engine.Tables.ReadRecords(instanceID, rep.Keys, func(idx uint, avroData []byte, timestamp time.Time) error {
		// decode the avro data
		obj, err := stream.Codec.UnmarshalAvro(avroData)
		if err != nil {
			return fmt.Errorf("unable to decode avro data")
		}

		// assert that the decoded data is a map
		data, err := stream.Codec.ConvertFromAvroNative(obj, true)
		if err != nil {
			return fmt.Errorf("unable to decode avro data")
		}

		// assign timestamp into data
		data["@meta"] = map[string]interface{}{
			"timestamp": timeutil.UnixMilli(timestamp),
		}

		// assign key to value
		records[idx] = data

		return nil
	})

	if err != nil {
		log.Panicf("error reading Tables for instance ID '%s': %v", instanceID.String(), err.Error())
	}

	// push to dispatch channel
	b.dispatch <- Dispatch{
		Filter:  filter,
		Records: records,
	}

	// log metrics
	elapsed := time.Since(startTime)
	log.Printf("Sent %d row(s) from instance %s to %d client(s) in %s", len(records), instanceID.String(), n, elapsed)

	return nil
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

	c.mu.Lock()
	b.subscriptionsMu.RLock() // only read because deletes happen on child map, which does not require concurrency
	for filter := range c.Subscriptions {
		delete(b.subscriptions[filter], c)
	}
	b.subscriptionsMu.RUnlock()
	c.mu.Unlock()

	c.Close()
}

// process a client request
func (b *Broker) processRequest(r Request) {
	// use the request's message type to process it accordingly
	switch r.Message.Type {
	case connectionInitMsgType:
		r.Client.SendConnectionAck()
	case connectionTerminateMsgType:
		r.Client.Terminate(websocket.CloseNormalClosure, "terminated")
	case startMsgType:
		b.processStartRequest(r)
	case stopMsgType:
		b.processStopRequest(r)
	default:
		r.Client.SendConnectionError("unrecognized message type")
		r.Client.Terminate(websocket.CloseProtocolError, "unrecognized message type")
	}
}

// process a startMsgType
func (b *Broker) processStartRequest(r Request) {
	// check format
	query, ok := r.Message.Payload["query"].(string)
	if !ok {
		r.Client.SendError(r.Message.ID, "Error! Query key required in payload")
		return
	}

	// get instanceID that the user would like to subscribe to
	// FUTURE: the query will be a proper graphql query
	instanceID := uuid.FromStringOrNil(query)

	// get instanceID info
	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		r.Client.SendError(r.Message.ID, "Error! That instance_id doesn't exist")
		return
	}

	// check auth
	if !r.Client.Secret.ReadsProject(stream.ProjectID) {
		r.Client.SendError(r.Message.ID, "Error! You don't have permission to read that stream.")
		return
	}

	// "convert" to filter (in future, may be more elaborate)
	filter := SubscriptionFilter(instanceID)

	// check if client has already used the same ID
	err := r.Client.TrackSubscription(filter, r.Message.ID)
	if err != nil {
		r.Client.SendError(r.Message.ID, err.Error())
		return
	}

	// add client to subscriptions
	b.subscriptionsMu.Lock()
	if b.subscriptions[filter] == nil {
		b.subscriptions[filter] = make(map[*Client]bool)
	}
	b.subscriptions[filter][r.Client] = true
	b.subscriptionsMu.Unlock()
}

// process a stopMsgType
func (b *Broker) processStopRequest(r Request) {
	// remove subscription from client
	filter, ok := r.Client.UntrackSubscription(r.Message.ID)
	if !ok {
		r.Client.SendError(r.Message.ID, "Error! The ID you supplied is not subscribed to any stream")
		return
	}

	// remove subscription from broker
	b.subscriptionsMu.RLock() // only read because editing child map
	delete(b.subscriptions[filter], r.Client)
	b.subscriptionsMu.RUnlock()

	// send complete (as per spec)
	r.Client.SendComplete(r.Message.ID)
}

// process messages from pubsub (route to relevant clients)
func (b *Broker) processMessage(d Dispatch) {
	// get subscribers
	b.subscriptionsMu.RLock()
	subscribers := b.subscriptions[d.Filter]
	b.subscriptionsMu.RUnlock()

	// push records to all subscribers
	for c := range subscribers {
		subID := c.GetSubscriptionID(d.Filter)
		for _, record := range d.Records {
			c.SendData(subID, record)
		}
	}
}
