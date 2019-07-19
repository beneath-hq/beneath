package gateway

import (
	"fmt"
	"log"
	"net/http"

	"github.com/beneath-core/beneath-go/control/auth"
	"github.com/beneath-core/beneath-go/control/model"
	"github.com/beneath-core/beneath-go/core/httputil"
	"github.com/beneath-core/beneath-go/core/jsonutil"
	pb "github.com/beneath-core/beneath-go/proto"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	uuid "github.com/satori/go.uuid"
)

// ListenAndServeHTTP serves a HTTP API
func ListenAndServeHTTP(port int) error {
	log.Printf("HTTP server running on port %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), httpHandler())
}

func httpHandler() http.Handler {
	handler := chi.NewRouter()

	// handler.Use(middleware.RealIP) // TODO: Uncomment if IPs are a problem behind nginx
	handler.Use(middleware.Logger)
	handler.Use(middleware.Recoverer)
	handler.Use(auth.HTTPMiddleware)

	// TODO: Add health checks

	// TODO: Add graphql
	// GraphQL endpoints
	// handler.Get("/graphql")
	// handler.Get("/projects/{projectName}/graphql")

	// REST endpoints
	handler.Method("GET", "/projects/{projectName}/streams/{streamName}/details", httputil.AppHandler(getStreamDetails))
	handler.Method("GET", "/projects/{projectName}/streams/{streamName}", httputil.AppHandler(getFromProjectAndStream))
	handler.Method("GET", "/streams/instances/{instanceID}", httputil.AppHandler(getFromInstance))
	handler.Method("POST", "/streams/instances/{instanceID}", httputil.AppHandler(postToInstance))

	return handler
}

func getStreamDetails(w http.ResponseWriter, r *http.Request) error {
	// get instance ID
	projectName := chi.URLParam(r, "projectName")
	streamName := chi.URLParam(r, "streamName")
	instanceID := model.FindInstanceIDByNameAndProject(streamName, projectName)
	if instanceID == uuid.Nil {
		return httputil.NewError(404, "instance for stream not found")
	}

	// get stream details
	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		return httputil.NewError(404, "stream not found")
	}

	// create json response
	json, err := jsonutil.Marshal(map[string]interface{}{
		"current_instance_id": instanceID,
		"project_id":          stream.ProjectID,
		"public":              stream.Public,
		"external":            stream.External,
		"batch":               stream.Batch,
		"manual":              stream.Manual,
		"key_fields":          stream.KeyCodec.GetKeyFields(),
		"avro_schema":         stream.AvroCodec.GetSchema(),
	})
	if err != nil {
		return httputil.NewError(500, err.Error())
	}

	// write
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
	return nil
}

func getFromProjectAndStream(w http.ResponseWriter, r *http.Request) error {
	projectName := chi.URLParam(r, "projectName")
	streamName := chi.URLParam(r, "streamName")
	instanceID := model.FindInstanceIDByNameAndProject(streamName, projectName)
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
	key := auth.GetKey(r.Context())

	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		return httputil.NewError(404, "stream not found")
	}

	if !key.ReadsProject(stream.ProjectID) {
		return httputil.NewError(403, "token doesn't grant right to read this stream")
	}

	// TODO
	// Read from BT in accordance with how we end up writing it
	// Support filter, limit, page (see https://docs.hasura.io/1.0/graphql/manual/queries/query-filters.html)

	w.Write([]byte(fmt.Sprintf("Hello Stream Instance %s", instanceID.String())))
	return nil
}

func postToInstance(w http.ResponseWriter, r *http.Request) error {
	// get auth
	key := auth.GetKey(r.Context())

	// get instance ID
	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
	if err != nil {
		return httputil.NewError(404, "instance not found -- malformed ID")
	}

	// get stream
	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		return httputil.NewError(404, "stream not found")
	}

	// check allowed to write stream
	if !key.WritesStream(stream) {
		return httputil.NewError(403, "token doesn't grant right to read this stream")
	}

	// decode json body
	var body interface{}
	err = jsonutil.Unmarshal(r.Body, &body)
	if err != nil {
		return httputil.NewError(400, "request body must be json")
	}

	// get objects passed in body
	var objects []interface{}
	switch bodyT := body.(type) {
	case []interface{}:
		objects = bodyT
	case map[string]interface{}:
		objects = []interface{}{bodyT}
	default:
		return httputil.NewError(400, "request body must be an array or an object")
	}

	// convert objects into records
	records := make([]*pb.Record, len(objects))
	for idx, objV := range objects {
		// check it's a map
		obj, ok := objV.(map[string]interface{})
		if !ok {
			return httputil.NewError(400, fmt.Sprintf("record at index %d is not an object", idx))
		}

		// get meta field
		meta, ok := obj["@meta"].(map[string]interface{})
		if !ok {
			return httputil.NewError(400, "must provide '@meta' field for every record")
		}

		// get sequence number as uint64
		sequenceNumber, err := jsonutil.ParseUint64(meta["sequence_number"])
		if err != nil {
			return httputil.NewError(400, "must provide '@meta.sequence_number' as number or numeric string for every record (must fit in a 64-bit unsigned integer)")
		}

		// encode as avro
		avroData, err := stream.AvroCodec.Marshal(obj)
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// compute key (only used for size check)
		keyData, err := stream.KeyCodec.Marshal(obj)
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// check sizes
		err = Engine.CheckSize(len(keyData), len(avroData))
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// save the record
		records[idx] = &pb.Record{
			AvroData:       avroData,
			SequenceNumber: sequenceNumber,
		}
	}

	// queue write request (publishes to Pubsub)
	err = Engine.Streams.QueueWriteRequest(&pb.WriteRecordsRequest{
		InstanceId: instanceID.Bytes(),
		Records:    records,
	})
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// Done
	return nil
}
