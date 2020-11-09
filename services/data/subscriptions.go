package data

import (
	"context"
	"fmt"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/infra/engine/driver"
	pb "gitlab.com/beneath-hq/beneath/infra/engine/proto"
)

// SubscriptionMessage encapsulates a subscription update dispatched by a Broker
type SubscriptionMessage struct {
	Iterator   driver.RecordsIterator
	PrevCursor []byte
}

// subscriptions manages a map of instances to subscription callbacks
type subscriptions struct {
	sync.RWMutex
	i    int
	m    map[uuid.UUID]*callbacks
	open bool
}

// callbacks manages a set of subscription callbacks
type callbacks struct {
	sync.RWMutex
	m map[int]func(SubscriptionMessage)
}

// ServeSubscriptions brokers subscriptions until ctx is cancelled or an error occurs
func (s *Service) ServeSubscriptions(ctx context.Context) error {
	s.subscriptions = subscriptions{
		m:    make(map[uuid.UUID]*callbacks),
		open: true,
	}

	err := s.ReadWriteReports(ctx, s.handleWriteReport)
	if err != nil {
		return err
	}

	// TODO: gracefully shutdown subscriptions

	s.subscriptions.open = false
	s.Logger.Infow("subscriptions broker shutdown gracefully")
	return nil
}

// Subscribe opens a new subscription on an instance
func (s *Service) Subscribe(instanceID uuid.UUID, cursor []byte, cb func(SubscriptionMessage)) (context.CancelFunc, error) {
	if !s.subscriptions.open {
		return nil, fmt.Errorf("Cannot handle Subscribe because Service isn't serving subscriptions")
	}

	s.subscriptions.Lock()
	cbs := s.subscriptions.m[instanceID]
	if cbs == nil {
		cbs = &callbacks{m: make(map[int]func(SubscriptionMessage))}
		s.subscriptions.m[instanceID] = cbs
	}
	s.subscriptions.i++
	cbID := s.subscriptions.i
	s.subscriptions.Unlock()

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

// handleWriteReport processes a write report
func (s *Service) handleWriteReport(ctx context.Context, rep *pb.WriteReport) error {
	startTime := time.Now()

	instanceID := uuid.FromBytesOrNil(rep.InstanceId)

	s.subscriptions.RLock()
	cbs := s.subscriptions.m[instanceID]
	s.subscriptions.RUnlock()

	if cbs == nil {
		return nil
	}

	cbs.RLock()
	n := len(cbs.m)
	if n == 0 {
		cbs.RUnlock()
		return nil
	}

	msg := SubscriptionMessage{}
	for _, cb := range cbs.m {
		cb(msg)
	}

	cbs.RUnlock()

	s.Logger.Infow(
		"subscription dispatch",
		"instance", instanceID.String(),
		"clients", n,
		"elapsed", time.Since(startTime),
	)

	return nil
}
