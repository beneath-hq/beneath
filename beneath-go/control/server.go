package control

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vektah/gqlparser/ast"

	"github.com/beneath-core/beneath-go/core/log"

	"github.com/beneath-core/beneath-go/control/auth"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/migrations"
	"github.com/beneath-core/beneath-go/control/resolver"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/middleware"
	"github.com/beneath-core/beneath-go/core/segment"
	"github.com/beneath-core/beneath-go/db"
	analytics "gopkg.in/segmentio/analytics-go.v3"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/gqlerror"
)

type configSpecification struct {
	ControlPort  int    `envconfig:"CONTROL_PORT" required:"true" default:"8080"`
	ControlHost  string `envconfig:"CONTROL_HOST" required:"true"`
	FrontendHost string `envconfig:"FRONTEND_HOST" required:"true"`

	RedisURL         string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresUser     string `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`

	StreamsDriver   string `envconfig:"ENGINE_STREAMS_DRIVER" required:"true"`
	TablesDriver    string `envconfig:"ENGINE_TABLES_DRIVER" required:"true"`
	WarehouseDriver string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`

	SessionSecret    string `envconfig:"CONTROL_SESSION_SECRET" required:"true"`
	GithubAuthID     string `envconfig:"CONTROL_GITHUB_AUTH_ID" required:"true"`
	GithubAuthSecret string `envconfig:"CONTROL_GITHUB_AUTH_SECRET" required:"true"`
	GoogleAuthID     string `envconfig:"CONTROL_GOOGLE_AUTH_ID" required:"true"`
	GoogleAuthSecret string `envconfig:"CONTROL_GOOGLE_AUTH_SECRET" required:"true"`

	SegmentServersideKey string `envconfig:"SEGMENT_SERVERSIDE_KEY" required:"true"`
}

var (
	// Config for control
	Config configSpecification
)

func init() {
	// load config
	core.LoadConfig("beneath", &Config)

	// connect postgres, redis and engine
	db.InitPostgres(Config.PostgresHost, Config.PostgresUser, Config.PostgresPassword)
	db.InitRedis(Config.RedisURL)
	db.InitEngine(Config.StreamsDriver, Config.TablesDriver, Config.WarehouseDriver)

	// run migrations
	migrations.MustRunUp(db.DB)

	// configure auth
	auth.InitGoth(&auth.GothConfig{
		ClientHost:       Config.FrontendHost,
		SessionSecret:    Config.SessionSecret,
		BackendHost:      Config.ControlHost,
		GithubAuthID:     Config.GithubAuthID,
		GithubAuthSecret: Config.GithubAuthSecret,
		GoogleAuthID:     Config.GoogleAuthID,
		GoogleAuthSecret: Config.GoogleAuthSecret,
	})

	// init segment
	segment.InitClient(Config.SegmentServersideKey)
}

// ListenAndServeHTTP serves the GraphQL API on HTTP
func ListenAndServeHTTP(port int) error {
	router := chi.NewRouter()

	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.DefaultCompress)
	router.Use(middleware.InjectTags)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Auth)
	router.Use(middleware.IPRateLimit())
	router.Use(SegmentMiddleware)

	// Add CORS
	router.Use(cors.New(cors.Options{
		AllowedOrigins: []string{
			Config.FrontendHost,
			Config.ControlHost,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	// Authentication endpoints
	router.Mount("/auth", auth.Router())

	// Add health check
	router.Get("/", healthCheck)
	router.Get("/healthz", healthCheck)

	// Add playground
	router.Handle("/playground", handler.Playground("Beneath", "/graphql"))

	// Add graphql server
	router.Handle("/graphql", handler.GraphQL(
		makeExecutableSchema(),
		makeQueryLoggingMiddleware(),
		makeGraphQLErrorPresenter(),
		handler.RecoverFunc(func(ctx context.Context, err interface{}) error {
			panic(err)
		}),
	))

	// Serve
	log.S.Infow("control http started", "port", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if db.Healthy() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	} else {
		log.S.Errorf("Control database health check failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func makeExecutableSchema() graphql.ExecutableSchema {
	return gql.NewExecutableSchema(gql.Config{Resolvers: &resolver.Resolver{}})
}

func makeGraphQLErrorPresenter() handler.Option {
	return handler.ErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		tags := middleware.GetTags(ctx)
		if q, ok := tags.Query.(map[string]interface{}); ok {
			q["error"] = err
		}
		// Uncomment this line to print resolver error details in the console
		// fmt.Printf("Error in GraphQL Resolver: %s", err.Error())
		return graphql.DefaultErrorPresenter(ctx, err)
	})
}

func makeQueryLoggingMiddleware() handler.Option {
	return handler.RequestMiddleware(func(ctx context.Context, next func(ctx context.Context) []byte) []byte {
		tags := middleware.GetTags(ctx)
		reqCtx := graphql.GetRequestContext(ctx)
		tags.Query = logInfoFromRequestContext(reqCtx)
		// The following also includes variables, but not good for right to be forgotten (GDPR)
		// tags.Query = map[string]interface{}{
		// 	"q":    logInfoFromRequestContext(reqCtx),
		// 	"vars": reqCtx.Variables,
		// }
		return next(ctx)
	})
}

func logInfoFromRequestContext(ctx *graphql.RequestContext) interface{} {
	var queries []interface{}
	for _, op := range ctx.Doc.Operations {
		for _, sel := range op.SelectionSet {
			if field, ok := sel.(*ast.Field); ok {
				queries = append(queries, map[string]interface{}{
					"op":    op.Operation,
					"name":  op.Name,
					"field": field.Name,
				})
			}
		}
	}
	return queries
}

// SegmentMiddleware tracks the event and sends data to segment
func SegmentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// before reaching the gql library
		next.ServeHTTP(w, r)
		// when the gql library is done
		tags := middleware.GetTags(r.Context())

		// UserID or ServiceID, and SecretID (can be null)
		for _, query := range tags.Query.([]map[string]interface{}) {
			segment.Client.Enqueue(analytics.Track{
				UserId: "test123", // tags.anonymousID,
				Event:  "GraphQL Event",
				Properties: analytics.NewProperties().
					Set("secretID", tags.Secret.SecretID).
					Set("userID", tags.Secret.UserID).
					Set("serviceID", tags.Secret.ServiceID).
					Set("ipAdress", r.RemoteAddr).
					Set("gqlOp", query["op"]).
					Set("gqlOpName", query["name"]),
			})
		}
	})
}
