package gateway

import (
	"encoding/json"
	"net/http"
)

// HTTPError represents an error with a HTTP status code
type HTTPError struct {
	Code    int
	Message string
}

// NewHTTPError creates a new HTTPError
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{code, message}
}

func (e *HTTPError) Error() string {
	return e.Message
}

// WriteHTTPError writes a JSON-formatted error
func WriteHTTPError(w http.ResponseWriter, err error) {
	// get code
	code := 400
	if httperr, ok := err.(*HTTPError); ok {
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
		WriteHTTPError(w, err)
	}
}
