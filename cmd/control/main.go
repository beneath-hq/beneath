package main

import (
	"fmt"
	"net/http"

	"github.com/beneath-core/control"
	"github.com/beneath-core/control/auth"
	"github.com/beneath-core/control/migrations"
	"github.com/beneath-core/control/payments"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/internal/segment"
	"github.com/beneath-core/pkg/envutil"
	"github.com/beneath-core/pkg/log"
)

type configSpecification struct {
	ControlPort  int    `envconfig:"CONTROL_PORT" required:"true" default:"8080"`
	ControlHost  string `envconfig:"CONTROL_HOST" required:"true"`
	FrontendHost string `envconfig:"FRONTEND_HOST" required:"true"`

	RedisURL         string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresDB       string `envconfig:"CONTROL_POSTGRES_DB" required:"true"`
	PostgresUser     string `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`

	MQDriver        string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LookupDriver    string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`

	PaymentsDrivers  []string `envconfig:"CONTROL_PAYMENTS_DRIVERS" required:"true"`
	SegmentSecret    string   `envconfig:"CONTROL_SEGMENT_SECRET" required:"true"`
	StripeSecret     string   `envconfig:"CONTROL_STRIPE_SECRET" required:"true"`
	SessionSecret    string   `envconfig:"CONTROL_SESSION_SECRET" required:"true"`
	GithubAuthID     string   `envconfig:"CONTROL_GITHUB_AUTH_ID" required:"true"`
	GithubAuthSecret string   `envconfig:"CONTROL_GITHUB_AUTH_SECRET" required:"true"`
	GoogleAuthID     string   `envconfig:"CONTROL_GOOGLE_AUTH_ID" required:"true"`
	GoogleAuthSecret string   `envconfig:"CONTROL_GOOGLE_AUTH_SECRET" required:"true"`
}

func main() {
	// load config
	var config configSpecification
	envutil.LoadConfig("beneath", &config)

	// Init logging
	log.InitLogger()

	// connect postgres, redis, engine, and payment drivers
	hub.InitPostgres(config.PostgresHost, config.PostgresDB, config.PostgresUser, config.PostgresPassword)
	hub.InitRedis(config.RedisURL)
	hub.InitEngine(config.MQDriver, config.LookupDriver, config.WarehouseDriver)
	hub.SetPaymentDrivers(payments.InitDrivers(config.PaymentsDrivers))

	// run migrations
	migrations.MustRunUp(hub.DB)

	// init segment
	segment.InitClient(config.SegmentSecret)

	// configure auth
	auth.InitGoth(&auth.GothConfig{
		ClientHost:       config.FrontendHost,
		SessionSecret:    config.SessionSecret,
		BackendHost:      config.ControlHost,
		GithubAuthID:     config.GithubAuthID,
		GithubAuthSecret: config.GithubAuthSecret,
		GoogleAuthID:     config.GoogleAuthID,
		GoogleAuthSecret: config.GoogleAuthSecret,
	})

	// run handler
	handler := control.Handler(config.ControlHost, config.FrontendHost)
	log.S.Infow("control http started", "port", config.ControlPort)
	log.S.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.ControlPort), handler))
}
