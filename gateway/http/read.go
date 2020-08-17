package http

import (
	"net/http"

	"github.com/mr-tron/base58"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/gateway/api"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/jsonutil"
)

type readArgs struct {
	Cursor []byte
	Limit  int
}

func getFromCursor(w http.ResponseWriter, r *http.Request) error {
	args, err := parseReadArgs(r)
	if err != nil {
		return err
	}

	return handleRead(w, r, args.Cursor, args.Limit, nil)
}

func handleRead(w http.ResponseWriter, r *http.Request, cursor []byte, limit int, passChangeCursor []byte) error {
	res, errr := api.HandleRead(r.Context(), &api.ReadRequest{
		Cursor:     cursor,
		Limit:      int32(limit),
		ReturnJSON: true,
	})
	if errr != nil {
		return errr.HTTP()
	}

	// make meta
	meta := make(map[string]interface{})
	if res.InstanceID != uuid.Nil {
		meta["instance_id"] = res.InstanceID
	} else if res.JobID != uuid.Nil {
		meta["job_id"] = res.JobID
	}
	if res.NextCursor != nil {
		meta["next_cursor"] = base58.Encode(res.NextCursor)
	}
	if passChangeCursor != nil {
		meta["change_cursor"] = base58.Encode(passChangeCursor)
	}

	// prepare result for encoding
	encode := map[string]interface{}{
		"meta": meta,
		"data": res.JSON,
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
