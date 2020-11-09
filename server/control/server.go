package control

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"go.uber.org/zap"

	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/server/control/gql"
	"gitlab.com/beneath-hq/beneath/server/control/resolver"
	"gitlab.com/beneath-hq/beneath/services/middleware"
	"gitlab.com/beneath-hq/beneath/services/organization"
	"gitlab.com/beneath-hq/beneath/services/permissions"
	"gitlab.com/beneath-hq/beneath/services/project"
	"gitlab.com/beneath-hq/beneath/services/secret"
	"gitlab.com/beneath-hq/beneath/services/service"
	"gitlab.com/beneath-hq/beneath/services/stream"
	"gitlab.com/beneath-hq/beneath/services/usage"
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
	Router *chi.Mux

	Options       *ServerOptions
	Logger        *zap.SugaredLogger
	Usage         *usage.Service
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
	logger *zap.Logger,
	usage *usage.Service,
	middleware *middleware.Service,
	organization *organization.Service,
	permissions *permissions.Service,
	project *project.Service,
	secret *secret.Service,
	service *service.Service,
	stream *stream.Service,
	user *user.Service,
) *Server {
	l := logger.Named("control.server")
	router := chi.NewRouter()
	server := &Server{
		Router:        router,
		Options:       opts,
		Logger:        l.Sugar(),
		Usage:         usage,
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
	router.Use(middleware.LoggerMiddleware(l))
	router.Use(middleware.RecovererMiddleware(l))
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

	// Add graphql server
	gqlsrv := handler.New(server.makeExecutableSchema())
	gqlsrv.AddTransport(transport.GET{})
	gqlsrv.AddTransport(transport.POST{})
	gqlsrv.Use(extension.Introspection{})
	gqlsrv.AroundResponses(middleware.QueryLoggingGQLMiddleware)
	gqlsrv.SetErrorPresenter(middleware.DefaultGQLErrorPresenter)
	gqlsrv.SetRecoverFunc(middleware.DefaultGQLRecoverFunc(l))
	router.Handle("/graphql", gqlsrv)
	router.Handle("/playground", playground.Handler("Beneath", "/graphql"))

	return server
}

// Run starts the control server
func (s *Server) Run(ctx context.Context) error {
	s.Logger.Infof("serving on port %d", s.Options.Port)
	httpServer := &http.Server{Handler: s.Router}
	return httputil.ListenAndServeContext(ctx, httpServer, s.Options.Port)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	healthy := true // TODO
	if healthy {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	} else {
		s.Logger.Errorf("control database health check failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (s *Server) makeExecutableSchema() graphql.ExecutableSchema {
	return gql.NewExecutableSchema(gql.Config{Resolvers: &resolver.Resolver{
		Usage:         s.Usage,
		Organizations: s.Organizations,
		Permissions:   s.Permissions,
		Projects:      s.Projects,
		Secrets:       s.Secrets,
		Services:      s.Services,
		Streams:       s.Streams,
		Users:         s.Users,
	}})
}
