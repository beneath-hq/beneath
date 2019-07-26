package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
