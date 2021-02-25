package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/services/data"
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

func (a *app) getFromOrganizationAndProjectAndStream(w http.ResponseWriter, r *http.Request) error {
	// if cursor is set, go straight to fetching from cursor (the instance ID is backed into the cursor)
	if r.URL.Query().Get("cursor") != "" {
		return a.getFromCursor(w, r)
	}

	// parse url params
	organizationName := toBackendName(chi.URLParam(r, "organizationName"))
	projectName := toBackendName(chi.URLParam(r, "projectName"))
	streamName := toBackendName(chi.URLParam(r, "streamName"))

	// find primary instance ID for stream
	instanceID := a.StreamService.FindPrimaryInstanceIDByOrganizationProjectAndName(r.Context(), organizationName, projectName, streamName)
	if instanceID == uuid.Nil {
		return httputil.NewError(404, "instance for stream not found")
	}

	// run query
	return a.handleQuery(w, r, instanceID)
}

func (a *app) getFromInstance(w http.ResponseWriter, r *http.Request) error {
	// if cursor is set, go straight to fetching from cursor (the instance ID is backed into the cursor)
	if r.URL.Query().Get("cursor") != "" {
		return a.getFromCursor(w, r)
	}

	// parse instance ID
	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
	if err != nil {
		return httputil.NewError(404, "instance not found -- malformed ID")
	}

	// run query
	return a.handleQuery(w, r, instanceID)
}

func (a *app) handleQuery(w http.ResponseWriter, r *http.Request, instanceID uuid.UUID) error {
	// read body
	args, err := parseQueryArgs(r)
	if err != nil {
		return err
	}

	// create cursors
	var replayCursors, changeCursors [][]byte
	switch args.Type {
	case queryArgsTypeIndex:
		res, err := a.DataService.HandleQueryIndex(r.Context(), &data.QueryIndexRequest{
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
		res, err := a.DataService.HandleQueryLog(r.Context(), &data.QueryLogRequest{
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
	if replayCursor != nil {
		return a.handleRead(w, r, replayCursor, args.Limit, changeCursor)
	}
	return a.writeReadResponse(w, instanceID, uuid.Nil, replayCursor, changeCursor, nil)
}

func parseQueryArgs(r *http.Request) (queryArgs, error) {
	args := queryArgs{}

	// read limit
	limit, err := parseIntParam("limit", r.URL.Query().Get("limit"))
	if err != nil {
		return args, httputil.NewError(http.StatusBadRequest, err.Error())
	}
	args.Limit = limit

	// read type
	t := r.URL.Query().Get("type")
	if t == "" || t == string(queryArgsTypeIndex) {
		args.Type = queryArgsTypeIndex
	} else if t == string(queryArgsTypeLog) {
		args.Type = queryArgsTypeLog
	} else {
		return args, httputil.NewError(400, fmt.Sprintf("unrecognized query type '%s'; options are '%s' (default) and '%s'", args.Type, queryArgsTypeIndex, queryArgsTypeLog))
	}

	// read filter
	args.Filter = r.URL.Query().Get("filter")

	// read peek
	args.Peek, err = parseBoolParam("peek", r.URL.Query().Get("peek"))
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
