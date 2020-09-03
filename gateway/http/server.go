package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/rs/cors"

	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/pkg/ws"
)

// Handler serves the gateway HTTP API
func Handler() http.Handler {
	handler := chi.NewRouter()

	handler.Use(chimiddleware.RealIP)
	handler.Use(chimiddleware.DefaultCompress)
	handler.Use(middleware.InjectTags)
	handler.Use(middleware.Logger)
	handler.Use(middleware.Recoverer)
	handler.Use(middleware.Auth)
	handler.Use(middleware.IPRateLimit())

	// Add CORS
	handler.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           7200,
		Debug:            false,
	}).Handler)

	// Add health check
	handler.Get("/", healthCheck)
	handler.Get("/healthz", healthCheck)

	// Add ping
	handler.Method("GET", "/v1/-/ping", httputil.AppHandler(getPing))

	// index and log endpoints
	handler.Method("GET", "/v1/{organizationName}/{projectName}/{streamName}", httputil.AppHandler(getFromOrganizationAndProjectAndStream))
	handler.Method("GET", "/v1/-/instances/{instanceID}", httputil.AppHandler(getFromInstance))

	// write endpoint
	handler.Method("POST", "/v1/{organizationName}/{projectName}/{streamName}", httputil.AppHandler(postToOrganizationAndProjectAndStream))
	handler.Method("POST", "/v1/-/instances/{instanceID}", httputil.AppHandler(postToInstance))

	// warehouse job endpoints
	handler.Method("GET", "/v1/-/warehouse/{jobID}", httputil.AppHandler(getFromWarehouseJob))
	handler.Method("POST", "/v1/-/warehouse", httputil.AppHandler(postToWarehouseJob))

	// read endpoint
	handler.Method("GET", "/v1/-/cursor", httputil.AppHandler(getFromCursor))

	// create websocket broker and start accepting new connections on /ws
	wss := ws.NewBroker(&wsServer{})
	handler.Method("GET", "/v1/-/ws", httputil.AppHandler(wss.HTTPHandler))

	return handler
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if hub.Healthy() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	} else {
		log.S.Errorf("Gateway database health check failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func parseBoolParam(name string, val string) (bool, error) {
	if val == "" {
		return false, nil
	} else if val == "true" {
		return true, nil
	} else if val == "false" {
		return false, nil
	}

	return false, fmt.Errorf("expected '%s' parameter to be 'true' or 'false'", name)
}

func parseIntParam(name string, val string) (int, error) {
	if val == "" {
		return 0, nil
	}

	res, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("couldn't parse '%s' as integer", name)
	}

	return res, nil
}

func toBackendName(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "-", "_"))
}
