package pubsub

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MaxMessageSize implements MessageQueue
func (p PubSub) MaxMessageSize() int {
	// The total request size is capped at 10MB, we set it slightly lower to not run into problems
	return 9999360 // floor1024(10MB)
}

// RegisterTopic implements MessageQueue
func (p PubSub) RegisterTopic(name string) error {
	qualified := fmt.Sprintf("%s-%s", p.Opts.TopicPrefix, name)
	topic, err := p.Client.CreateTopic(context.Background(), qualified)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			return err
		}
		// err is AlreadyExists
		topic = p.Client.Topic(qualified)
	}
	p.Topics[name] = topic
	return nil
}

// Publish implements MessageQueue
func (p PubSub) Publish(ctx context.Context, topic string, msg []byte) error {
	// push
	result := p.Topics[topic].Publish(ctx, &pubsub.Message{
		Data: msg,
	})

	// blocks until ack'ed by pubsub
	_, err := result.Get(ctx)
	return err
}

// Subscribe implements MessageQueue
func (p PubSub) Subscribe(ctx context.Context, topic string, name string, persistent bool, fn func(ctx context.Context, msg []byte) error) error {
	// create name
	fullName := fmt.Sprintf("%s-%s", p.Opts.SubscriptionPrefix, name)

	// get subscription
	var sub *pubsub.Subscription
	if persistent {
		sub = p.getPersistentSubscription(ctx, p.Topics[topic], fullName)
	} else {
		sub = p.getEphemeralSubscription(ctx, p.Topics[topic], fullName, p.Opts.SubscriberID)
	}

	// receive messages forever (or until error occurs)
	cctx, cancel := context.WithCancel(ctx)
	return sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		// ephemeral subscriptions we ack immediately
		if !persistent {
			msg.Ack()
		}

		// trigger callback function
		err := fn(ctx, msg.Data)
		if err != nil {
			// TODO: we'll want to keep the pipeline going in the future when things are stable
			log.S.Errorf("couldn't process %s record: %s", topic, err.Error())
			msg.Nack()
			cancel()
			return
		}

		// persistent subscriptions are ack'ed after succesful processing
		if persistent {
			msg.Ack()
		}
	})

}

// getPersistantSubscription finds or creates a topic subscription
func (p PubSub) getPersistentSubscription(ctx context.Context, topic *pubsub.Topic, name string) *pubsub.Subscription {
	// create/get subscriber to topic
	subscription, err := p.Client.CreateSubscription(context.Background(), name, pubsub.SubscriptionConfig{
		Topic:       topic,
		AckDeadline: 20 * time.Second,
	})

	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			panic(fmt.Errorf("error creating subscription '%s': %v", name, err))
		} else {
			subscription = p.Client.Subscription(name)
		}
	}

	return subscription
}

// getEphemeralSubscription produces a subscription that does its best to not keep a backlog and to delete itself when
// you stop using it
func (p PubSub) getEphemeralSubscription(ctx context.Context, topic *pubsub.Topic, name string, id string) *pubsub.Subscription {
	// compose name
	subname := fmt.Sprintf("%s-%s", name, id)

	// create/get subscriber to topic
	subscription, err := p.Client.CreateSubscription(context.Background(), subname, pubsub.SubscriptionConfig{
		Topic:             topic,
		AckDeadline:       20 * time.Second,
		RetentionDuration: 10 * time.Minute, // minimum
		ExpirationPolicy:  24 * time.Hour,   // minimum
	})

	// fail only if error is not an already exists error
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			panic(fmt.Errorf("error creating subscription '%s': %v", subname, err))
		} else {
			subscription = p.Client.Subscription(subname)
		}
	}

	// in case the subscription already exists, skip accummulated data on it
	err = subscription.SeekToTime(context.Background(), time.Now())
	if err != nil {
		status, ok := status.FromError(err)
		if ok && status.Code() == codes.Unimplemented && p.Opts.EmulatorHost != "" {
			// Seek not implemented on Emulator, ignore
		} else {
			panic(fmt.Errorf("error seeking on subscription '%s': %v", subname, err))
		}
	}

	return subscription
}

// Reset implements beneath.Service
func (p PubSub) Reset(ctx context.Context) error {
	for name, topic := range p.Topics {
		it := topic.Subscriptions(ctx)
		for {
			sub, err := it.Next()
			if err == iterator.Done {
				break
			}
			err = sub.Delete(ctx)
			if err != nil {
				return err
			}
		}
		err := topic.Delete(ctx)
		if err != nil {
			return err
		}
		err = p.RegisterTopic(name)
		if err != nil {
			return err
		}
	}
	return nil
}
