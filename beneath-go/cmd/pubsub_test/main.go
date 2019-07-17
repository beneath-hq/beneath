// Command line tool to run/rollback migrations
// Essentially stolen from https://github.com/go-pg/migrations/blob/master/example/main.go

package main

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	// prepare pubsub client
	client, err := pubsub.NewClient(context.Background(), "beneathcrypto")
	if err != nil {
		log.Fatalf("could not create pubsub client: %v", err)
	}

	// create test topic
	topic, err := client.CreateTopic(context.Background(), "test-topic")
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Fatalf("error creating topic: %v", err)
		}
	}

	// publish
	ctx := context.Background()
	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte("hello world"),
	})

	// blocks until ack'ed by pubsub
	_, err = result.Get(ctx)
	if err != nil {
		log.Fatalf("error publishing: %v", err)
	}
}
