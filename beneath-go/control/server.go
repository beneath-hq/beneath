package control

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/beneath-core/beneath-go/core/httputil"

	"github.com/vektah/gqlparser/ast"

	"github.com/beneath-core/beneath-go/core/log"

	"github.com/beneath-core/beneath-go/control/auth"
	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/migrations"
	"github.com/beneath-core/beneath-go/control/resolver"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/middleware"
	"github.com/beneath-core/beneath-go/core/segment"
	"github.com/beneath-core/beneath-go/core/stripe"
	"github.com/beneath-core/beneath-go/db"
	stripe_go "github.com/stripe/stripe-go"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	uuid "github.com/satori/go.uuid"
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

	MQDriver        string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LookupDriver    string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`

	SegmentSecret    string `envconfig:"CONTROL_SEGMENT_SECRET" required:"true"`
	StripeSecret     string `envconfig:"CONTROL_STRIPE_SECRET" required:"true"`
	SessionSecret    string `envconfig:"CONTROL_SESSION_SECRET" required:"true"`
	GithubAuthID     string `envconfig:"CONTROL_GITHUB_AUTH_ID" required:"true"`
	GithubAuthSecret string `envconfig:"CONTROL_GITHUB_AUTH_SECRET" required:"true"`
	GoogleAuthID     string `envconfig:"CONTROL_GOOGLE_AUTH_ID" required:"true"`
	GoogleAuthSecret string `envconfig:"CONTROL_GOOGLE_AUTH_SECRET" required:"true"`
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
	db.InitEngine(Config.MQDriver, Config.LookupDriver, Config.WarehouseDriver)

	// run migrations
	migrations.MustRunUp(db.DB)

	// init segment
	segment.InitClient(Config.SegmentSecret)

	// init stripe
	stripe.InitClient(Config.StripeSecret)

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
	router.Use(segmentMiddleware)

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

	// Add stripe webhook
	router.Handle("/webhook", httputil.AppHandler(handleStripeWebhook))

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

func segmentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// run the request first, thus setting tags.Payload
		next.ServeHTTP(w, r)

		tags := middleware.GetTags(r.Context())
		logs, ok := tags.Payload.([]gqlLog)
		if !ok {
			return
		}

		for _, l := range logs {
			name := "GQL: " + l.Name
			segment.TrackHTTP(r, name, l)
		}
	})
}

// TODO: move this code to its own server
// TODO(review): Yes; Expose the webhook from Stripe, then register it in the control servere;
//               and not under "/webhook", but rather "/billing/stripe/webhook"
func handleStripeWebhook(w http.ResponseWriter, req *http.Request) error {
	ctx := context.Background()
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.S.Errorf("Error reading request body: %v\\n", err)
		return err
	}

	event := stripe_go.Event{}

	if err := json.Unmarshal(payload, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.S.Errorf("Failed to parse webhook body json: %v\\n", err.Error())
		return err
	}

	switch event.Type {
	case "setup_intent.succeeded":
		var setupIntent stripe_go.SetupIntent
		err := json.Unmarshal(event.Data.Raw, &setupIntent)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.S.Errorf("Error parsing webhook JSON: %v\\n", err)
			return err
		}

		organization := entity.FindOrganization(ctx, uuid.FromStringOrNil(setupIntent.Metadata["OrganizationID"]))
		if organization == nil {
			panic("organization not found")
		}

		paymentMethod := stripe.RetrievePaymentMethod(setupIntent.PaymentMethod.ID)

		customer, err := stripe.CreateCustomer(organization.Name, paymentMethod.BillingDetails.Email, setupIntent.PaymentMethod)
		if err != nil {
			log.S.Errorf("Stripe error: %s", err.Error())
			return err
		}

		organization.UpdateStripeCustomerID(ctx, customer.ID)
		organization.UpdatePaymentMethod(ctx, entity.CardPaymentMethod)
		organization.UpdateBillingPlanID(ctx, uuid.FromStringOrNil(setupIntent.Metadata["BillingPlanID"]))
		// TODO: update organization user permissions? maybe put this in the organization.UpdateBillingPlanID() function
	case "setup_intent.setup_failed":
		// stripe.HandleSetupIntentFailed(setupIntent)
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.S.Errorf("Unexpected event type: %s", event.Type)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
