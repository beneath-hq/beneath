package control

import (
	"fmt"
	"log"
	"net/http"

	"github.com/beneath-core/beneath-go/control/db"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/migrations"
	"github.com/beneath-core/beneath-go/control/resolver"
	"github.com/beneath-core/beneath-go/core"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
)

type configSpecification struct {
	HTTPPort    int    `envconfig:"PORT" default:"4000"`
	RedisURL    string `envconfig:"REDIS_URL" required:"true"`
	PostgresURL string `envconfig:"POSTGRES_URL" required:"true"`
}

var (
	// Config for control
	Config configSpecification
)

func init() {
	// load config
	core.LoadConfig("beneath", &Config)

	// connect postgres and redis
	db.InitPostgres(Config.PostgresURL)
	db.InitRedis(Config.RedisURL)

	// run migrations
	migrations.MustRunUp(db.DB)
}

// ListenAndServeHTTP serves the GraphQL API on HTTP
func ListenAndServeHTTP(port int) error {
	router := chi.NewRouter()

	// Add CORS
	router.Use(cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:4000",
			"https://beneath.network",
		},
		AllowCredentials: true,
		Debug:            true,
	}).Handler)

	// Add health check
	router.Get("/", healthCheck)
	router.Get("/healthz", healthCheck)

	// Add playground
	router.Handle("/playground", handler.Playground("Beneath", "/graphql"))

	// Add graphql server
	router.Handle("/graphql",
		handler.GraphQL(gql.NewExecutableSchema(gql.Config{Resolvers: &resolver.Resolver{}})),
	)

	// Serve
	log.Printf("HTTP server running on port %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	_, err := db.DB.Exec("SELECT 1")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.WriteHeader(200)
	w.Write([]byte("success"))
}
