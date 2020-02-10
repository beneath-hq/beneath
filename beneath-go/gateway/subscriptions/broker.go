package subscriptions

import (
	"context"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/beneath-go/engine"
	"github.com/beneath-core/beneath-go/engine/driver"
)

// Broker manages cursor update subscriptions
type Broker struct {
	Engine *engine.Engine
}

// NewBroker initalizes a new broker
func NewBroker(eng *engine.Engine) *Broker {
	context.WithCancel(context.Background())
	return &Broker{
		Engine: eng,
	}
}

// Subscribe opens a new subscription on an instance
func (b *Broker) Subscribe(instanceID uuid.UUID, cursor []byte, cb func(Message)) (context.CancelFunc, error) {
	panic("not implemented")
}

// Message encapsulates a subscription update dispatched by a Broker
type Message struct {
	Iterator   driver.RecordsIterator
	PrevCursor []byte
}
