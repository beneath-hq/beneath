package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
)

var topics []*pubsub.Topic
var subs []*pubsub.Subscription

func main() {

	// set environment variables
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	os.Setenv("PUBSUB_PROJECT_ID", "beneathcrypto")

	// prepare pubsub client
	client, err := pubsub.NewClient(context.Background(), "beneathcrypto")
	if err != nil {
		log.Fatalf("could not create pubsub client: %v", err)
	}

	// create subscriber to topic
	sub, err := client.CreateSubscription(context.Background(), "mySubscriptionName3", pubsub.SubscriptionConfig{
		Topic:       client.Topic("test-topic"),
		AckDeadline: 20 * time.Second,
	})
	if err != nil {
		log.Printf(err.Error())
		log.Printf("could not subscribe to topic")
	} else {
		log.Printf("Created subscription: %v\n", sub)
	}

	// consume messages
	ctx := context.Background()
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		fmt.Printf("Got message: %q\n", string(msg.Data))
	})
	if err != nil {
		log.Fatalf("Error!")
	}
}
