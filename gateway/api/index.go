package api

import (
	"context"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/gateway/util"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/queryparse"
)

// QueryIndexRequest is a request to HandleQueryIndex
type QueryIndexRequest struct {
	InstanceID uuid.UUID
	Partitions int32
	Filter     string
}

// QueryIndexResponse is a result from HandleQueryIndex
type QueryIndexResponse struct {
	ReplayCursors [][]byte
	ChangeCursors [][]byte
}

type queryIndexTags struct {
	InstanceID uuid.UUID `json:"instance,omitempty"`
	Partitions int32     `json:"partitions,omitempty"`
	Filter     string    `json:"filter,omitempty"`
}

// HandleQueryIndex handles an index query request
func HandleQueryIndex(ctx context.Context, req *QueryIndexRequest) (*QueryIndexResponse, *Error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, newErrorf(http.StatusUnauthorized, "not authenticated")
	}

	// set payload
	payload := queryIndexTags{
		InstanceID: req.InstanceID,
		Partitions: req.Partitions,
		Filter:     req.Filter,
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

	// get filter
	where, err := queryparse.JSONStringToQuery(req.Filter)
	if err != nil {
		return nil, newErrorf(http.StatusBadRequest, "couldn't parse 'where': %s", err.Error())
	}

	// run query
	replayCursors, changeCursors, err := hub.Engine.Lookup.ParseQuery(ctx, stream, stream, entity.EfficientStreamInstance(req.InstanceID), where, true, int(req.Partitions))
	if err != nil {
		return nil, newErrorf(http.StatusBadRequest, "error parsing query: %s", err.Error())
	}

	// done
	return &QueryIndexResponse{
		ReplayCursors: wrapCursors(util.IndexCursorType, req.InstanceID, replayCursors),
		ChangeCursors: wrapCursors(util.LogCursorType, req.InstanceID, changeCursors),
	}, nil
}
