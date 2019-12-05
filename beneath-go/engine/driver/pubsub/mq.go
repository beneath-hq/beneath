package pubsub

import "context"

// MaxMessageSize implements beneath.MessageQueue
func (p PubSub) MaxMessageSize() int {
	panic("todo")
}

// RegisterTopic implements beneath.MessageQueue
func (p PubSub) RegisterTopic(name string) error {
	panic("todo")
}

// Publish implements beneath.MessageQueue
func (p PubSub) Publish(topic string, msg []byte) error {
	panic("todo")
}

// Subscribe implements beneath.MessageQueue
func (p PubSub) Subscribe(topic string, name string, persistant bool, fn func(ctx context.Context, msg []byte) error) error {
	panic("todo")
}
