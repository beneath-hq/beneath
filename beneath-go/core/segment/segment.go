package segment

import (
	"os"

	analytics "gopkg.in/segmentio/analytics-go.v3"
)

// Client enables us to send data to Segment
var Client analytics.Client

// InitClient initializes the Segment client
func InitClient(segmentKey string) {
	// if in development, set segment key to our development source key
	// TODO: I might want to create new Segment sources for the server-side analytics. Currently, we have sources explicitly labeled as front-end. Different sources have different keys.
	if segmentKey == "" {
		os.Setenv("SEGMENT_SERVERSIDE_KEY", "NIql6eQCaA5DKDXkGutNCNfFB2o97fvx")
	}

	Client = analytics.New(segmentKey)
}
