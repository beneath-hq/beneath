package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/beneath-core/pkg/queryparse"

	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/pkg/httputil"
	"github.com/beneath-core/pkg/jsonutil"
	"github.com/beneath-core/internal/middleware"
	"github.com/beneath-core/pkg/timeutil"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/gateway"
)

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

	// prepare cursors info
	cursors := make(map[string][]byte)

	// read body
	body, err := parseQueryBody(r)
	if err != nil {
		return err
	}

	// create cursor if not provided
	if body.Cursor == nil {
		// make query
		filter, err := queryparse.JSONStringToQuery(string(body.Filter))
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("couldn't parse filter query: %s", err.Error()))
		}

		// run query
		replayCursors, changeCursors, err := hub.Engine.Lookup.ParseQuery(r.Context(), stream, stream, stream, filter, body.Compact, 1)
		if err != nil {
			return httputil.NewError(400, err.Error())
		}

		// set
		body.Cursor = replayCursors[0]
		if len(changeCursors) != 0 {
			cursors["changes"] = changeCursors[0]
		}
	}

	// update payload
	payload.Limit = body.Limit
	payload.Filter = string(body.Filter)
	payload.Compact = body.Compact
	payload.Cursor = body.Cursor
	middleware.SetTagsPayload(r.Context(), payload)

	// read rows from engine
	it, err := hub.Engine.Lookup.ReadCursor(r.Context(), stream, stream, stream, body.Cursor, body.Limit)
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// begin json object
	result := make([]interface{}, 0, 1)
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

		// set timestamp
		data["@meta"] = map[string]int64{"timestamp": timeutil.UnixMilli(record.GetTimestamp())}

		// track
		result = append(result, data)
		bytesRead += len(avro)
	}

	// set next cursor
	if next := it.NextCursor(); next != nil {
		cursors["next"] = next
	}

	// prepare result for encoding
	encode := map[string]interface{}{
		"data":    result,
		"cursors": cursors,
	}

	// write and finish
	w.Header().Set("Content-Type", "application/json")
	err = jsonutil.MarshalWriter(encode, w)
	if err != nil {
		return err
	}

	// track read metrics
	gateway.Metrics.TrackRead(stream.StreamID, int64(len(result)), int64(bytesRead))
	if !secret.IsAnonymous() {
		gateway.Metrics.TrackRead(secret.GetOwnerID(), int64(len(result)), int64(bytesRead))
	}

	// update payload
	payload.BytesRead = bytesRead
	middleware.SetTagsPayload(r.Context(), payload)

	return nil
}

type queryBody struct {
	Limit   int
	Cursor  []byte
	Filter  json.RawMessage
	Compact bool
}

func parseQueryBody(r *http.Request) (queryBody, error) {
	// defaults
	body := queryBody{
		Compact: true,
	}

	// parse body
	err := jsonutil.Unmarshal(r.Body, &body)
	if err != nil {
		if err != io.EOF {
			return body, httputil.NewError(400, "couldn't parse body -- is it valid JSON?")
		}

		// parse from url parameters instead

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
		if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
			// convert cursor to bytes
			cursor, err := base64.StdEncoding.DecodeString(cursorStr)
			if err != nil {
				return body, httputil.NewError(400, "invalid cursor")
			}
			body.Cursor = cursor
		}

		// read compact
		if compactRaw := r.URL.Query().Get("compact"); compactRaw != "" {
			if compactRaw == "true" {
				body.Compact = true
			} else if compactRaw == "false" {
				body.Compact = false
			} else {
				return body, httputil.NewError(400, fmt.Sprintf("expected 'compact' parameter to be 'true' or 'false'"))
			}
		}

		// read filter
		if filter := r.URL.Query().Get("filter"); filter != "" {
			body.Filter = []byte(filter)
		}
	}

	// check limit
	if body.Limit == 0 {
		body.Limit = defaultReadLimit
	} else if body.Limit > maxReadLimit {
		return body, httputil.NewError(400, fmt.Sprintf("limit exceeds maximum of %d", maxReadLimit))
	}

	// check query isn't provided if cursor is set
	if body.Cursor != nil {
		if body.Filter != nil || !body.Compact {
			return body, httputil.NewError(400, "you cannot provide 'filter' or 'compact' parameters along with 'cursor'")
		}
	}

	// done
	return body, nil
}
