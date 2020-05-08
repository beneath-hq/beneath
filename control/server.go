package control

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/beneath-hq/beneath/control/auth"
	"gitlab.com/beneath-hq/beneath/control/gql"
	"gitlab.com/beneath-hq/beneath/control/resolver"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/log"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/gqlerror"
)

// Handler serves the GraphQL API on HTTP
func Handler(host string, frontendHost string) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.DefaultCompress)
	router.Use(middleware.InjectTags)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Auth)
	router.Use(middleware.IPRateLimit())

	// Add CORS
	router.Use(cors.New(cors.Options{
		AllowedOrigins: []string{
			host,
			frontendHost,
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

	// Add error testing
	router.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("Testing error at %v", time.Now()))
	})

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

	// Add payments handlers
	for name, driver := range hub.PaymentDrivers {
		for subpath, handler := range driver.GetHTTPHandlers() {
			path := fmt.Sprintf("/billing/%s/%s", name, subpath)
			router.Handle(path, httputil.AppHandler(handler))
		}
	}

	return router
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if hub.Healthy() {
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

type gqlLog struct {
	Op    string                 `json:"op"`
	Name  string                 `json:"name,omitempty"`
	Field string                 `json:"field"`
	Error error                  `json:"error,omitempty"`
	Vars  map[string]interface{} `json:"vars,omitempty"`
}

func makeQueryLoggingMiddleware() handler.Option {
	return handler.RequestMiddleware(func(ctx context.Context, next func(ctx context.Context) []byte) []byte {
		reqCtx := graphql.GetRequestContext(ctx)
		middleware.SetTagsPayload(ctx, logInfoFromRequestContext(reqCtx))
		return next(ctx)
	})
}

func logInfoFromRequestContext(ctx *graphql.RequestContext) interface{} {
	var queries []gqlLog
	for _, op := range ctx.Doc.Operations {
		for _, sel := range op.SelectionSet {
			if field, ok := sel.(*ast.Field); ok {
				name := op.Name
				if name == "" {
					name = "Unnamed"
				}
				queries = append(queries, gqlLog{
					Op:    string(op.Operation),
					Name:  name,
					Field: field.Name,
					Vars:  ctx.Variables,
				})
			}
		}
	}
	return queries
}

func makeGraphQLErrorPresenter() handler.Option {
	return handler.ErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		tags := middleware.GetTags(ctx)
		if q, ok := tags.Payload.(gqlLog); ok {
			q.Error = err
		}
		// Uncomment this line to print resolver error details in the console
		// fmt.Printf("Error in GraphQL Resolver: %s", err.Error())
		return graphql.DefaultErrorPresenter(ctx, err)
	})
}
