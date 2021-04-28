package data

import (
	"context"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/models"
)

// SubscribeRequest is a request to HandleSubscribe
type SubscribeRequest struct {
	Secret models.Secret
	Cursor []byte
	Cb     func(msg SubscriptionMessage)
}

// SubscribeResponse is a result from HandleSubscribe
type SubscribeResponse struct {
	InstanceID uuid.UUID
	Cancel     context.CancelFunc
}

// HandleSubscribe starts a new subscription.
// Note: Basically just packaging a bunch of things the HTTP and GRPC servers need to
// do before calling s.Subscribe.
func (s *Service) HandleSubscribe(ctx context.Context, req *SubscribeRequest) (*SubscribeResponse, error) {
	// parse cursor
	cursor, err := CursorFromBytes(req.Cursor)
	if err != nil {
		return nil, newError(http.StatusBadRequest, err.Error())
	}

	// ensure log cursor
	if cursor.GetType() != LogCursorType {
		return nil, newErrorf(http.StatusBadRequest, "cannot subscribe to non-log cursor")
	}

	// read instanceID
	instanceID := cursor.GetID()

	// get cached stream
	stream := s.Streams.FindCachedInstance(ctx, instanceID)
	if stream == nil {
		return nil, newErrorf(http.StatusNotFound, "stream not found")
	}

	// check permissions
	perms := s.Permissions.StreamPermissionsForSecret(ctx, req.Secret, stream.StreamID, stream.ProjectID, stream.Public)
	if !perms.Read {
		return nil, newErrorf(http.StatusForbidden, "token doesn't grant right to read this stream")
	}

	// check usage
	err = s.Usage.CheckReadQuota(ctx, req.Secret)
	if err != nil {
		return nil, newError(http.StatusTooManyRequests, err.Error())
	}

	// start subscription
	cancel, err := s.Subscribe(instanceID, cursor.GetPayload(), req.Cb)
	if err != nil {
		return nil, err
	}

	return &SubscribeResponse{
		InstanceID: instanceID,
		Cancel:     cancel,
	}, nil
}
