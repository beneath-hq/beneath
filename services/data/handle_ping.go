package data

import (
	"context"
	"net/http"

	version "github.com/hashicorp/go-version"

	"gitlab.com/beneath-hq/beneath/services/middleware"
	"gitlab.com/beneath-hq/beneath/services/data/clientversion"
)

// PingRequest is a request to HandlePing
type PingRequest struct {
	ClientID      string
	ClientVersion string
}

// PingResponse is a result from HandlePing
type PingResponse struct {
	Authenticated      bool
	VersionStatus      string
	RecommendedVersion string
}

type pingTags struct {
	ClientID      string `json:"client_id,omitempty"`
	ClientVersion string `json:"client_version,omitempty"`
}

// HandlePing handles a ping request from a client library
func (s *Service) HandlePing(ctx context.Context, req *PingRequest) (*PingResponse, *Error) {
	middleware.SetTagsPayload(ctx, pingTags{
		ClientID:      req.ClientID,
		ClientVersion: req.ClientVersion,
	})

	spec := clientversion.Specs[req.ClientID]
	if spec.IsZero() {
		return nil, newError(http.StatusBadRequest, "unrecognized client ID")
	}

	v, err := version.NewVersion(req.ClientVersion)
	if err != nil {
		return nil, newError(http.StatusBadRequest, "client version is not a valid semver")
	}

	status := ""
	if v.GreaterThanOrEqual(spec.RecommendedVersion) {
		status = "stable"
	} else if v.GreaterThanOrEqual(spec.WarningVersion) {
		status = "warning"
	} else {
		status = "deprecated"
	}

	secret := middleware.GetSecret(ctx)
	return &PingResponse{
		Authenticated:      !secret.IsAnonymous(),
		VersionStatus:      status,
		RecommendedVersion: spec.RecommendedVersion.String(),
	}, nil
}
