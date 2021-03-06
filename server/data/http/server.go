package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"go.uber.org/zap"

	"github.com/beneath-hq/beneath/pkg/httputil"
	"github.com/beneath-hq/beneath/pkg/ws"
	"github.com/beneath-hq/beneath/services/data"
	"github.com/beneath-hq/beneath/services/middleware"
	"github.com/beneath-hq/beneath/services/secret"
	"github.com/beneath-hq/beneath/services/table"
)

type app struct {
	Logger        *zap.SugaredLogger
	DataService   *data.Service
	SecretService *secret.Service
	TableService  *table.Service
}

// NewServer creates and returns the data HTTP server
func NewServer(logger *zap.Logger, data *data.Service, middleware *middleware.Service, secret *secret.Service, table *table.Service) *http.Server {
	l := logger.Named("http")
	app := &app{
		Logger:        l.Sugar(),
		DataService:   data,
		SecretService: secret,
		TableService:  table,
	}
	router := chi.NewRouter()

	corsOptions := cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           7200,
		Debug:            false,
	}

	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.DefaultCompress)
	router.Use(cors.New(corsOptions).Handler)
	router.Use(middleware.InjectTagsMiddleware)
	router.Use(middleware.LoggerMiddleware(l))
	router.Use(middleware.RecovererMiddleware(l))
	router.Use(middleware.AuthMiddleware)
	router.Use(middleware.IPRateLimitMiddleware)

	// Add health check
	router.Get("/", app.healthCheck)
	router.Get("/healthz", app.healthCheck)

	// Add ping
	router.Method("GET", "/v1/-/ping", httputil.AppHandler(app.getPing))

	// index and log endpoints
	router.Method("GET", "/v1/{organizationName}/{projectName}/{tableName}", httputil.AppHandler(app.getFromOrganizationAndProjectAndTable))
	router.Method("GET", "/v1/-/instances/{instanceID}", httputil.AppHandler(app.getFromInstance))

	// write endpoint
	router.Method("POST", "/v1/{organizationName}/{projectName}/{tableName}", httputil.AppHandler(app.postToOrganizationAndProjectAndTable))
	router.Method("POST", "/v1/-/instances/{instanceID}", httputil.AppHandler(app.postToInstance))

	// warehouse job endpoints
	router.Method("GET", "/v1/-/warehouse/{jobID}", httputil.AppHandler(app.getFromWarehouseJob))
	router.Method("POST", "/v1/-/warehouse", httputil.AppHandler(app.postToWarehouseJob))

	// read endpoint
	router.Method("GET", "/v1/-/cursor", httputil.AppHandler(app.getFromCursor))

	// create websocket broker and start accepting new connections on /ws
	wss := ws.NewBroker(app)
	router.Method("GET", "/v1/-/ws", httputil.AppHandler(wss.HTTPHandler))

	return &http.Server{
		Handler: router,
	}
}

func (a *app) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
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
