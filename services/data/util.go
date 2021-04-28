package data

import (
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-hq/beneath/pkg/httputil"
)

// Error represents a generalized GRPC and HTTP error
type Error struct {
	httpStatus int
	msg        string
}

func newError(httpStatus int, msg string) *Error {
	return &Error{httpStatus: httpStatus, msg: msg}
}

func newErrorf(httpStatus int, format string, a ...interface{}) *Error {
	return &Error{httpStatus: httpStatus, msg: fmt.Sprintf(format, a...)}
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.msg
}

// GRPC returns a GRPC-friendly error
func (e *Error) GRPC() error {
	var code codes.Code
	switch e.httpStatus {
	case http.StatusForbidden:
		code = codes.PermissionDenied
	case http.StatusTooManyRequests:
		code = codes.ResourceExhausted
	default:
		code = codes.InvalidArgument
	}
	return status.Error(code, e.msg)
}

// HTTP returns a HTTP-friendly error
func (e *Error) HTTP() error {
	return httputil.NewError(e.httpStatus, e.msg)
}
