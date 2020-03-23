package http

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/gateway"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/internal/middleware"
	"github.com/beneath-core/pkg/httputil"
	"github.com/beneath-core/pkg/jsonutil"
	"github.com/beneath-core/pkg/timeutil"
)

func getLatestFromProjectAndStream(w http.ResponseWriter, r *http.Request) error {
	projectName := toBackendName(chi.URLParam(r, "projectName"))
	streamName := toBackendName(chi.URLParam(r, "streamName"))
	instanceID := entity.FindInstanceIDByNameAndProject(r.Context(), streamName, projectName)
	if instanceID == uuid.Nil {
		return httputil.NewError(404, "instance for stream not found")
	}

	return getLatestFromInstanceID(w, r, instanceID)
}

func getLatestFromInstance(w http.ResponseWriter, r *http.Request) error {
	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
	if err != nil {
		return httputil.NewError(404, "instance not found -- malformed ID")
	}

	return getLatestFromInstanceID(w, r, instanceID)
}

func getLatestFromInstanceID(w http.ResponseWriter, r *http.Request, instanceID uuid.UUID) error {
	// get auth
	secret := middleware.GetSecret(r.Context())

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(r.Context(), instanceID)
	if stream == nil {
		return httputil.NewError(404, "stream not found")
	}

	// set log payload
	payload := peekTags{
		InstanceID: instanceID.String(),
	}
	middleware.SetTagsPayload(r.Context(), payload)

	// check allowed to read stream
	perms := secret.StreamPermissions(r.Context(), stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Read {
		return httputil.NewError(403, "secret doesn't grant right to read this stream")
	}

	// if batch, check committed (sort of silly to peek batch data, but we'll allow it)
	if stream.Batch && !stream.Committed {
		return httputil.NewError(400, "batch has not yet been committed, and so can't be read")
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
	body, err := parsePeekBody(r)
	if err != nil {
		return err
	}

	// create cursor if not provided
	if body.Cursor == nil {
		// run query
		rewindCursor, changeCursor, err := hub.Engine.Lookup.Peek(r.Context(), stream, stream, stream)
		if err != nil {
			// there's no user input to cause an error, so it's an internal errr
			panic(err)
		}

		// set
		body.Cursor = rewindCursor
		if changeCursor != nil {
			cursors["changes"] = changeCursor
		}
	}

	// update payload
	payload.Limit = body.Limit
	middleware.SetTagsPayload(r.Context(), payload)

	// read rows from engine
	it, err := hub.Engine.Lookup.ReadCursor(r.Context(), stream, stream, stream, body.Cursor, body.Limit)
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// begin json object
	result := make([]interface{}, 0, body.Limit)
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
		cursors["next"] = next
	}

	// prepare result for encoding
	encode := map[string]interface{}{
		"data":   result,
		"cursor": cursors,
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

type peekBody struct {
	Cursor []byte
	Limit  int
}

func parsePeekBody(r *http.Request) (peekBody, error) {
	// defaults
	body := peekBody{}

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
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		// convert cursor to bytes
		cursor, err := base64.StdEncoding.DecodeString(cursorStr)
		if err != nil {
			return body, httputil.NewError(400, "invalid cursor")
		}
		body.Cursor = cursor
	}

	// check limit
	if body.Limit == 0 {
		body.Limit = defaultReadLimit
	} else if body.Limit > maxReadLimit {
		return body, httputil.NewError(400, fmt.Sprintf("limit exceeds maximum of %d", maxReadLimit))
	}

	// done
	return body, nil
}
