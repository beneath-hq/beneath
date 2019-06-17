package pubsub

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	pb "github.com/beneath-core/beneath-gateway/beneath/beneath_proto"
	"github.com/golang/protobuf/proto"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// configSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type configSpecification struct {
	ProjectID          string `envconfig:"PROJECT_ID" required:"true"`
	InstanceID         string `envconfig:"INSTANCE_ID" required:"true"`
	WriteRequestsTopic string `envconfig:"WRITE_REQUESTS_TOPIC" required:"true"`
}

// Pubsub implements beneath.StreamsDriver
type Pubsub struct {
	client             *pubsub.Client
	writeRequestsTopic *pubsub.Topic
}

// New returns a new
func New() *Pubsub {
	// parse config from env
	var config configSpecification
	err := envconfig.Process("beneath_pubsub", &config)
	if err != nil {
		log.Fatalf("pubsub: %s", err.Error())
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
			log.Fatalf("error creating topic: %v", err)
		}
	}

	// create instance
	return &Pubsub{
		client:             client,
		writeRequestsTopic: writeRequestsTopic,
	}
}

// GetMaxMessageSize implements beneath.StreamsDriver
func (p *Pubsub) GetMaxMessageSize() int {
	return 10000000
}

// PushWriteRequest implements beneath.StreamsDriver
func (p *Pubsub) PushWriteRequest(req *pb.WriteInternalRecordsRequest) error {
	// encode message
	msg, err := proto.Marshal(req)
	if err != nil {
		log.Panicf("error marshalling WriteInternalRecordsRequest: %v", err)
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
	result := p.writeRequestsTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	// blocks until ack'ed by pubsub
	_, err = result.Get(ctx)
	return err
}
