package segment

import (
	analytics "gopkg.in/segmentio/analytics-go.v3"
)

// Client enables us to send data to Segment
var Client analytics.Client

// InitClient initializes the Segment client
func InitClient(segmentKey string) {
	// if in development, set segment key to our development source key
	if segmentKey == "" {
		segmentKey = "NIql6eQCaA5DKDXkGutNCNfFB2o97fvx"
	}

	Client = analytics.New(segmentKey)
}
