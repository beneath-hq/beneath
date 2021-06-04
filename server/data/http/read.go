package http

import (
	"net/http"

	"github.com/mr-tron/base58"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/pkg/httputil"
	"github.com/beneath-hq/beneath/pkg/jsonutil"
	"github.com/beneath-hq/beneath/services/data"
)

type readArgs struct {
	Cursor []byte
	Limit  int
}

func (a *app) getFromCursor(w http.ResponseWriter, r *http.Request) error {
	args, err := parseReadArgs(r)
	if err != nil {
		return err
	}

	return a.handleRead(w, r, args.Cursor, args.Limit, nil)
}

func (a *app) handleRead(w http.ResponseWriter, r *http.Request, cursor []byte, limit int, passChangeCursor []byte) error {
	res, errr := a.DataService.HandleRead(r.Context(), &data.ReadRequest{
		Cursor:     cursor,
		Limit:      int32(limit),
		ReturnJSON: true,
	})
	if errr != nil {
		return errr.HTTP()
	}

	return a.writeReadResponse(w, res.InstanceID, res.JobID, res.NextCursor, passChangeCursor, res.JSON)
}

func (a *app) writeReadResponse(w http.ResponseWriter, instanceID uuid.UUID, jobID uuid.UUID, nextCursor []byte, changeCursor []byte, data interface{}) error {
	// make meta
	meta := make(map[string]interface{})
	if instanceID != uuid.Nil {
		meta["instance_id"] = instanceID
	} else if jobID != uuid.Nil {
		meta["job_id"] = jobID
	}
	if nextCursor != nil {
		meta["next_cursor"] = base58.Encode(nextCursor)
	}
	if changeCursor != nil {
		meta["change_cursor"] = base58.Encode(changeCursor)
	}

	// prepare result for encoding
	encode := map[string]interface{}{
		"meta": meta,
		"data": data,
	}

	// write and finish
	w.Header().Set("Content-Type", "application/json")
	err := jsonutil.MarshalWriter(encode, w)
	if err != nil {
		return err
	}

	return nil
}

func parseReadArgs(r *http.Request) (readArgs, error) {
	args := readArgs{}

	limit, err := parseIntParam("limit", r.URL.Query().Get("limit"))
	if err != nil {
		return args, httputil.NewError(http.StatusBadRequest, err.Error())
	}
	args.Limit = limit

	cursor := r.URL.Query().Get("cursor")
	if cursor != "" {
		args.Cursor, err = base58.Decode(cursor)
		if err != nil {
			return args, httputil.NewError(http.StatusBadRequest, "invalid cursor")
		}
	}

	if args.Cursor == nil {
		return args, httputil.NewError(http.StatusBadRequest, "expected url argument 'cursor'")
	}

	if r.URL.Query().Get("type") != "" || r.URL.Query().Get("filter") != "" || r.URL.Query().Get("peek") != "" {
		return args, httputil.NewError(400, "you cannot provide query parameters ('type', 'filter' or 'peek') along with a 'cursor'")
	}

	return args, nil
}
