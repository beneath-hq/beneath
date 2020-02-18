package integration

import (
	"context"
	"net"
	"net/http/httptest"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/beneath-core/control"
	"github.com/beneath-core/control/auth"
	"github.com/beneath-core/control/migrations"
	"github.com/beneath-core/control/payments"
	"github.com/beneath-core/control/taskqueue/worker"
	"github.com/beneath-core/gateway"
	gwgrpc "github.com/beneath-core/gateway/grpc"
	pb "github.com/beneath-core/gateway/grpc/proto"
	"github.com/beneath-core/gateway/http"
	"github.com/beneath-core/gateway/pipeline"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/internal/segment"
	"github.com/beneath-core/pkg/envutil"
	"github.com/beneath-core/pkg/log"
)

type configSpecification struct {
	ControlPort  int    `envconfig:"CONTROL_PORT" required:"true" default:"8080"`
	ControlHost  string `envconfig:"CONTROL_HOST" required:"true"`
	FrontendHost string `envconfig:"FRONTEND_HOST" required:"true"`

	GatewayPortHTTP int `envconfig:"GATEWAY_PORT" default:"8080"`
	GatewayPortGRPC int `envconfig:"GATEWAY_PORT_GRPC" default:"9090"`

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

var (
	controlHTTP *httptest.Server
	gatewayHTTP *httptest.Server
	gatewayGRPC pb.GatewayClient
)

func TestMain(m *testing.M) {
	// must run at project root to detect configs/
	os.Chdir("../..")

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

	// init segment
	segment.InitClient(config.SegmentSecret)

	// Init gateway globals
	gateway.InitMetrics()
	gateway.InitSubscriptions(hub.Engine)

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

	// run migrations
	_, _, err := migrations.Run(hub.DB, "reset")
	panicIf(err)
	migrations.MustRunUp(hub.DB)

	// flush redis
	_, err = hub.Redis.FlushAll().Result()
	panicIf(err)

	// create control server
	controlHTTP = httptest.NewServer(control.Handler("", ""))
	defer controlHTTP.Close()

	// create gateway HTTP server
	gatewayHTTP = httptest.NewServer(http.Handler())
	defer gatewayHTTP.Close()

	// create gateway GRPC server
	grpcListener := bufconn.Listen(1024 * 1024)
	grpcServer := gwgrpc.Server()
	go func() {
		err := grpcServer.Serve(grpcListener)
		panicIf(err)
	}()
	defer grpcServer.Stop()

	// create gateway GRPC client
	dialer := grpc.WithContextDialer(func(ctx context.Context, url string) (net.Conn, error) {
		return grpcListener.Dial()
	})
	gatewayClient, err := grpc.DialContext(context.Background(), "", dialer, grpc.WithInsecure())
	panicIf(err)
	defer gatewayClient.Close()
	gatewayGRPC = pb.NewGatewayClient(gatewayClient)

	// start pipeline
	go func() {
		err := pipeline.Run()
		panicIf(err)
	}()

	// start taskqueue
	go func() {
		err := worker.Work()
		panicIf(err)
	}()

	// run tests
	code := m.Run()

	// reset database
	_, _, err = migrations.Run(hub.DB, "reset")
	panicIf(err)

	// flush redis
	_, err = hub.Redis.FlushAll().Result()
	panicIf(err)

	// exit with code
	os.Exit(code)
}
