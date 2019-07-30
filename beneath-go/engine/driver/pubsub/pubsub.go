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
	MetricsTopic              string `envconfig:"METRICS_TOPIC" required:"true"`
	// TODO: Add subscription names; every gateway (probably 5 of them) need to have their own subscription
	MetricsSubscription0 string `envconfig:"METRICS_SUBSCRIPTION_0" required:"true"`
	MetricsSubscription1 string `envconfig:"METRICS_SUBSCRIPTION_1" required:"true"`
	MetricsSubscription2 string `envconfig:"METRICS_SUBSCRIPTION_2" required:"true"` // is this the right way to do it?
}

// Pubsub implements beneath.StreamsDriver
type Pubsub struct {
	config             *configSpecification
	Client             *pubsub.Client
	WriteRequestsTopic *pubsub.Topic
	MetricsTopic       *pubsub.Topic
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

	// create metrics topic
	metricsTopic, err := client.CreateTopic(context.Background(), config.MetricsTopic)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Panicf("error creating topic: %v", err)
		} else {
			metricsTopic = client.Topic(config.MetricsTopic)
		}
	}

	// create instance
	return &Pubsub{
		config:             &config,
		Client:             client,
		WriteRequestsTopic: writeRequestsTopic,
		MetricsTopic:       metricsTopic,
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
	sub := p.getWriteRequestsSubscription()
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

// get WriteRequests subscription
func (p *Pubsub) getWriteRequestsSubscription() *pubsub.Subscription {
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

// QueueMetricsMessage implements beneath.StreamsDriver
func (p *Pubsub) QueueMetricsMessage(metrics *pb.StreamMetricsPacket) error {
	// encode message
	msg, err := proto.Marshal(metrics)
	if err != nil {
		log.Panicf("error marshalling Metrics: %v", err)
	}

	// check encoded message size
	if len(msg) > p.GetMaxMessageSize() {
		return fmt.Errorf(
			"metrics have invalid size <%d> (max message size is <%d bytes>)",
			len(msg), p.GetMaxMessageSize(),
		)
	}

	// push
	ctx := context.Background()
	result := p.MetricsTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	// blocks until ack'ed by pubsub
	_, err = result.Get(ctx)
	return err
}

// ReadMetricsMessage implements beneath.StreamsDriver
func (p *Pubsub) ReadMetricsMessage(fn func(*pb.StreamMetricsPacket) error) error {
	// prepare subscription and context
	subscriberID := "0" // TODO: make this dynamic
	sub := p.getMetricsSubscription(subscriberID)
	cctx, cancel := context.WithCancel(context.Background())

	// receive pubsub messages forever (or until error occurs)
	err := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		// called for every single message on pubsub
		// these messages will have been written by QueueMetricsMessage

		// unmarshal metrics from pubsub
		metrics := &pb.StreamMetricsPacket{}
		err := proto.Unmarshal(msg.Data, metrics)
		if err != nil {
			// standard error handling for pubsub library
			log.Printf("couldn't unmarshal metrics: %s", err.Error())
			cancel()
			return
		}

		// trigger callback function
		err = fn(metrics)
		if err != nil {
			// TODO: we'll want to keep the pipeline going in the future when things are stable
			// log error and cancel
			log.Printf("couldn't process metrics: %s", err.Error())
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

// get Metrics subscription for the relevant subscriber
func (p *Pubsub) getMetricsSubscription(subscriberID string) *pubsub.Subscription {
	// get relevant subscriber config
	subName := ""
	switch subscriberID {
	case "0":
		subName = p.config.MetricsSubscription0
	case "1":
		subName = p.config.MetricsSubscription1
	case "2":
		subName = p.config.MetricsSubscription2
	default:
		log.Fatalf("Could not create subscriber to Metrics topic: unrecognized subscriberID")
	}

	// create/get subscriber to topic
	subscription, err := p.Client.CreateSubscription(context.Background(), subName, pubsub.SubscriptionConfig{
		Topic:       p.MetricsTopic,
		AckDeadline: 20 * time.Second,
	})

	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Panicf("error creating subscription: %v", err)
		} else {
			subscription = p.Client.Subscription(subName)
		}
	}

	return subscription
}
