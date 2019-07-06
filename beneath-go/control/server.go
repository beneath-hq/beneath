package control

import (
	"fmt"
	"log"
	"net/http"

	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/resolver"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
)

// ListenAndServeHTTP serves the GraphQL API on HTTP
func ListenAndServeHTTP(port int) error {
	router := chi.NewRouter()

	// Add CORS
	router.Use(cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:8080",
		},
		AllowCredentials: true,
		Debug:            true,
	}).Handler)

	// Add playground
	router.Handle("/", handler.Playground("Beneath", "/graphql"))

	// Add graphql server
	router.Handle("/graphql",
		handler.GraphQL(gql.NewExecutableSchema(gql.Config{Resolvers: &resolver.Resolver{}})),
	)

	log.Printf("HTTP server running on port %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}
