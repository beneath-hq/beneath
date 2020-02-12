package http

import (
	"net/http"

	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/core/httputil"
	"github.com/beneath-core/core/jsonutil"
	"github.com/beneath-core/core/middleware"
)

func getStreamDetails(w http.ResponseWriter, r *http.Request) error {
	// get instance ID
	projectName := toBackendName(chi.URLParam(r, "projectName"))
	streamName := toBackendName(chi.URLParam(r, "streamName"))
	instanceID := entity.FindInstanceIDByNameAndProject(r.Context(), streamName, projectName)
	if instanceID == uuid.Nil {
		return httputil.NewError(404, "instance for stream not found")
	}

	// set log payload
	middleware.SetTagsPayload(r.Context(), streamDetailsTags{
		Stream:  streamName,
		Project: projectName,
	})

	// get stream details
	stream := entity.FindCachedStreamByCurrentInstanceID(r.Context(), instanceID)
	if stream == nil {
		return httputil.NewError(404, "stream not found")
	}

	// check allowed to read stream
	secret := middleware.GetSecret(r.Context())
	perms := secret.StreamPermissions(r.Context(), stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Read {
		return httputil.NewError(403, "secret doesn't grant right to read this stream")
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
		"primary_index":       stream.Codec.PrimaryIndex,
		"secondary_indexes":   stream.Codec.SecondaryIndexes,
		"avro_schema":         stream.Codec.AvroSchemaString,
	})
	if err != nil {
		return httputil.NewError(500, err.Error())
	}

	// write
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
	return nil
}
