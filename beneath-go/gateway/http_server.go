package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/beneath-core/beneath-go/core/log"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core/httputil"
	"github.com/beneath-core/beneath-go/core/jsonutil"
	"github.com/beneath-core/beneath-go/core/middleware"
	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/gateway/websockets"
	pb "github.com/beneath-core/beneath-go/proto"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	uuid "github.com/satori/go.uuid"
)

// ListenAndServeHTTP serves a HTTP API
func ListenAndServeHTTP(port int) error {
	log.S.Infow("gateway http started", "port", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), httpHandler())
}

func httpHandler() http.Handler {
	handler := chi.NewRouter()

	handler.Use(chimiddleware.RealIP)
	handler.Use(chimiddleware.DefaultCompress)
	handler.Use(middleware.InjectTags)
	handler.Use(middleware.Logger)
	handler.Use(middleware.Recoverer)
	handler.Use(middleware.Auth)
	handler.Use(middleware.IPRateLimit())

	// Add CORS
	handler.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	// Add health check
	handler.Get("/", healthCheck)
	handler.Get("/healthz", healthCheck)

	// TODO: Add graphql
	// GraphQL endpoints
	// handler.Get("/graphql")
	// handler.Get("/projects/{projectName}/graphql")

	// create websocket broker and start accepting new connections on /ws
	broker := websockets.NewBroker(db.Engine, Metrics)
	handler.Method("GET", "/ws", httputil.AppHandler(broker.HTTPHandler))

	// REST endpoints
	handler.Method("GET", "/projects/{projectName}/streams/{streamName}/details", httputil.AppHandler(getStreamDetails))
	handler.Method("GET", "/projects/{projectName}/streams/{streamName}", httputil.AppHandler(getFromProjectAndStream))
	handler.Method("GET", "/projects/{projectName}/streams/{streamName}/latest", httputil.AppHandler(getLatestFromProjectAndStream))
	handler.Method("GET", "/streams/instances/{instanceID}", httputil.AppHandler(getFromInstance))
	handler.Method("GET", "/streams/instances/{instanceID}/latest", httputil.AppHandler(getLatestFromInstance))
	handler.Method("POST", "/streams/instances/{instanceID}", httputil.AppHandler(postToInstance))

	return handler
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if db.Healthy() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	} else {
		log.S.Errorf("Gateway database health check failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func getStreamDetails(w http.ResponseWriter, r *http.Request) error {
	// get auth
	secret := middleware.GetSecret(r.Context())

	// get instance ID
	projectName := toBackendName(chi.URLParam(r, "projectName"))
	streamName := toBackendName(chi.URLParam(r, "streamName"))
	instanceID := entity.FindInstanceIDByNameAndProject(r.Context(), streamName, projectName)
	if instanceID == uuid.Nil {
		return httputil.NewError(404, "instance for stream not found")
	}

	// get stream details
	stream := entity.FindCachedStreamByCurrentInstanceID(r.Context(), instanceID)
	if stream == nil {
		return httputil.NewError(404, "stream not found")
	}

	// check allowed to read stream
	if !stream.Public {
		perms := secret.StreamPermissions(r.Context(), stream.StreamID, stream.ProjectID, stream.External)
		if !perms.Read {
			return httputil.NewError(403, "secret doesn't grant right to read this stream")
		}
	}

	// create json response
	json, err := jsonutil.Marshal(map[string]interface{}{
		"current_instance_id": instanceID,
		"project_id":          stream.ProjectID,
		"project_name":        stream.ProjectName,
		"stream_name":         stream.StreamName,
		"public":              stream.Public,
		"external":            stream.External,
		"batch":               stream.Batch,
		"manual":              stream.Manual,
		"key_fields":          stream.Codec.GetKeyFields(),
		"avro_schema":         stream.Codec.GetAvroSchema(),
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
	projectName := toBackendName(chi.URLParam(r, "projectName"))
	streamName := toBackendName(chi.URLParam(r, "streamName"))
	instanceID := entity.FindInstanceIDByNameAndProject(r.Context(), streamName, projectName)
	if instanceID == uuid.Nil {
		return httputil.NewError(404, "instance for stream not found")
	}

	return getFromInstanceID(w, r, instanceID)
}

func getLatestFromProjectAndStream(w http.ResponseWriter, r *http.Request) error {
	projectName := toBackendName(chi.URLParam(r, "projectName"))
	streamName := toBackendName(chi.URLParam(r, "streamName"))
	instanceID := entity.FindInstanceIDByNameAndProject(r.Context(), streamName, projectName)
	if instanceID == uuid.Nil {
		return httputil.NewError(404, "instance for stream not found")
	}

	return getLatestFromInstanceID(w, r, instanceID)
}

func getFromInstance(w http.ResponseWriter, r *http.Request) error {
	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
	if err != nil {
		return httputil.NewError(404, "instance not found -- malformed ID")
	}

	return getFromInstanceID(w, r, instanceID)
}

func getLatestFromInstance(w http.ResponseWriter, r *http.Request) error {
	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
	if err != nil {
		return httputil.NewError(404, "instance not found -- malformed ID")
	}

	return getLatestFromInstanceID(w, r, instanceID)
}

func getFromInstanceID(w http.ResponseWriter, r *http.Request, instanceID uuid.UUID) error {
	// get auth
	secret := middleware.GetSecret(r.Context())

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(r.Context(), instanceID)
	if stream == nil {
		return httputil.NewError(404, "stream not found")
	}

	// check allowed to read stream
	if !stream.Public {
		perms := secret.StreamPermissions(r.Context(), stream.StreamID, stream.ProjectID, stream.External)
		if !perms.Read {
			return httputil.NewError(403, "secret doesn't grant right to read this stream")
		}
	}

	// check quota
	if secret != nil {
		usage := Metrics.GetCurrentUsage(r.Context(), secret.BillingID())
		ok := secret.CheckReadQuota(usage)
		if !ok {
			return httputil.NewError(429, "you have exhausted your monthly quota")
		}
	}

	// read body
	var body map[string]interface{}
	err := jsonutil.Unmarshal(r.Body, &body)
	if err == io.EOF {
		// no body -- try reading from url parameters
		body = make(map[string]interface{})

		// read limit
		if limit := r.URL.Query().Get("limit"); limit != "" {
			body["limit"] = limit
		}

		// read where
		if where := r.URL.Query().Get("where"); where != "" {
			var whereParsed interface{}
			err := jsonutil.UnmarshalBytes([]byte(where), &whereParsed)
			if err != nil {
				return httputil.NewError(400, "couldn't parse where url parameter as json")
			}
			body["where"] = whereParsed
		}

		// read page
		if after := r.URL.Query().Get("after"); after != "" {
			var afterParsed interface{}
			err := jsonutil.UnmarshalBytes([]byte(after), &afterParsed)
			if err != nil {
				return httputil.NewError(400, "couldn't parse page url parameter as json")
			}
			body["after"] = afterParsed
		}
	} else if err != nil {
		return httputil.NewError(400, "couldn't parse body -- is it valid JSON?")
	}

	// make sure there's no accidental keys
	for k := range body {
		if k != "where" && k != "limit" && k != "after" {
			return httputil.NewError(400, "unrecognized query key '%s'; valid keys are 'where' and 'limit'", k)
		}
	}

	// get limit
	limit, err := parseLimit(body["limit"])
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// get key range where clause (if no where, it will be nil)
	var whereQuery queryparse.Query
	if body["where"] != nil {
		// get where as map
		where, ok := body["where"].(map[string]interface{})
		if !ok {
			return httputil.NewError(400, "expected 'where' to be a json object")
		}

		// make query
		whereQuery, err = queryparse.JSONToQuery(where)
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("couldn't parse where query: %s", err.Error()))
		}
	}

	// adapt key range based on after (for pagination), if present
	var afterQuery queryparse.Query
	if body["after"] != nil {
		// get after as map
		after, ok := body["after"].(map[string]interface{})
		if !ok {
			return httputil.NewError(400, "expected 'after' to be a json object")
		}

		// make query
		afterQuery, err = queryparse.JSONToQuery(after)
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("couldn't parse after query: %s", err.Error()))
		}
	}

	// set key range
	keyRange, err := stream.Codec.MakeKeyRange(whereQuery, afterQuery)
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// prepare write (we'll be writing as we get data, not in one batch)
	w.Header().Set("Content-Type", "application/json")

	// begin json object
	result := make([]interface{}, 0, 1)
	bytesRead := 0

	// read rows from engine
	err = db.Engine.Tables.ReadRecordRange(r.Context(), instanceID, keyRange, limit, func(avroData []byte, timestamp time.Time) error {
		// decode avro
		data, err := stream.Codec.UnmarshalAvro(avroData)
		if err != nil {
			return err
		}

		// convert to json friendly
		data, err = stream.Codec.ConvertFromAvroNative(data, true)
		if err != nil {
			return err
		}

		// set timestamp
		data["@meta"] = map[string]int64{"timestamp": timeutil.UnixMilli(timestamp)}

		// done
		result = append(result, data)
		bytesRead += len(avroData)
		return nil
	})
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// prepare result for encoding
	var encode interface{}
	if keyRange.CheckUnique() {
		if len(result) > 0 {
			encode = map[string]interface{}{"data": result[0]}
		} else {
			encode = map[string]interface{}{"data": nil}
		}
	} else {
		encode = map[string]interface{}{"data": result}
	}

	// write and finish
	w.Header().Set("Content-Type", "application/json")
	err = jsonutil.MarshalWriter(encode, w)
	if err != nil {
		return err
	}

	// track read metrics
	Metrics.TrackRead(instanceID, int64(len(result)), int64(bytesRead))
	if secret != nil {
		Metrics.TrackRead(secret.BillingID(), int64(len(result)), int64(bytesRead))
	}

	return nil
}

func getLatestFromInstanceID(w http.ResponseWriter, r *http.Request, instanceID uuid.UUID) error {
	// get auth
	secret := middleware.GetSecret(r.Context())

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(r.Context(), instanceID)
	if stream == nil {
		return httputil.NewError(404, "stream not found")
	}

	// check allowed to read stream
	if !stream.Public {
		perms := secret.StreamPermissions(r.Context(), stream.StreamID, stream.ProjectID, stream.External)
		if !perms.Read {
			return httputil.NewError(403, "secret doesn't grant right to read this stream")
		}
	}

	// check isn't batch
	if stream.Batch {
		return httputil.NewError(400, "cannot get latest records for batch streams")
	}

	// check quota
	if secret != nil {
		usage := Metrics.GetCurrentUsage(r.Context(), secret.BillingID())
		ok := secret.CheckReadQuota(usage)
		if !ok {
			return httputil.NewError(429, "you have exhausted your monthly quota")
		}
	}

	// read body
	var body map[string]interface{}
	err := jsonutil.Unmarshal(r.Body, &body)
	if err == io.EOF {
		// no body -- try reading from url parameters
		body = make(map[string]interface{})

		// read limit
		if limit := r.URL.Query().Get("limit"); limit != "" {
			body["limit"] = limit
		}

		// read before
		if before := r.URL.Query().Get("before"); before != "" {
			body["before"] = before
		}
	} else if err != nil {
		return httputil.NewError(400, "couldn't parse body -- is it valid JSON?")
	}

	// make sure there's no accidental keys
	for k := range body {
		if k != "limit" && k != "before" {
			return httputil.NewError(400, "unrecognized query key '%s'; valid keys are 'limit' and 'before'", k)
		}
	}

	// get limit
	limit, err := parseLimit(body["limit"])
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// get before
	before, err := timeutil.Parse(body["before"], true)
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// prepare write (we'll be writing as we get data, not in one batch)
	w.Header().Set("Content-Type", "application/json")

	// begin json object
	result := make([]interface{}, 0, limit)
	bytesRead := 0

	// read rows from engine
	err = db.Engine.Tables.ReadLatestRecords(r.Context(), instanceID, limit, before, func(avroData []byte, timestamp time.Time) error {
		// decode avro
		data, err := stream.Codec.UnmarshalAvro(avroData)
		if err != nil {
			return err
		}

		// convert to json friendly
		data, err = stream.Codec.ConvertFromAvroNative(data, true)
		if err != nil {
			return err
		}

		// set timestamp
		data["@meta"] = map[string]int64{"timestamp": timeutil.UnixMilli(timestamp)}

		// done
		result = append(result, data)
		bytesRead += len(avroData)
		return nil
	})
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// prepare result for encoding
	encode := map[string]interface{}{"data": result}

	// write and finish
	w.Header().Set("Content-Type", "application/json")
	err = jsonutil.MarshalWriter(encode, w)
	if err != nil {
		return err
	}

	// track read metrics
	Metrics.TrackRead(instanceID, int64(len(result)), int64(bytesRead))
	if secret != nil {
		Metrics.TrackRead(secret.BillingID(), int64(len(result)), int64(bytesRead))
	}

	return nil
}

func postToInstance(w http.ResponseWriter, r *http.Request) error {
	// get auth
	secret := middleware.GetSecret(r.Context())

	// get instance ID
	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
	if err != nil {
		return httputil.NewError(404, "instance not found -- malformed ID")
	}

	// get stream
	stream := entity.FindCachedStreamByCurrentInstanceID(r.Context(), instanceID)
	if stream == nil {
		return httputil.NewError(404, "stream not found")
	}

	// check allowed to write stream
	perms := secret.StreamPermissions(r.Context(), stream.StreamID, stream.ProjectID, stream.External)
	if !perms.Write {
		return httputil.NewError(403, "secret doesn't grant right to write this stream")
	}

	// check quota
	usage := Metrics.GetCurrentUsage(r.Context(), secret.BillingID())
	ok := secret.CheckWriteQuota(usage)
	if !ok {
		return httputil.NewError(429, "you have exhausted your monthly quota")
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
	bytesWritten := 0
	for idx, objV := range objects {
		// check it's a map
		obj, ok := objV.(map[string]interface{})
		if !ok {
			return httputil.NewError(400, fmt.Sprintf("record at index %d is not an object", idx))
		}

		// get timestamp
		timestamp := time.Now()
		meta, ok := obj["@meta"].(map[string]interface{})
		if ok {
			raw := meta["timestamp"]
			if raw != nil {
				raw, err := jsonutil.ParseInt64(meta["timestamp"])
				if err != nil {
					return httputil.NewError(400, "couldn't parse '@meta.timestamp' as number or numeric string for record at index %d", idx)
				}
				timestamp = timeutil.FromUnixMilli(raw)
			}
		}

		// convert to avro native for encoding
		avroNative, err := stream.Codec.ConvertToAvroNative(obj, true)
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// encode as avro
		avroData, err := stream.Codec.MarshalAvro(avroNative)
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// compute key (only used for size check)
		keyData, err := stream.Codec.MarshalKey(avroNative)
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// check sizes
		err = db.Engine.CheckSize(len(keyData), len(avroData))
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// save the record
		records[idx] = &pb.Record{
			AvroData:  avroData,
			Timestamp: timeutil.UnixMilli(timestamp),
		}

		// increment bytes written
		bytesWritten += len(avroData)
	}

	// queue write request (publishes to Pubsub)
	err = db.Engine.Streams.QueueWriteRequest(r.Context(), &pb.WriteRecordsRequest{
		InstanceId: instanceID.Bytes(),
		Records:    records,
	})
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// track write metrics
	Metrics.TrackWrite(instanceID, int64(len(records)), int64(bytesWritten))
	if secret != nil {
		Metrics.TrackWrite(secret.BillingID(), int64(len(records)), int64(bytesWritten))
	}

	// Done
	return nil
}

func parseLimit(val interface{}) (int, error) {
	limit := defaultRecordsLimit
	if val != nil {
		switch num := val.(type) {
		case string:
			l, err := strconv.Atoi(num)
			if err != nil {
				return 0, fmt.Errorf("couldn't parse limit as integer")
			}
			limit = l
		case json.Number:
			l, err := num.Int64()
			if err != nil {
				return 0, fmt.Errorf("couldn't parse limit as integer")
			}
			limit = int(l)
		default:
			return 0, fmt.Errorf("couldn't parse limit as integer")
		}
	}

	// check limit is valid
	if limit == 0 {
		return 0, fmt.Errorf("limit cannot be 0")
	} else if limit > maxRecordsLimit {
		return 0, fmt.Errorf("limit exceeds maximum of %d", maxRecordsLimit)
	}

	return limit, nil
}
