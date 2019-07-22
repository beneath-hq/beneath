package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	// set environment variables
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	os.Setenv("PUBSUB_PROJECT_ID", "beneathcrypto")

	// prepare pubsub client
	client, err := pubsub.NewClient(context.Background(), "beneathcrypto")
	if err != nil {
		log.Fatalf("could not create pubsub client: %v", err)
	}

	// create test topic
	topicName := "test-topic"
	topic, err := client.CreateTopic(context.Background(), topicName)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Fatalf("error creating topic: %v", err)
		} else {
			topic = client.Topic(topicName)
		}
	}

	// publish
	ctx := context.Background()
	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte("hello world!!!"),
	})

	// blocks until ack'ed by pubsub
	_, err = result.Get(ctx)
	if err != nil {
		log.Fatalf("error publishing: %v", err)
	}

	// cleanup
	topic.Stop()
	client.Close()
}
