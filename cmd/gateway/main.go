package main

import (
	"fmt"
	"net"
	"net/http"

	"golang.org/x/sync/errgroup"

	"github.com/beneath-core/pkg/envutil"
	"github.com/beneath-core/pkg/log"
	"github.com/beneath-core/internal/segment"
	"github.com/beneath-core/internal/hub"
	gw "github.com/beneath-core/gateway"
	gwgrpc "github.com/beneath-core/gateway/grpc"
	gwhttp "github.com/beneath-core/gateway/http"
)

type configSpecification struct {
	HTTPPort         int    `envconfig:"GATEWAY_PORT" default:"8080"`
	GRPCPort         int    `envconfig:"GATEWAY_PORT_GRPC" default:"9090"`
	SegmentSecret    string `envconfig:"GATEWAY_SEGMENT_SECRET" required:"true"`
	MQDriver         string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LookupDriver     string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver  string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL         string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresUser     string `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`
}

func main() {
	// Config for gateway
	var config configSpecification
	envutil.LoadConfig("beneath", &config)

	// Init connections
	hub.InitPostgres(config.PostgresHost, config.PostgresUser, config.PostgresPassword)
	hub.InitRedis(config.RedisURL)
	hub.InitEngine(config.MQDriver, config.LookupDriver, config.WarehouseDriver)

	// Init gateway globals
	gw.InitMetrics()
	gw.InitSubscriptions(hub.Engine)

	// Init segment
	segment.InitClient(config.SegmentSecret)

	// coordinates multiple servers
	group := new(errgroup.Group)

	// http server
	group.Go(func() error {
		return listenAndServeHTTP(config.HTTPPort)
	})

	// gRPC server
	group.Go(func() error {
		return listenAndServeGRPC(config.GRPCPort)
	})

	// run simultaneously
	log.S.Fatal(group.Wait())
}

func listenAndServeHTTP(port int) error {
	log.S.Infow("gateway http started", "port", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), gwhttp.Handler())
}

func listenAndServeGRPC(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	log.S.Infow("gateway grpc started", "port", port)
	return gwgrpc.Server().Serve(lis)
}
