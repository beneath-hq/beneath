package mq

import (
	"context"
	"fmt"
	"math/rand"

	"go.uber.org/zap"
)

// MessageQueue encapsulates functionality for message passing in Beneath
type MessageQueue interface {
	// MaxMessageSize should return the maximum allowed byte size of published messages
	MaxMessageSize() int

	// RegisterTopic should register a new topic for message passing
	RegisterTopic(name string) error

	// Publish should issue a new message to all the topic's subscribers
	Publish(ctx context.Context, topic string, msg []byte) error

	// Subscribe should create a subscription for new messages on the topic.
	// If persistent, messages missed when offline should accumulate and be delivered on reconnect.
	Subscribe(ctx context.Context, topic string, name string, persistent bool, fn func(ctx context.Context, msg []byte) error) error

	// Reset should clear all data in the service (useful during testing)
	Reset(ctx context.Context) error
}

// Driver is a function that creates a MessageQueue from a config object
type Driver func(logger *zap.Logger, opts *Options) (MessageQueue, error)

// drivers is a registry of MessageQueue drivers
var drivers = make(map[string]Driver)

// AddDriver registers a new driver (by passing the driver's constructor)
func AddDriver(name string, driver Driver) {
	if drivers[name] != nil {
		panic(fmt.Errorf("MessageQueue driver already registered with name '%s'", name))
	}
	drivers[name] = driver
}

// Options for opening a message queue connection
type Options struct {
	DriverName    string                 `mapstructure:"driver"`
	SubscriberID  string                 `mapstructure:"subscriber_id"`
	DriverOptions map[string]interface{} `mapstructure:",remain"`
}

// NewMessageQueue constructs a new MessageQueue
func NewMessageQueue(logger *zap.Logger, opts *Options) (MessageQueue, error) {
	if opts.SubscriberID == "" {
		opts.SubscriberID = fmt.Sprintf("%x", rand.Uint64())
	}
	constructor := drivers[opts.DriverName]
	if constructor == nil {
		return nil, fmt.Errorf("'%s' is not a valid MessageQueue driver", opts.DriverName)
	}
	return constructor(logger.Named("mq"), opts)
}
