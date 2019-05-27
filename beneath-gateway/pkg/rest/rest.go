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
	handler.Get("/projects/{projectName}/streams/{streamName}", getStream)
	handler.Post("/projects/{projectName}/streams/{streamName}", postStream)

	return handler
}

func getStream(w http.ResponseWriter, r *http.Request) {
	// projectName := chi.URLParam(r, "projectName")
	// streamName := chi.URLParam(r, "streamName")

	// SPEC
	// Read Authorization, check token and get user
	// Check user has read access to stream (or stream is public if not user)
	// (Read from BT in accordance with how we end up writing it)

	w.Write([]byte("hello world"))
}

func postStream(w http.ResponseWriter, r *http.Request) {
	// projectName := chi.URLParam(r, "projectName")
	// streamName := chi.URLParam(r, "streamName")

	// SPEC
	// - Read Authorization, check token and get user
	// - Check user has write access to project
	// - Get schema for stream
	// - Read payload (JSON) and encode with schema
	// - Write to pubsub
}
