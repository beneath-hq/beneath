package data

import (
	"context"
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/infrastructure/engine/driver"
	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
	pb "gitlab.com/beneath-hq/beneath/server/data/grpc/proto"
	"gitlab.com/beneath-hq/beneath/services/middleware"
)

// ReadRequest is a request to HandleRead
type ReadRequest struct {
	Cursor     []byte
	Limit      int32
	ReturnPB   bool
	ReturnJSON bool
}

// ReadResponse is a result from HandleRead
type ReadResponse struct {
	PB         []*pb.Record
	JSON       []interface{}
	NextCursor []byte
	InstanceID uuid.UUID
	JobID      uuid.UUID
}

type readTags struct {
	InstanceID  *uuid.UUID `json:"instance,omitempty"`
	JobID       *uuid.UUID `json:"job,omitempty"`
	Cursor      []byte     `json:"cursor,omitempty"`
	Limit       int32      `json:"limit,omitempty"`
	BytesRead   int        `json:"bytes,omitempty"`
	RecordsRead int        `json:"records,omitempty"`
}

const (
	defaultReadLimit = 50
	maxReadLimit     = 1000
)

// HandleRead handles a read request
func (s *Service) HandleRead(ctx context.Context, req *ReadRequest) (*ReadResponse, *Error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, newErrorf(http.StatusUnauthorized, "not authenticated")
	}

	// set payload
	payload := readTags{
		Cursor: req.Cursor,
		Limit:  req.Limit,
	}
	middleware.SetTagsPayload(ctx, payload)

	// parse cursor
	cursor, err := CursorFromBytes(req.Cursor)
	if err != nil {
		return nil, newErrorf(http.StatusBadRequest, "%s", err.Error())
	}

	// check limit
	if req.Limit == 0 {
		req.Limit = defaultReadLimit
	} else if req.Limit > maxReadLimit {
		return nil, newErrorf(http.StatusBadRequest, "limit exceeds maximum (%d)", maxReadLimit)
	}

	// check quota
	err = s.Metrics.CheckReadQuota(ctx, secret)
	if err != nil {
		return nil, newError(http.StatusTooManyRequests, err.Error())
	}

	// make response
	resp := &ReadResponse{}

	// get result iterator
	var it driver.RecordsIterator
	streamID := uuid.Nil
	instanceID := uuid.Nil

	if cursor.GetType() == LogCursorType || cursor.GetType() == IndexCursorType {
		// get instanceID
		instanceID = cursor.GetID()
		resp.InstanceID = instanceID

		// update tags
		payload.InstanceID = &instanceID
		middleware.SetTagsPayload(ctx, payload)

		// get cached stream
		stream := s.Streams.FindCachedInstance(ctx, instanceID)
		if stream == nil {
			return nil, newError(http.StatusNotFound, "stream not found")
		}
		streamID = stream.StreamID

		// check permissions
		perms := s.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Public)
		if !perms.Read {
			return nil, newErrorf(http.StatusForbidden, "token doesn't grant right to read this stream")
		}

		// get it
		it, err = s.Engine.Lookup.ReadCursor(ctx, stream, stream, models.EfficientStreamInstance(instanceID), cursor.GetPayload(), int(req.Limit))
		if err != nil {
			return nil, newErrorf(http.StatusBadRequest, "%s", err.Error())
		}
	} else if cursor.GetType() == WarehouseCursorType {
		// get jobID
		jobID := cursor.GetID()
		resp.JobID = jobID

		// update tags
		payload.JobID = &jobID
		middleware.SetTagsPayload(ctx, payload)

		// get it
		it, err = s.Engine.Warehouse.ReadWarehouseCursor(ctx, cursor.GetPayload(), int(req.Limit))
		if err != nil {
			return nil, newErrorf(http.StatusBadRequest, "%s", err.Error())
		}
	} else {
		panic(fmt.Errorf("cannot handle cursor type"))
	}

	// build response
	nbytes := 0
	nrecords := 0
	if req.ReturnPB {
		// save records as PBs
		for it.Next() {
			record := it.Record()

			recordProto := &pb.Record{
				AvroData:  record.GetAvro(),
				Timestamp: timeutil.UnixMilli(record.GetTimestamp()),
			}

			nbytes += len(recordProto.AvroData)
			resp.PB = append(resp.PB, recordProto)
		}
		nrecords = len(resp.PB)
	} else if req.ReturnJSON {
		// save records as JSON-friendly maps
		for it.Next() {
			record := it.Record()
			avro := record.GetAvro() // to calculate bytes read
			data := record.GetJSON()

			if cursor.GetType() == WarehouseCursorType {
				data["@meta"] = map[string]interface{}{
					"timestamp": timeutil.UnixMilli(record.GetTimestamp()),
				}
			} else {
				data["@meta"] = map[string]interface{}{
					"key":       record.GetPrimaryKey(),
					"timestamp": timeutil.UnixMilli(record.GetTimestamp()),
				}
			}

			// track
			nbytes += len(avro)
			resp.JSON = append(resp.JSON, data)
		}
		nrecords = len(resp.JSON)
	}

	// set next cursor
	resp.NextCursor = wrapCursor(cursor.GetType(), cursor.GetID(), it.NextCursor())

	// track read metrics
	s.Metrics.TrackRead(ctx, secret, streamID, instanceID, int64(nrecords), int64(nbytes))

	// update log message
	payload.BytesRead = nbytes
	payload.RecordsRead = nrecords
	middleware.SetTagsPayload(ctx, payload)

	// done
	return resp, nil
}
