package api

import (
	"context"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/gateway/util"
	"gitlab.com/beneath-hq/beneath/hub"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
)

// QueryLogRequest is a request to HandleQueryLog
type QueryLogRequest struct {
	InstanceID uuid.UUID
	Partitions int32
	Peek       bool
}

// QueryLogResponse is a result from HandleQueryLog
type QueryLogResponse struct {
	ReplayCursors [][]byte
	ChangeCursors [][]byte
}

type queryLogTags struct {
	InstanceID uuid.UUID `json:"instance,omitempty"`
	Partitions int32     `json:"partitions,omitempty"`
	Peek       bool      `json:"peek,omitempty"`
}

// HandleQueryLog handles a log query request
func HandleQueryLog(ctx context.Context, req *QueryLogRequest) (*QueryLogResponse, *Error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, newErrorf(http.StatusUnauthorized, "not authenticated")
	}

	// set payload
	payload := queryLogTags{
		InstanceID: req.InstanceID,
		Partitions: req.Partitions,
		Peek:       req.Peek,
	}
	middleware.SetTagsPayload(ctx, payload)

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, req.InstanceID)
	if stream == nil {
		return nil, newErrorf(http.StatusNotFound, "stream not found")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public)
	if !perms.Read {
		return nil, newErrorf(http.StatusForbidden, "token doesn't grant right to read this stream")
	}

	// run peek query
	if req.Peek {
		// check partitions == 1 on peek
		if req.Partitions > 1 {
			return nil, newErrorf(http.StatusBadRequest, "cannot return more than one partition for a peek")
		}

		// run query
		replayCursor, changeCursor, err := hub.Engine.Lookup.Peek(ctx, stream, stream, entity.EfficientStreamInstance(req.InstanceID))
		if err != nil {
			return nil, newErrorf(http.StatusBadRequest, "error parsing query: %s", err.Error())
		}

		// done
		return &QueryLogResponse{
			ReplayCursors: [][]byte{wrapCursor(util.LogCursorType, req.InstanceID, replayCursor)},
			ChangeCursors: [][]byte{wrapCursor(util.LogCursorType, req.InstanceID, changeCursor)},
		}, nil
	}

	// run normal query
	replayCursors, changeCursors, err := hub.Engine.Lookup.ParseQuery(ctx, stream, stream, entity.EfficientStreamInstance(req.InstanceID), nil, false, int(req.Partitions))
	if err != nil {
		return nil, newErrorf(http.StatusBadRequest, "error parsing query: %s", err.Error())
	}

	// done
	return &QueryLogResponse{
		ReplayCursors: wrapCursors(util.LogCursorType, req.InstanceID, replayCursors),
		ChangeCursors: wrapCursors(util.LogCursorType, req.InstanceID, changeCursors),
	}, nil
}
