package pubsub

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/mitchellh/mapstructure"

	"gitlab.com/beneath-hq/beneath/infrastructure/mq"
)

func init() {
	mq.AddDriver("pubsub", newPubsub)
}

// PubSub implements beneath.MessageQueue
type PubSub struct {
	Client *pubsub.Client
	Topics map[string]*pubsub.Topic
	Opts   *Options
}

// Options for creating a Pubsub
type Options struct {
	ProjectID          string `mapstructure:"project_id"`
	SubscriberID       string `mapstructure:"subscriber_id"`
	TopicPrefix        string `mapstructure:"topic_prefix"`
	SubscriptionPrefix string `mapstructure:"subscription_prefix"`
	EmulatorHost       string `mapstructure:"emulator_host"`
}

func newPubsub(optsMap map[string]interface{}) (mq.MessageQueue, error) {
	// decode options
	var opts Options
	err := mapstructure.Decode(optsMap, &opts)
	if err != nil {
		return nil, fmt.Errorf("error decoding pubsub options: %s", err.Error())
	}

	// connect to emulator if set
	if opts.EmulatorHost != "" {
		os.Setenv("PUBSUB_PROJECT_ID", opts.ProjectID)
		os.Setenv("PUBSUB_EMULATOR_HOST", opts.EmulatorHost)
	}

	// prepare pubsub client
	client, err := pubsub.NewClient(context.Background(), opts.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to pubsub: %s", err)
	}

	return &PubSub{
		Client: client,
		Topics: make(map[string]*pubsub.Topic),
		Opts:   &opts,
	}, nil
}
