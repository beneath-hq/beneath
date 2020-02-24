package subscriptions

import (
	"context"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/engine"
	"github.com/beneath-core/engine/driver"
	pb "github.com/beneath-core/engine/proto"
	"github.com/beneath-core/pkg/log"
)

// Message encapsulates a subscription update dispatched by a Broker
type Message struct {
	Iterator   driver.RecordsIterator
	PrevCursor []byte
}

// Broker manages cursor update subscriptions
type Broker struct {
	Engine        *engine.Engine
	subscriptions subscriptions
}

// subscriptions manages a map of instances to subscription callbacks
type subscriptions struct {
	sync.RWMutex
	i int
	m map[uuid.UUID]*callbacks
}

// callbacks manages a set of subscription callbacks
type callbacks struct {
	sync.RWMutex
	m map[int]func(Message)
}

// NewBroker initalizes a new broker
func NewBroker(eng *engine.Engine) *Broker {
	b := &Broker{
		Engine: eng,
		subscriptions: subscriptions{
			m: make(map[uuid.UUID]*callbacks),
		},
	}
	go b.runForever()
	return b
}

// Subscribe opens a new subscription on an instance
func (b *Broker) Subscribe(instanceID uuid.UUID, cursor []byte, cb func(Message)) (context.CancelFunc, error) {
	b.subscriptions.Lock()
	cbs := b.subscriptions.m[instanceID]
	if cbs == nil {
		cbs = &callbacks{m: make(map[int]func(Message))}
		b.subscriptions.m[instanceID] = cbs
	}
	b.subscriptions.i++
	cbID := b.subscriptions.i
	b.subscriptions.Unlock()

	cbs.Lock()
	cbs.m[cbID] = cb
	cbs.Unlock()

	cancel := func() {
		cbs.Lock()
		delete(cbs.m, cbID)
		cbs.Unlock()
	}
	return cancel, nil
}

// runForever subscribes to new write reports
func (b *Broker) runForever() {
	err := b.Engine.ReadWriteReports(b.handleWriteReport)
	if err != nil {
		panic(err)
	}
}

// handleWriteReport processes a write report
func (b *Broker) handleWriteReport(ctx context.Context, rep *pb.WriteReport) error {
	startTime := time.Now()

	instanceID := uuid.FromBytesOrNil(rep.InstanceId)

	b.subscriptions.RLock()
	cbs := b.subscriptions.m[instanceID]
	b.subscriptions.RUnlock()

	if cbs == nil {
		return nil
	}

	cbs.RLock()
	n := len(cbs.m)
	if n == 0 {
		cbs.RUnlock()
		return nil
	}

	msg := Message{}
	for _, cb := range cbs.m {
		cb(msg)
	}

	cbs.RUnlock()

	log.S.Infow(
		"subscription dispatch",
		"instance", instanceID.String(),
		"clients", n,
		"elapsed", time.Since(startTime),
	)

	return nil
}
