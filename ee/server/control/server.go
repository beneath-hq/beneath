package control

import (
	"net/http"

	"github.com/go-chi/chi"
	"gitlab.com/beneath-hq/beneath/ee/services/billing"
	"gitlab.com/beneath-hq/beneath/services/organization"
)

// Server is the enterprise control server.
// It provides *additional* enterprise edition related functionality.
// It doesn't run standalone, but gets mounted on the "/ee" path of the
// non-enterprise control server's Mux.
type Server struct {
	Router *chi.Mux

	Billing      *billing.Service
	Organization *organization.Service
}

// NewServer returns a new enterprise control server
func NewServer(billing *billing.Service, organization *organization.Service) *Server {
	router := chi.NewRouter()
	server := &Server{
		Router:       router,
		Billing:      billing,
		Organization: organization,
	}

	// // Add payments handlers
	// for name, driver := range hub.PaymentDrivers {
	// 	for subpath, handler := range driver.GetHTTPHandlers() {
	// 		path := fmt.Sprintf("/billing/%s/%s", name, subpath)
	// 		router.Handle(path, httputil.AppHandler(handler))
	// 	}
	// }

	// Add graphql server
	// gqlHandler := handler.New(server.makeExecutableSchema())
	// gqlHandler.AroundResponses(graphqlQueryLoggingMiddleware)
	// gqlHandler.SetErrorPresenter(graphqlErrorPresenter)
	// gqlHandler.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
	// 	panic(err)
	// })
	// router.Handle("/graphql", gqlHandler)

	router.Get("/hello", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		res.Write([]byte("hello world"))
	})

	return server
}
