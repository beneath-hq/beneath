package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"

	pb "gitlab.com/beneath-hq/beneath/infrastructure/engine/proto"
	"gitlab.com/beneath-hq/beneath/infrastructure/mq"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// Msg represents an event handled by Bus
type Msg interface{}

// HandlerFunc represents an event handler.
// It's signature must be (func(context.Context, Msg) error).
type HandlerFunc interface{}

// Bus dispatches events to listener. Published events are processed
// immediately by sync listeners (and will use any DB transactions in ctx),
// and in a background worker by async listeners (passed over MQ).
type Bus struct {
	MQ mq.MessageQueue

	msgTypes       map[string]reflect.Type
	syncListeners  map[string][]HandlerFunc
	asyncListeners map[string][]HandlerFunc
}

const (
	asyncTopic        = "bus"
	asyncSubscription = "bus-worker"
)

// asyncMsg wraps a pb.BusMsg
type asyncMsg struct {
	ID        uuid.UUID
	Name      string
	Timestamp time.Time
	Payload   Msg
}

// NewBus creates a new Bus
func NewBus(mq mq.MessageQueue) (*Bus, error) {
	err := mq.RegisterTopic(asyncTopic)
	if err != nil {
		return nil, err
	}

	err = mq.RegisterTopic(controlEventsTopic)
	if err != nil {
		return nil, err
	}

	return &Bus{
		MQ:             mq,
		msgTypes:       make(map[string]reflect.Type),
		syncListeners:  make(map[string][]HandlerFunc),
		asyncListeners: make(map[string][]HandlerFunc),
	}, nil
}

// AddSyncListener registers a synchronous msg handler.
// Calls to Publish do not return until every sync listener
// has processed the msg. If any listener returns an error,
// subsequent listeners are *not* called, and the error is
// returned to the publisher.
func (b *Bus) AddSyncListener(handler HandlerFunc) {
	b.addListener(b.syncListeners, handler)
}

// AddAsyncListener registers an async msg handler, which
// handles published messages (passed through MQ) in a background
// worker.
func (b *Bus) AddAsyncListener(handler HandlerFunc) {
	b.addListener(b.asyncListeners, handler)
}

func (b *Bus) addListener(listeners map[string][]HandlerFunc, handler HandlerFunc) {
	// check handler takes two args and returns one
	handlerType := reflect.TypeOf(handler)
	if handlerType.NumIn() != 2 {
		panic(fmt.Errorf("expected handler <%v> to take two args, a context and a msg", handler))
	}
	if handlerType.NumOut() != 1 {
		panic(fmt.Errorf("expected handler <%v> to return one arg, an error", handler))
	}

	// check handler returns an error
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !handlerType.Out(0).Implements(errorType) {
		panic(fmt.Errorf("expected handler <%v> to return one arg, an error", handler))
	}

	// check first arg is a context.Context
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !handlerType.In(0).Implements(ctxType) {
		panic(fmt.Errorf("expected handler <%v> to take two args, a context and a msg", handler))
	}

	// check second arg is a struct pointer
	msgType := handlerType.In(1)
	if msgType.Kind() != reflect.Ptr || msgType.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("expected handler <%v> to take two args, a context and a msg", handler))
	}

	// get msg type and name
	msgName := msgType.Name()

	// track msg type if it doesn't exist
	_, exists := b.msgTypes[msgName]
	if !exists {
		b.msgTypes[msgName] = msgType
	}

	// add handler to listeners
	_, exists = listeners[msgName]
	if !exists {
		listeners[msgName] = make([]HandlerFunc, 0)
	}
	listeners[msgName] = append(listeners[msgName], handler)
}

// Publish publishes a message to sync and async listeners.
// If a sync listener fails, its error is returned and async
// listeners will not be called. Async listeners run in a background
// worker, so msg will be serialized with msgpack.
func (b *Bus) Publish(ctx context.Context, msg Msg) error {
	msgName := reflect.TypeOf(msg).Elem().Name()

	err := b.callListeners(ctx, b.syncListeners[msgName], msg)
	if err != nil {
		return err
	}

	err = b.publishAsync(ctx, msgName, msg)
	if err != nil {
		return err
	}

	return nil
}

// PublishDelayed will schedule a message for delayed handling.
// It's not currently implemented.
func (b *Bus) PublishDelayed(ctx context.Context, d time.Duration, msg Msg) error {
	panic("not implemented")
}

// PublishScheduled will schedule a message for delayed handling at a specific time.
// It's not currently implemented.
func (b *Bus) PublishScheduled(ctx context.Context, t time.Time, msg Msg) error {
	panic("not implemented")
}

func (b *Bus) callListeners(ctx context.Context, listeners []HandlerFunc, msg Msg) error {
	if len(listeners) == 0 {
		return nil
	}

	params := make([]reflect.Value, 2)
	params[0] = reflect.ValueOf(ctx)
	params[1] = reflect.ValueOf(msg)

	for _, handler := range listeners {
		ret := reflect.ValueOf(handler).Call(params)
		err := ret[0].Interface()
		if err != nil {
			return err.(error)
		}
	}

	return nil
}

func (b *Bus) publishAsync(ctx context.Context, msgName string, msg Msg) error {
	listeners := b.asyncListeners[msgName]
	if len(listeners) == 0 {
		return nil
	}

	amsg := &asyncMsg{
		ID:        uuid.NewV4(),
		Name:      msgName,
		Timestamp: time.Now(),
		Payload:   msg,
	}

	data, err := b.marshalAsyncMsg(ctx, amsg)
	if err != nil {
		return err
	}

	err = b.MQ.Publish(ctx, asyncTopic, data)
	if err != nil {
		return err
	}

	log.S.Infow("bus_async_publish", "id", amsg.ID.String(), "name", msgName, "bytes", len(data))
	return nil
}

func (b *Bus) marshalAsyncMsg(ctx context.Context, amsg *asyncMsg) ([]byte, error) {
	payload, err := msgpack.Marshal(amsg.Payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling msg %s with msgpack: %s", amsg.Name, err.Error())
	}

	pbmsg := &pb.BusMsg{
		Id:        amsg.ID.Bytes(),
		Name:      amsg.Name,
		Timestamp: timeutil.UnixMilli(amsg.Timestamp),
		Payload:   payload,
	}

	return proto.Marshal(pbmsg)
}

func (b *Bus) unmarshalAsyncMsg(ctx context.Context, data []byte) (*asyncMsg, error) {
	pbmsg := &pb.BusMsg{}
	err := proto.Unmarshal(data, pbmsg)
	if err != nil {
		return nil, err
	}

	msgType, ok := b.msgTypes[pbmsg.Name]
	if !ok {
		return nil, fmt.Errorf("cannot find msg type with name '%s'", pbmsg.Name)
	}

	payload := reflect.New(msgType).Interface().(Msg) // pointer to struct
	err = msgpack.Unmarshal(pbmsg.Payload, payload)
	if err != nil {
		return nil, err
	}

	amsg := &asyncMsg{
		ID:        uuid.FromBytesOrNil(pbmsg.Id),
		Name:      pbmsg.Name,
		Timestamp: timeutil.FromUnixMilli(pbmsg.Timestamp),
		Payload:   payload,
	}

	return amsg, nil
}

// Run subscribes to MQ for published async messages and calls the relevant
// async listeners.
// It's suitable for running in a worker process, but care should be taken to
// make sure all the correct listeners have been instantiated in the process.
func (b *Bus) Run(ctx context.Context) error {
	return b.MQ.Subscribe(ctx, asyncTopic, asyncSubscription, true, func(ctx context.Context, data []byte) error {
		startTime := time.Now()

		amsg, err := b.unmarshalAsyncMsg(ctx, data)
		if err != nil {
			log.S.Errorf("bus async worker got unmarshal err, nacking...: %s", err)
			return err
		}

		listeners := b.asyncListeners[amsg.Name]
		if len(listeners) == 0 {
			log.S.Errorf("bus async worker got msg with no listeners, the correct services are likely not created in this process, acking...: %x", data)
			return nil
		}

		err = b.callListeners(ctx, listeners, amsg.Payload)
		if err != nil {
			log.S.Errorf("bus async worker got err when calling listeners, nacking...: %s", err)
			return err
		}

		log.S.Infow("bus_async_processed", "id", amsg.ID.String(), "name", amsg.Name, "time", amsg.Timestamp, "elapsed", time.Since(startTime))

		return nil
	})
}

// TODO: Remove this, but not right now

const (
	controlEventsTopic        = "control-events"
	controlEventsSubscription = "control-events-worker"
)

// ControlEvent describes a control-plane event
type ControlEvent struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// PublishControlEvent publishes a control-plane event to controlEventsTopic
// for external consumption (e.g. for BI purposes)
func (b *Bus) PublishControlEvent(ctx context.Context, name string, data map[string]interface{}) error {
	msg := ControlEvent{
		ID:        uuid.NewV4(),
		Name:      name,
		Timestamp: time.Now(),
		Data:      data,
	}
	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if len(json) > b.MQ.MaxMessageSize() {
		return fmt.Errorf("control event message %v has invalid size", json)
	}
	return b.MQ.Publish(ctx, controlEventsTopic, json)
}

// SubscribeControlEvents subscribes to all control events
func (b *Bus) SubscribeControlEvents(ctx context.Context, fn func(context.Context, *ControlEvent) error) error {
	return b.MQ.Subscribe(ctx, controlEventsTopic, controlEventsSubscription, true, func(ctx context.Context, msg []byte) error {
		t := &ControlEvent{}
		err := json.Unmarshal(msg, t)
		if err != nil {
			return err
		}
		return fn(ctx, t)
	})
}
