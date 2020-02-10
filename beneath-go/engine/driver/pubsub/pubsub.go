package pubsub

import (
	"context"
	"os"
	"sync"

	"cloud.google.com/go/pubsub"

	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/engine/driver"
)

// configSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type configSpecification struct {
	ProjectID          string `envconfig:"PROJECT_ID" required:"true"`
	SubscriberID       string `envconfig:"SUBSCRIBER_ID" required:"true"`
	EmulatorHost       string `envconfig:"EMULATOR_HOST" required:"false"`
	TopicPrefix        string `envconfig:"TOPIC_PREFIX" required:"false"`
	SubscriptionPrefix string `envconfig:"SUBSCRIPTION_PREFIX" required:"false"`
}

// PubSub implements beneath.MessageQueue
type PubSub struct {
	config *configSpecification
	Client *pubsub.Client
	Topics map[string]*pubsub.Topic
}

// Global
var global PubSub
var once sync.Once

func createGlobal() {
	// parse config from env
	var config configSpecification
	core.LoadConfig("beneath_engine_pubsub", &config)

	// if EMULATOR_HOST set, configure pubsub for the emulator
	if config.EmulatorHost != "" {
		os.Setenv("PUBSUB_PROJECT_ID", config.ProjectID)
		os.Setenv("PUBSUB_EMULATOR_HOST", config.EmulatorHost)
	}

	// prepare pubsub client
	client, err := pubsub.NewClient(context.Background(), config.ProjectID)
	if err != nil {
		panic(err)
	}

	// create instance
	global = PubSub{
		config: &config,
		Client: client,
		Topics: make(map[string]*pubsub.Topic),
	}
}

// GetMessageQueue returns a Google PubSub implementation of beneath.MessageQueue
func GetMessageQueue() driver.MessageQueue {
	once.Do(createGlobal)
	return global
}
