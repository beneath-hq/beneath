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

	"github.com/beneath-core/beneath-go/core/httputil"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/core/middleware"
	"github.com/beneath-core/beneath-go/core/ws"
	"github.com/beneath-core/beneath-go/db"
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
	handler.Method("GET", "/ws", httputil.AppHandler(wss.HTTPHandler))

	// write endpoint
	handler.Method("POST", "/streams/instances/{instanceID}", httputil.AppHandler(postToInstance))

	// query endpoints
	handler.Method("GET", "/projects/{projectName}/streams/{streamName}", httputil.AppHandler(getFromProjectAndStream))
	handler.Method("GET", "/streams/instances/{instanceID}", httputil.AppHandler(getFromInstance))

	// peek endpoints
	// handler.Method("GET", "/projects/{projectName}/streams/{streamName}/latest", httputil.AppHandler(getLatestFromProjectAndStream))
	// handler.Method("GET", "/streams/instances/{instanceID}/latest", httputil.AppHandler(getLatestFromInstance))

	// meta endpoint
	handler.Method("GET", "/projects/{projectName}/streams/{streamName}/details", httputil.AppHandler(getStreamDetails))

	return handler
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if db.Healthy() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	} else {
		log.S.Errorf("Gateway database health check failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

type writeTags struct {
	InstanceID   string `json:"instance_id,omitempty"`
	RecordsCount int    `json:"records,omitempty"`
	BytesWritten int    `json:"bytes,omitempty"`
}

type queryTags struct {
	InstanceID string `json:"instance_id,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     []byte `json:"cursor,omitempty"`
	Filter     string `json:"filter,omitempty"`
	Compact    bool   `json:"compact,omitempty"`
	Partitions int32  `json:"partitions,omitempty"`
	BytesRead  int    `json:"bytes,omitempty"`
}

type peekTags struct {
	InstanceID string `json:"instance_id,omitempty"`
	Limit      int32  `json:"limit,omitempty"`
}

type streamDetailsTags struct {
	Stream  string `json:"stream"`
	Project string `json:"project"`
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
