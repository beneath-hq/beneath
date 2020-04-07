package http

import (
	"encoding/json"
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

const (
	defaultReadLimit = 50
	maxReadLimit     = 1000
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

	// create websocket broker and start accepting new connections on /ws
	wss := ws.NewBroker(&wsServer{})
	handler.Method("GET", "/v1/-/ws", httputil.AppHandler(wss.HTTPHandler))

	// write endpoint
	handler.Method("POST", "/v1/-/instances/{instanceID}", httputil.AppHandler(postToInstance))

	// query endpoints
	handler.Method("GET", "/v1/{organizationName}/{projectName}/streams/{streamName}", httputil.AppHandler(getFromOrganizationAndProjectAndStream))
	handler.Method("GET", "/v1/-/instances/{instanceID}", httputil.AppHandler(getFromInstance))

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

func parseLimit(val interface{}) (int, error) {
	limit := defaultReadLimit
	if val != nil {
		switch num := val.(type) {
		case string:
			l, err := strconv.Atoi(num)
			if err != nil {
				return 0, fmt.Errorf("couldn't parse limit as integer")
			}
			limit = l
		case json.Number:
			l, err := num.Int64()
			if err != nil {
				return 0, fmt.Errorf("couldn't parse limit as integer")
			}
			limit = int(l)
		default:
			return 0, fmt.Errorf("couldn't parse limit as integer")
		}
	}

	// check limit is valid
	if limit == 0 {
		return 0, fmt.Errorf("limit cannot be 0")
	} else if limit > maxReadLimit {
		return 0, fmt.Errorf("limit exceeds maximum of %d", maxReadLimit)
	}

	return limit, nil
}

func toBackendName(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "-", "_"))
}
