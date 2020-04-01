package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/mr-tron/base58"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/gateway"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/internal/middleware"
	"github.com/beneath-core/pkg/httputil"
	"github.com/beneath-core/pkg/jsonutil"
	"github.com/beneath-core/pkg/queryparse"
	"github.com/beneath-core/pkg/timeutil"
)

const (
	queryBodyTypeNone  = ""
	queryBodyTypeIndex = "index"
	queryBodyTypeLog   = "log"
)

type queryBody struct {
	Limit       int             `json:"limit,omitempty"`
	Cursor      string          `json:"cursor,omitempty"`
	CursorBytes []byte          `json:"-"`
	Type        string          `json:"type,omitempty"`
	Filter      json.RawMessage `json:"filter,omitempty"`
	Peek        bool            `json:"peek,omitempty"`
}

type queryTags struct {
	InstanceID string    `json:"instance_id,omitempty"`
	Body       queryBody `json:"body,omitempty"`
	BytesRead  int       `json:"bytes,omitempty"`
}

type queryResponseMeta struct {
	InstanceID   uuid.UUID `json:"instance_id,omitempty"`
	NextCursor   string    `json:"next_cursor,omitempty"`
	ChangeCursor string    `json:"change_cursor,omitempty"`
}

func getFromProjectAndStream(w http.ResponseWriter, r *http.Request) error {
	projectName := toBackendName(chi.URLParam(r, "projectName"))
	streamName := toBackendName(chi.URLParam(r, "streamName"))
	instanceID := entity.FindInstanceIDByNameAndProject(r.Context(), streamName, projectName)
	if instanceID == uuid.Nil {
		return httputil.NewError(404, "instance for stream not found")
	}

	return getFromInstanceID(w, r, instanceID)
}

func getFromInstance(w http.ResponseWriter, r *http.Request) error {
	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
	if err != nil {
		return httputil.NewError(404, "instance not found -- malformed ID")
	}

	return getFromInstanceID(w, r, instanceID)
}

func getFromInstanceID(w http.ResponseWriter, r *http.Request, instanceID uuid.UUID) error {
	// get auth
	secret := middleware.GetSecret(r.Context())

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(r.Context(), instanceID)
	if stream == nil {
		return httputil.NewError(404, "stream not found")
	}

	// set log payload
	payload := queryTags{
		InstanceID: instanceID.String(),
	}
	middleware.SetTagsPayload(r.Context(), payload)

	// check allowed to read stream
	perms := secret.StreamPermissions(r.Context(), stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Read {
		return httputil.NewError(403, "secret doesn't grant right to read this stream")
	}

	// check quota
	if !secret.IsAnonymous() {
		usage := gateway.Metrics.GetCurrentUsage(r.Context(), secret.GetOwnerID())
		ok := secret.CheckReadQuota(usage)
		if !ok {
			return httputil.NewError(429, "you have exhausted your monthly quota")
		}
	}

	// read body
	body, err := parseQueryBody(r)
	if err != nil {
		return err
	}

	// prepare meta
	meta := queryResponseMeta{
		InstanceID: instanceID,
	}

	// create cursor if not provided
	// TODO: clean up this mess once we've refactored the engine interfaces
	if body.CursorBytes == nil {
		if body.Type == queryBodyTypeIndex {
			// make filter
			filter, err := queryparse.JSONStringToQuery(string(body.Filter))
			if err != nil {
				return httputil.NewError(400, fmt.Sprintf("couldn't parse 'filter': %s", err.Error()))
			}

			// run query
			replayCursors, changeCursors, err := hub.Engine.Lookup.ParseQuery(r.Context(), stream, stream, stream, filter, true, 1)
			if err != nil {
				return httputil.NewError(400, err.Error())
			}

			// set cursors
			body.CursorBytes = replayCursors[0]
			if len(changeCursors) != 0 {
				meta.ChangeCursor = base58.Encode(changeCursors[0])
			}
		} else if body.Type == queryBodyTypeLog {
			// depends on whether it's a peek
			var replayCursor, changeCursor []byte
			if body.Peek {
				replayCursor, changeCursor, err = hub.Engine.Lookup.Peek(r.Context(), stream, stream, stream)
				if err != nil {
					return httputil.NewError(400, err.Error())
				}
			} else {
				replayCursors, changeCursors, err := hub.Engine.Lookup.ParseQuery(r.Context(), stream, stream, stream, nil, false, 1)
				if err != nil {
					return httputil.NewError(400, err.Error())
				}
				replayCursor = replayCursors[0]
				if len(changeCursors) > 0 {
					changeCursor = changeCursors[0]
				}
			}

			// set cursors
			body.CursorBytes = replayCursor
			if changeCursor != nil {
				meta.ChangeCursor = base58.Encode(changeCursor)
			}
		}
	}

	// update payload
	payload.Body = body
	middleware.SetTagsPayload(r.Context(), payload)

	// read rows from engine
	it, err := hub.Engine.Lookup.ReadCursor(r.Context(), stream, stream, stream, body.CursorBytes, body.Limit)
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// how much memory to pre-allocate for result
	cap := body.Limit
	if body.Type == queryBodyTypeIndex && body.Filter != nil {
		// underlying assumption: there'll be a lot of unique lookups over REST
		cap = 1
	}

	// begin json object
	result := make([]interface{}, 0, cap)
	bytesRead := 0

	for it.Next() {
		record := it.Record()
		avro := record.GetAvro()

		// decode avro
		data, err := stream.Codec.UnmarshalAvro(avro)
		if err != nil {
			return httputil.NewError(400, err.Error())
		}

		// convert to json friendly
		data, err = stream.Codec.ConvertFromAvroNative(data, true)
		if err != nil {
			return httputil.NewError(400, err.Error())
		}

		// set meta
		data["@meta"] = map[string]interface{}{
			"key":       record.GetPrimaryKey(),
			"timestamp": timeutil.UnixMilli(record.GetTimestamp()),
		}

		// track
		result = append(result, data)
		bytesRead += len(avro)
	}

	// set next cursor
	if next := it.NextCursor(); next != nil {
		meta.NextCursor = base58.Encode(next)
	}

	// prepare result for encoding
	encode := map[string]interface{}{
		"meta": meta,
		"data": result,
	}

	// write and finish
	w.Header().Set("Content-Type", "application/json")
	err = jsonutil.MarshalWriter(encode, w)
	if err != nil {
		return err
	}

	// track read metrics
	gateway.Metrics.TrackRead(stream.StreamID, int64(len(result)), int64(bytesRead))
	gateway.Metrics.TrackRead(instanceID, int64(len(result)), int64(bytesRead))
	if !secret.IsAnonymous() {
		gateway.Metrics.TrackRead(secret.GetOwnerID(), int64(len(result)), int64(bytesRead))
	}

	// update payload
	payload.BytesRead = bytesRead
	middleware.SetTagsPayload(r.Context(), payload)

	return nil
}

func parseQueryBody(r *http.Request) (queryBody, error) {
	// defaults
	body := queryBody{}

	// parse body
	err := jsonutil.Unmarshal(r.Body, &body)
	if err != nil {
		if err != io.EOF {
			return body, httputil.NewError(400, "couldn't parse body -- is it valid JSON?")
		}
	}

	// url parameters, if present, supercede body

	// read limit
	if limitRaw := r.URL.Query().Get("limit"); limitRaw != "" {
		// get limit
		limit, err := parseLimit(limitRaw)
		if err != nil {
			return body, httputil.NewError(400, err.Error())
		}
		body.Limit = limit
	}

	// read cursor
	body.Cursor = r.URL.Query().Get("cursor")

	// read type
	body.Type = r.URL.Query().Get("type")

	// read peek
	if peekRaw := r.URL.Query().Get("peek"); peekRaw != "" {
		if peekRaw == "true" {
			body.Peek = true
		} else if peekRaw == "false" {
			body.Peek = false
		} else {
			return body, httputil.NewError(400, fmt.Sprintf("expected 'compact' parameter to be 'true' or 'false'"))
		}
	}

	// read filter
	if filter := r.URL.Query().Get("filter"); filter != "" {
		body.Filter = []byte(filter)
	}

	// check params are consistent

	// check cursor
	if body.Cursor != "" {
		cursor, err := base58.Decode(body.Cursor)
		if err != nil {
			return body, httputil.NewError(400, "invalid cursor")
		}
		body.CursorBytes = cursor
	}

	// check limit
	if body.Limit == 0 {
		body.Limit = defaultReadLimit
	} else if body.Limit > maxReadLimit {
		return body, httputil.NewError(400, fmt.Sprintf("limit exceeds maximum of %d", maxReadLimit))
	}

	// check nothing else if cursor is set
	if body.CursorBytes != nil {
		if body.Type != "" || body.Filter != nil || body.Peek {
			return body, httputil.NewError(400, "you cannot provide query parameters ('type', 'filter' or 'peek') along with a 'cursor'")
		}

		// succesful for a read
		return body, nil
	}

	// checking only query now (not cursor reads)

	// check body type
	if body.Type == queryBodyTypeNone {
		body.Type = queryBodyTypeIndex
	} else if body.Type != queryBodyTypeIndex && body.Type != queryBodyTypeLog {
		return body, httputil.NewError(400, fmt.Sprintf("unrecognized query type '%s'; options are '%s' (default) and '%s'", body.Type, queryBodyTypeIndex, queryBodyTypeLog))
	}

	// check index-only params
	if body.Type == queryBodyTypeLog && body.Filter != nil {
		return body, httputil.NewError(400, "you cannot provide the 'filter' parameter for a 'log' query")
	}

	// check log-only params
	if body.Type == queryBodyTypeIndex && body.Peek {
		return body, httputil.NewError(400, "you cannot provide the 'peek' parameter for a 'query' query")
	}

	// done
	return body, nil
}
