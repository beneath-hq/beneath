package pubsub

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/beneath-core/beneath-go/core"
	pb "github.com/beneath-core/beneath-go/proto"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// configSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type configSpecification struct {
	ProjectID                 string `envconfig:"PROJECT_ID" required:"true"`
	InstanceID                string `envconfig:"INSTANCE_ID" required:"true"`
	EmulatorHost              string `envconfig:"EMULATOR_HOST" required:"false"`
	WriteRequestsTopic        string `envconfig:"WRITE_REQUESTS_TOPIC" required:"true"`
	WriteRequestsSubscription string `envconfig:"WRITE_REQUESTS_SUBSCRIPTION" required:"true"`
}

// Pubsub implements beneath.StreamsDriver
type Pubsub struct {
	config             *configSpecification
	Client             *pubsub.Client
	WriteRequestsTopic *pubsub.Topic
}

// New returns a new
func New() *Pubsub {
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
		log.Fatalf("could not create pubsub client: %v", err)
	}

	// create write requests topic
	writeRequestsTopic, err := client.CreateTopic(context.Background(), config.WriteRequestsTopic)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Panicf("error creating topic: %v", err)
		} else {
			writeRequestsTopic = client.Topic(config.WriteRequestsTopic)
		}
	}

	// create instance
	return &Pubsub{
		config:             &config,
		Client:             client,
		WriteRequestsTopic: writeRequestsTopic,
	}
}

// GetMaxMessageSize implements beneath.StreamsDriver
func (p *Pubsub) GetMaxMessageSize() int {
	return 10000000
}

// QueueWriteRequest implements beneath.StreamsDriver
func (p *Pubsub) QueueWriteRequest(req *pb.WriteRecordsRequest) error {
	// encode message
	msg, err := proto.Marshal(req)
	if err != nil {
		log.Panicf("error marshalling WriteRecordsRequest: %v", err)
	}

	// check encoded message size
	if len(msg) > p.GetMaxMessageSize() {
		return fmt.Errorf(
			"write request has invalid size <%d> (max message size is <%d bytes>)",
			len(msg), p.GetMaxMessageSize(),
		)
	}

	// push
	ctx := context.Background()
	result := p.WriteRequestsTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	// blocks until ack'ed by pubsub
	_, err = result.Get(ctx)
	return err
}

// ReadWriteRequests implements beneath.StreamsDriver
func (p *Pubsub) ReadWriteRequests(fn func(*pb.WriteRecordsRequest) error) error {
	// prepare subscription and context
	sub := p.getSubscription()
	cctx, cancel := context.WithCancel(context.Background())

	// receive pubsub messages forever (or until error occurs)
	err := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		// called for every single message on pubsub
		// these messages will have been written by QueueWriteRequest

		// unmarshal write request from pubsub
		req := &pb.WriteRecordsRequest{}
		err := proto.Unmarshal(msg.Data, req)
		if err != nil {
			// standard error handling for pubsub library
			log.Printf("couldn't unmarshal write records request: %s", err.Error())
			cancel()
			return
		}

		// trigger callback function
		err = fn(req)
		if err != nil {
			// TODO: we'll want to keep the pipeline going in the future when things are stable
			// log error and cancel
			log.Printf("couldn't process write request: %s", err.Error())
			cancel()
			return
		}

		// ack the message -- all processing finished successfully
		msg.Ack()
	})

	// pubsub stopped listening for new messages for some reason
	// return the error
	return err
}

// get subscription
func (p *Pubsub) getSubscription() *pubsub.Subscription {
	// create/get subscriber to topic
	subscription, err := p.Client.CreateSubscription(context.Background(), p.config.WriteRequestsSubscription, pubsub.SubscriptionConfig{
		Topic:       p.WriteRequestsTopic,
		AckDeadline: 20 * time.Second,
	})

	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Panicf("error creating subscription: %v", err)
		} else {
			subscription = p.Client.Subscription(p.config.WriteRequestsSubscription)
		}
	}

	return subscription
}
