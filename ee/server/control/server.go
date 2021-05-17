package control

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/beneath-hq/beneath/ee/server/control/gql"
	"github.com/beneath-hq/beneath/ee/server/control/resolver"
	"github.com/beneath-hq/beneath/ee/services/billing"
	"github.com/beneath-hq/beneath/ee/services/payments"
	"github.com/beneath-hq/beneath/services/middleware"
	"github.com/beneath-hq/beneath/services/organization"
	"github.com/beneath-hq/beneath/services/permissions"
)

// Server is the enterprise control server.
// It provides *additional* enterprise edition related functionality.
// It doesn't run standalone, but gets mounted on the "/ee" path of the
// non-enterprise control server's Mux.
type Server struct {
	Router *chi.Mux

	Logger        *zap.Logger
	Billing       *billing.Service
	Payments      *payments.Service
	Permissions   *permissions.Service
	Organizations *organization.Service
}

// NewServer returns a new enterprise control server
func NewServer(logger *zap.Logger, billing *billing.Service, middleware *middleware.Service, payments *payments.Service, permissions *permissions.Service, organization *organization.Service) *Server {
	l := logger.Named("control.eeserver")
	router := chi.NewRouter()
	server := &Server{
		Router:        router,
		Logger:        l,
		Billing:       billing,
		Payments:      payments,
		Permissions:   permissions,
		Organizations: organization,
	}

	// not adding middleware since router will be mounted on existing non-EE control server at /ee

	router.Get("/hello", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		res.Write([]byte("hello world"))
	})

	// register payment handlers
	payments.RegisterHandlers(router)

	// Add graphql server
	gqlsrv := handler.New(server.makeExecutableSchema())
	gqlsrv.AddTransport(transport.GET{})
	gqlsrv.AddTransport(transport.POST{})
	gqlsrv.Use(extension.Introspection{})
	gqlsrv.AroundResponses(middleware.QueryLoggingGQLMiddleware)
	gqlsrv.SetErrorPresenter(middleware.DefaultGQLErrorPresenter)
	gqlsrv.SetRecoverFunc(middleware.DefaultGQLRecoverFunc(l))
	router.Handle("/graphql", gqlsrv)
	router.Handle("/playground", playground.Handler("Beneath", "/ee/graphql"))

	return server
}

func (s *Server) makeExecutableSchema() graphql.ExecutableSchema {
	return gql.NewExecutableSchema(gql.Config{Resolvers: &resolver.Resolver{
		Billing:       s.Billing,
		Permissions:   s.Permissions,
		Organizations: s.Organizations,
	}})
}
