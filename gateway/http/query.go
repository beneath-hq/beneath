package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/gateway/api"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
)

type queryArgsType string

const (
	queryArgsTypeIndex queryArgsType = "index"
	queryArgsTypeLog                 = "log"
)

type queryArgs struct {
	Limit  int
	Type   queryArgsType
	Filter string
	Peek   bool
}

func getFromOrganizationAndProjectAndStream(w http.ResponseWriter, r *http.Request) error {
	// if cursor is set, go straight to fetching from cursor (the instance ID is backed into the cursor)
	if chi.URLParam(r, "cursor") != "" {
		return getFromCursor(w, r)
	}

	// parse url params
	organizationName := toBackendName(chi.URLParam(r, "organizationName"))
	projectName := toBackendName(chi.URLParam(r, "projectName"))
	streamName := toBackendName(chi.URLParam(r, "streamName"))

	// find primary instance ID for stream
	instanceID := entity.FindInstanceIDByOrganizationProjectAndName(r.Context(), organizationName, projectName, streamName)
	if instanceID == uuid.Nil {
		return httputil.NewError(404, "instance for stream not found")
	}

	// run query
	return handleQuery(w, r, instanceID)
}

func getFromInstance(w http.ResponseWriter, r *http.Request) error {
	// if cursor is set, go straight to fetching from cursor (the instance ID is backed into the cursor)
	if chi.URLParam(r, "cursor") != "" {
		return getFromCursor(w, r)
	}

	// parse instance ID
	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
	if err != nil {
		return httputil.NewError(404, "instance not found -- malformed ID")
	}

	// run query
	return handleQuery(w, r, instanceID)
}

func handleQuery(w http.ResponseWriter, r *http.Request, instanceID uuid.UUID) error {
	// read body
	args, err := parseQueryArgs(r)
	if err != nil {
		return err
	}

	// create cursors
	var replayCursors, changeCursors [][]byte
	switch args.Type {
	case queryArgsTypeIndex:
		res, err := api.HandleQueryIndex(r.Context(), &api.QueryIndexRequest{
			InstanceID: instanceID,
			Partitions: 1,
			Filter:     args.Filter,
		})
		if err != nil {
			return err.HTTP()
		}
		replayCursors = res.ReplayCursors
		changeCursors = res.ChangeCursors
	case queryArgsTypeLog:
		res, err := api.HandleQueryLog(r.Context(), &api.QueryLogRequest{
			InstanceID: instanceID,
			Partitions: 1,
			Peek:       args.Peek,
		})
		if err != nil {
			return err.HTTP()
		}
		replayCursors = res.ReplayCursors
		changeCursors = res.ChangeCursors
	}

	// extract single cursors (there should be 0 or 1)
	var replayCursor, changeCursor []byte
	if len(replayCursors) != 0 {
		replayCursor = replayCursors[0]
	}
	if len(changeCursors) != 0 {
		changeCursor = changeCursors[0]
	}

	// run a read
	return handleRead(w, r, replayCursor, args.Limit, changeCursor)
}

func parseQueryArgs(r *http.Request) (queryArgs, error) {
	args := queryArgs{}

	// read limit
	limit, err := parseIntParam("limit", chi.URLParam(r, "limit"))
	if err != nil {
		return args, httputil.NewError(http.StatusBadRequest, err.Error())
	}
	args.Limit = limit

	// read type
	t := chi.URLParam(r, "type")
	if t == "" || t == string(queryArgsTypeIndex) {
		args.Type = queryArgsTypeIndex
	} else if t == string(queryArgsTypeLog) {
		args.Type = queryArgsTypeLog
	} else {
		return args, httputil.NewError(400, fmt.Sprintf("unrecognized query type '%s'; options are '%s' (default) and '%s'", args.Type, queryArgsTypeIndex, queryArgsTypeLog))
	}

	// read filter
	args.Filter = chi.URLParam(r, "filter")

	// read peek
	args.Peek, err = parseBoolParam("peek", chi.URLParam(r, "peek"))
	if err != nil {
		return args, httputil.NewError(http.StatusBadRequest, err.Error())
	}

	// check index-only params
	if args.Type == queryArgsTypeLog && args.Filter != "" {
		return args, httputil.NewError(400, "you cannot provide the 'filter' parameter for a 'log' query")
	}

	// check log-only params
	if args.Type == queryArgsTypeIndex && args.Peek {
		return args, httputil.NewError(400, "you cannot provide the 'peek' parameter for a 'query' query")
	}

	return args, nil
}
