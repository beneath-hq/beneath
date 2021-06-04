package pubsub

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"github.com/beneath-hq/beneath/infra/mq"
)

func init() {
	mq.AddDriver("pubsub", newPubsub)
}

// PubSub implements beneath.MessageQueue
type PubSub struct {
	Logger       *zap.SugaredLogger
	Client       *pubsub.Client
	Topics       map[string]*pubsub.Topic
	Opts         *Options
	SubscriberID string
}

// Options for creating a Pubsub
type Options struct {
	ProjectID          string `mapstructure:"project_id"`
	TopicPrefix        string `mapstructure:"topic_prefix"`
	SubscriptionPrefix string `mapstructure:"subscription_prefix"`
	EmulatorHost       string `mapstructure:"emulator_host"`
}

func newPubsub(logger *zap.Logger, opts *mq.Options) (mq.MessageQueue, error) {
	// decode options
	var psOpts Options
	err := mapstructure.Decode(opts.DriverOptions, &psOpts)
	if err != nil {
		return nil, fmt.Errorf("error decoding pubsub options: %s", err.Error())
	}

	// connect to emulator if set
	if psOpts.EmulatorHost != "" {
		os.Setenv("PUBSUB_PROJECT_ID", psOpts.ProjectID)
		os.Setenv("PUBSUB_EMULATOR_HOST", psOpts.EmulatorHost)
	}

	// prepare pubsub client
	client, err := pubsub.NewClient(context.Background(), psOpts.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to pubsub: %s", err)
	}

	return &PubSub{
		Logger:       logger.Named("pubsub").Sugar(),
		Client:       client,
		Topics:       make(map[string]*pubsub.Topic),
		Opts:         &psOpts,
		SubscriberID: opts.SubscriberID,
	}, nil
}
