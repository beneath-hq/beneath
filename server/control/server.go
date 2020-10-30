package control

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/server/control/gql"
	"gitlab.com/beneath-hq/beneath/server/control/resolver"
	"gitlab.com/beneath-hq/beneath/services/metrics"
	"gitlab.com/beneath-hq/beneath/services/middleware"
	"gitlab.com/beneath-hq/beneath/services/organization"
	"gitlab.com/beneath-hq/beneath/services/permissions"
	"gitlab.com/beneath-hq/beneath/services/project"
	"gitlab.com/beneath-hq/beneath/services/secret"
	"gitlab.com/beneath-hq/beneath/services/service"
	"gitlab.com/beneath-hq/beneath/services/stream"
	"gitlab.com/beneath-hq/beneath/services/user"
)

// ServerOptions are the options for creating a control server
type ServerOptions struct {
	Host          string
	Port          int
	FrontendHost  string `mapstructure:"frontend_host"`
	SessionSecret string `mapstructure:"session_secret"`
	Auth          AuthOptions
}

// Server is the control server
type Server struct {
	Router  *chi.Mux
	Options *ServerOptions

	Metrics       *metrics.Service
	Organizations *organization.Service
	Permissions   *permissions.Service
	Projects      *project.Service
	Secrets       *secret.Service
	Services      *service.Service
	Streams       *stream.Service
	Users         *user.Service
}

// NewServer returns a new control server
func NewServer(
	opts *ServerOptions,
	metrics *metrics.Service,
	middleware *middleware.Service,
	organization *organization.Service,
	permissions *permissions.Service,
	project *project.Service,
	secret *secret.Service,
	service *service.Service,
	stream *stream.Service,
	user *user.Service,
) *Server {
	router := chi.NewRouter()
	server := &Server{
		Router:        router,
		Options:       opts,
		Metrics:       metrics,
		Organizations: organization,
		Permissions:   permissions,
		Projects:      project,
		Secrets:       secret,
		Services:      service,
		Streams:       stream,
		Users:         user,
	}

	corsOptions := cors.Options{
		AllowedOrigins: []string{
			opts.Host,
			opts.FrontendHost,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}

	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.DefaultCompress)
	router.Use(cors.New(corsOptions).Handler)
	router.Use(middleware.InjectTagsMiddleware)
	router.Use(middleware.LoggerMiddleware)
	router.Use(middleware.RecovererMiddleware)
	router.Use(middleware.AuthMiddleware)
	router.Use(middleware.IPRateLimitMiddleware)

	// Add health check
	router.Get("/", server.healthCheck)
	router.Get("/healthz", server.healthCheck)

	// Register auth endpoints
	server.initGoth()
	server.registerAuth()

	// Add error testing
	router.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("Testing error at %v", time.Now()))
	})

	// Add playground
	router.Handle("/playground", playground.Handler("Beneath", "/graphql"))

	// Add graphql server
	gqlHandler := handler.New(server.makeExecutableSchema())
	gqlHandler.AroundResponses(graphqlQueryLoggingMiddleware)
	gqlHandler.SetErrorPresenter(graphqlErrorPresenter)
	gqlHandler.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
		panic(err)
	})
	router.Handle("/graphql", gqlHandler)

	return server
}

// Run starts the control server
func (s *Server) Run(ctx context.Context) error {
	log.S.Infof("serving control server on port %d", s.Options.Port)
	httpServer := &http.Server{Handler: s.Router}
	return httputil.ListenAndServeContext(ctx, httpServer, s.Options.Port)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	healthy := true // TODO
	if healthy {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	} else {
		log.S.Errorf("Control database health check failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (s *Server) makeExecutableSchema() graphql.ExecutableSchema {
	return gql.NewExecutableSchema(gql.Config{Resolvers: &resolver.Resolver{
		Metrics:       s.Metrics,
		Organizations: s.Organizations,
		Permissions:   s.Permissions,
		Projects:      s.Projects,
		Secrets:       s.Secrets,
		Services:      s.Services,
		Streams:       s.Streams,
		Users:         s.Users,
	}})
}

type gqlLog struct {
	Op    string                 `json:"op"`
	Name  string                 `json:"name,omitempty"`
	Field string                 `json:"field"`
	Error error                  `json:"error,omitempty"`
	Vars  map[string]interface{} `json:"vars,omitempty"`
}

func graphqlErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	tags := middleware.GetTags(ctx)
	if q, ok := tags.Payload.(gqlLog); ok {
		q.Error = err
	}
	// Uncomment this line to print resolver error details in the console
	// fmt.Printf("Error in GraphQL Resolver: %s", err.Error())
	return graphql.DefaultErrorPresenter(ctx, err)
}

func graphqlQueryLoggingMiddleware(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	reqCtx := graphql.GetRequestContext(ctx)
	middleware.SetTagsPayload(ctx, logInfoFromRequestContext(reqCtx))
	return next(ctx)
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
