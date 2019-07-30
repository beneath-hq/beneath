package gateway

import (
	"fmt"
	"log"
	"net/http"

	"github.com/beneath-core/beneath-go/control/auth"
	"github.com/beneath-core/beneath-go/core/httputil"
	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/gateway/websockets"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
)

// ListenAndServeWS serves a WebSocket API
func ListenAndServeWS(port int) error {
	log.Printf("Websockets server running on port %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), wsHandler())
}

func wsHandler() http.Handler {
	// prepare router
	handler := chi.NewRouter()

	// standard middleware
	handler.Use(middleware.Logger)
	handler.Use(middleware.Recoverer)

	// Add CORS
	handler.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	// check auth
	handler.Use(auth.HTTPMiddleware)

	// create broker and run in background
	broker := websockets.NewBroker(db.Engine)

	// accept new websockets on /ws
	handler.Handle("/ws", httputil.AppHandler(broker.HTTPHandler))

	return handler
}
