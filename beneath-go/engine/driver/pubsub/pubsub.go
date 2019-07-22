package pubsub

import (
	"context"
	"fmt"
	"log"
	"os"

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
	ProjectID          string `envconfig:"PROJECT_ID" required:"true"`
	InstanceID         string `envconfig:"INSTANCE_ID" required:"true"`
	EmulatorHost       string `envconfig:"EMULATOR_HOST" required:"false"`
	WriteRequestsTopic string `envconfig:"WRITE_REQUESTS_TOPIC" required:"true"`
}

// Pubsub implements beneath.StreamsDriver
type Pubsub struct {
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
