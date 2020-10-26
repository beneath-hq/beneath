package http

import (
	"net/http"

	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/jsonutil"
	"gitlab.com/beneath-hq/beneath/services/data"
)

func (a *app) postToOrganizationAndProjectAndStream(w http.ResponseWriter, r *http.Request) error {
	organizationName := toBackendName(chi.URLParam(r, "organizationName"))
	projectName := toBackendName(chi.URLParam(r, "projectName"))
	streamName := toBackendName(chi.URLParam(r, "streamName"))

	instanceID := a.StreamService.FindPrimaryInstanceIDByOrganizationProjectAndName(r.Context(), organizationName, projectName, streamName)
	if instanceID == uuid.Nil {
		return httputil.NewError(404, "instance for stream not found")
	}

	return a.handleWrite(w, r, instanceID)
}

func (a *app) postToInstance(w http.ResponseWriter, r *http.Request) error {
	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
	if err != nil {
		return httputil.NewError(404, "instance not found -- malformed ID")
	}

	return a.handleWrite(w, r, instanceID)
}

func (a *app) handleWrite(w http.ResponseWriter, r *http.Request, instanceID uuid.UUID) error {
	// decode json body
	var body interface{}
	err := jsonutil.Unmarshal(r.Body, &body)
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
		return httputil.NewError(400, "request body must be an object or an array of objects")
	}

	// call write
	res, errr := a.DataService.HandleWrite(r.Context(), &data.WriteRequest{
		InstanceRecords: map[uuid.UUID]*data.WriteRecords{
			instanceID: &data.WriteRecords{JSON: objects},
		},
	})
	if errr != nil {
		return errr.HTTP()
	}

	// result
	encode := map[string]interface{}{
		"meta": map[string]string{
			"write_id": res.WriteID.String(),
		},
	}

	// write and finish
	w.Header().Set("Content-Type", "application/json")
	err = jsonutil.MarshalWriter(encode, w)
	if err != nil {
		return err
	}

	// Done
	return nil
}
