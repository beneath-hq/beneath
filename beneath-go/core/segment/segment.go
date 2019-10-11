package segment

import (
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"
	analytics "gopkg.in/segmentio/analytics-go.v3"

	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/core/middleware"
)

// Client enables us to send data to Segment
var Client analytics.Client

// InitClient initializes the Segment client
func InitClient(segmentKey string) {
	Client = analytics.New(segmentKey)
}

// TrackHTTP logs a new HTTP-related event to segment
func TrackHTTP(r *http.Request, name string, payload interface{}) {
	if Client == nil {
		panic(fmt.Errorf("Must call segment.InitClient before calling segment.Track"))
	}

	tags := middleware.GetTags(r.Context())
	props := analytics.NewProperties().
		Set("ip", r.RemoteAddr).
		Set("payload", payload)

	// SecretID, UserID, and ServiceID can be null
	userID := ""
	if tags.Secret != nil {
		props.Set("secret", tags.Secret.SecretID.String())
		if tags.Secret.UserID != nil {
			userID = tags.Secret.UserID.String()
		} else if tags.Secret.ServiceID != nil {
			userID = tags.Secret.ServiceID.String()
		} else {
			panic(fmt.Errorf("expected UserID or ServiceID to be set"))
		}
	}

	// anonymous id
	aid := ""
	if tags.AnonymousID != uuid.Nil {
		aid = tags.AnonymousID.String()
	}

	// Doesn't make sense to log events we know nothing about
	if userID == "" && aid == "" {
		return
	}

	err := Client.Enqueue(analytics.Track{
		Event:       name,
		UserId:      userID,
		AnonymousId: aid,
		Properties:  props,
	})
	if err != nil {
		log.S.Errorf("Segment error: %s", err.Error())
	}
}
