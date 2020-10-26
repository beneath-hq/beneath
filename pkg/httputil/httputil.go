package httputil

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

// ListenAndServeContext serves a HTTP server and performs a graceful shutdown
// if/when ctx is cancelled.
func ListenAndServeContext(ctx context.Context, server *http.Server, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	var serveErr error
	go func() {
		serveErr = server.Serve(lis)
		cancel()
	}()

	<-cctx.Done()
	if serveErr == nil {
		// server.Serve always returns a non-nil err, so this must be a cancel on the parent ctx.
		// We perform a graceful shutdown.
		serveErr = server.Shutdown(context.Background())
	}

	return serveErr
}

// Error represents an error with a HTTP status code
type Error struct {
	Code    int
	Message string
}

// NewError creates a new HTTP error
func NewError(code int, format string, args ...interface{}) *Error {
	return &Error{code, fmt.Sprintf(format, args...)}
}

func (e *Error) Error() string {
	return e.Message
}

// WriteError writes a JSON-formatted error
func WriteError(w http.ResponseWriter, err error) {
	// get code
	code := 400
	if httperr, ok := err.(*Error); ok {
		code = httperr.Code
	}

	// set json headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// write header with status code
	w.WriteHeader(code)

	// write json error
	obj := map[string]string{"error": err.Error()}
	json, _ := json.Marshal(obj)
	w.Write(json)
}

// AppHandler packages HTTP handlers with error handling
type AppHandler func(http.ResponseWriter, *http.Request) error

func (fn AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err != nil {
		WriteError(w, err)
	}
}
