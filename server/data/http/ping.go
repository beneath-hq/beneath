package http

import (
	"net/http"

	"github.com/beneath-hq/beneath/pkg/jsonutil"
	"github.com/beneath-hq/beneath/services/data"
)

type pingArgs struct {
	ClientID      string
	ClientVersion string
}

func (a *app) getPing(w http.ResponseWriter, r *http.Request) error {
	// parse args
	args := pingArgs{}
	args.ClientID = r.URL.Query().Get("client_id")
	args.ClientVersion = r.URL.Query().Get("client_version")

	// handle
	res, errr := a.DataService.HandlePing(r.Context(), &data.PingRequest{
		ClientID:      args.ClientID,
		ClientVersion: args.ClientVersion,
	})
	if errr != nil {
		return errr.HTTP()
	}

	// prepare result for encoding
	encode := map[string]interface{}{
		"data": map[string]interface{}{
			"authenticated":       res.Authenticated,
			"version_status":      res.VersionStatus,
			"recommended_version": res.RecommendedVersion,
		},
	}

	// write and finish
	w.Header().Set("Content-Type", "application/json")
	err := jsonutil.MarshalWriter(encode, w)
	if err != nil {
		return err
	}

	return nil
}
