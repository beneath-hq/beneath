package rest

import (
	"net/http"

	"github.com/go-chi/chi"
)

// GetHandler returns a HTTP handler
func GetHandler() http.Handler {
	handler := chi.NewRouter()

	// TODO: Add graphql
	// GraphQL endpoints
	// handler.Get("/graphql")
	// handler.Get("/projects/{projectName}/graphql")

	// REST endpoints
	handler.Get("/projects/{projectName}/streams/{streamName}", getStreamHandler)
	// handler.Post("/projects/{projectName}/streams/{streamName}")

	return handler
}

func getStreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}
