package segment

import (
	"fmt"
	"net/http"
	"reflect"

	uuid "github.com/satori/go.uuid"
	analytics "gopkg.in/segmentio/analytics-go.v3"

	"github.com/beneath-core/internal/middleware"
	"github.com/beneath-core/pkg/envutil"
	"github.com/beneath-core/pkg/log"
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
		if envutil.GetEnv() == envutil.Test {
			return
		}
		panic(fmt.Errorf("Must call segment.InitClient before calling segment.Track"))
	}

	tags := middleware.GetTags(r.Context())
	props := analytics.NewProperties().
		Set("ip", r.RemoteAddr).
		Set("payload", payload)

	// SecretID, UserID, and ServiceID can be null
	userID := ""
	if tags.Secret != nil && !reflect.ValueOf(tags.Secret).IsNil() && !tags.Secret.IsAnonymous() {
		props.Set("secret", tags.Secret.GetSecretID().String())
		userID = tags.Secret.GetOwnerID().String()
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
