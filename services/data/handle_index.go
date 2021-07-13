package data

import (
	"context"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/pkg/queryparse"
	"github.com/beneath-hq/beneath/services/middleware"
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
func (s *Service) HandleQueryIndex(ctx context.Context, req *QueryIndexRequest) (*QueryIndexResponse, *Error) {
	// get auth
	secret := middleware.GetSecret(ctx)

	// set payload
	payload := queryIndexTags{
		InstanceID: req.InstanceID,
		Partitions: req.Partitions,
		Filter:     req.Filter,
	}
	middleware.SetTagsPayload(ctx, payload)

	// get cached table
	table := s.Tables.FindCachedInstance(ctx, req.InstanceID)
	if table == nil {
		return nil, newErrorf(http.StatusNotFound, "table not found")
	}

	// check permissions
	perms := s.Permissions.TablePermissionsForSecret(ctx, secret, table.TableID, table.ProjectID, table.Public)
	if !perms.Read {
		return nil, newErrorf(http.StatusForbidden, "token doesn't grant right to read this table")
	}

	// get filter
	where, err := queryparse.StringToQuery(req.Filter)
	if err != nil {
		return nil, newErrorf(http.StatusBadRequest, "couldn't parse 'where': %s", err.Error())
	}

	// run query
	replayCursors, changeCursors, err := s.Engine.Lookup.ParseQuery(ctx, table, table, models.EfficientTableInstance(req.InstanceID), where, true, int(req.Partitions))
	if err != nil {
		return nil, newErrorf(http.StatusBadRequest, "error parsing query: %s", err.Error())
	}

	// done
	return &QueryIndexResponse{
		ReplayCursors: wrapCursors(IndexCursorType, req.InstanceID, replayCursors),
		ChangeCursors: wrapCursors(LogCursorType, req.InstanceID, changeCursors),
	}, nil
}
