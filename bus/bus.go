package bus

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	pb "github.com/beneath-hq/beneath/infra/engine/proto"
	"github.com/beneath-hq/beneath/infra/mq"
	"github.com/beneath-hq/beneath/pkg/timeutil"
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
	Logger *zap.SugaredLogger
	MQ     mq.MessageQueue

	msgTypes              map[string]reflect.Type
	syncListeners         map[string][]HandlerFunc
	asyncListeners        map[string][]HandlerFunc
	asyncOrderedListeners map[string][]HandlerFunc
}

const (
	asyncTopic               = "bus"
	asyncSubscription        = "bus-worker"
	asyncOrderedTopic        = "bus-ordered"
	asyncOrderedSubscription = "bus-ordered-worker"
)

// asyncMsg wraps a pb.BusMsg
type asyncMsg struct {
	ID        uuid.UUID
	Name      string
	Timestamp time.Time
	Payload   Msg
}

// NewBus creates a new Bus
func NewBus(logger *zap.Logger, mq mq.MessageQueue) (*Bus, error) {
	err := mq.RegisterTopic(asyncTopic, false)
	if err != nil {
		return nil, err
	}

	err = mq.RegisterTopic(asyncOrderedTopic, true)
	if err != nil {
		return nil, err
	}

	return &Bus{
		Logger:                logger.Named("bus").Sugar(),
		MQ:                    mq,
		msgTypes:              make(map[string]reflect.Type),
		syncListeners:         make(map[string][]HandlerFunc),
		asyncListeners:        make(map[string][]HandlerFunc),
		asyncOrderedListeners: make(map[string][]HandlerFunc),
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

// AddAsyncOrderedListener registers an async msg handler, which
// handles published messages (passed through MQ) in a background
// worker. This one processes messages sequentially.
func (b *Bus) AddAsyncOrderedListener(handler HandlerFunc) {
	b.addListener(b.asyncOrderedListeners, handler)
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
	msgName := msgType.Elem().Name()

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
	msgPtr := reflect.TypeOf(msg)
	if msgPtr.Kind() != reflect.Ptr || msgPtr.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("expected published msg to be a struct pointer, got %T", msg))
	}
	msgName := msgPtr.Elem().Name()

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
	orderedListeners := b.asyncOrderedListeners[msgName]

	if len(listeners) == 0 && len(orderedListeners) == 0 {
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

	group, ctx := errgroup.WithContext(ctx)

	if len(listeners) > 0 {
		group.Go(func() error {
			err = b.MQ.Publish(ctx, asyncTopic, data, nil)
			if err != nil {
				return err
			}
			b.Logger.Infow("async publish", "id", amsg.ID.String(), "name", msgName, "bytes", len(data))
			return nil
		})
	}

	if len(orderedListeners) > 0 {
		group.Go(func() error {
			err = b.MQ.Publish(ctx, asyncOrderedTopic, data, &msgName)
			if err != nil {
				return err
			}
			b.Logger.Infow("async ordered publish", "id", amsg.ID.String(), "name", msgName, "bytes", len(data))
			return nil
		})
	}

	return group.Wait()
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

	payload := reflect.New(msgType.Elem()).Interface().(Msg) // pointer to struct
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
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return b.runAsyncSubscriber(ctx)
	})

	group.Go(func() error {
		return b.runAsyncOrderedSubscriber(ctx)
	})

	return group.Wait()
}

func (b *Bus) runAsyncSubscriber(ctx context.Context) error {
	return b.MQ.Subscribe(ctx, asyncTopic, asyncSubscription, true, false, func(ctx context.Context, data []byte) error {
		startTime := time.Now()

		amsg, err := b.unmarshalAsyncMsg(ctx, data)
		if err != nil {
			b.Logger.Errorf("async worker got unmarshal err, nacking...: %s", err)
			return err
		}

		listeners := b.asyncListeners[amsg.Name]
		if len(listeners) == 0 {
			b.Logger.Errorf("async worker got msg with no listeners, the correct services are likely not created in this process, acking...: %x", data)
			return nil
		}

		err = b.callListeners(ctx, listeners, amsg.Payload)
		if err != nil {
			b.Logger.Errorf("async worker got err when calling listeners, nacking...: %s", err)
			return err
		}

		b.Logger.Infow("async processed", "id", amsg.ID.String(), "name", amsg.Name, "time", amsg.Timestamp, "elapsed", time.Since(startTime))

		return nil
	})
}

func (b *Bus) runAsyncOrderedSubscriber(ctx context.Context) error {
	return b.MQ.Subscribe(ctx, asyncOrderedTopic, asyncOrderedSubscription, true, true, func(ctx context.Context, data []byte) error {
		startTime := time.Now()

		amsg, err := b.unmarshalAsyncMsg(ctx, data)
		if err != nil {
			b.Logger.Errorf("async ordered worker got unmarshal err, nacking...: %s", err)
			return err
		}

		listeners := b.asyncOrderedListeners[amsg.Name]
		if len(listeners) == 0 {
			b.Logger.Errorf("async ordered worker got msg with no listeners, the correct services are likely not created in this process, acking...: %x", data)
			return nil
		}

		err = b.callListeners(ctx, listeners, amsg.Payload)
		if err != nil {
			b.Logger.Errorf("async ordered worker got err when calling listeners, nacking...: %s", err)
			return err
		}

		b.Logger.Infow("async ordered processed", "id", amsg.ID.String(), "name", amsg.Name, "time", amsg.Timestamp, "elapsed", time.Since(startTime))

		return nil
	})
}
