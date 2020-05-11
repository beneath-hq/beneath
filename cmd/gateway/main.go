package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	gw "gitlab.com/beneath-hq/beneath/gateway"
	gwgrpc "gitlab.com/beneath-hq/beneath/gateway/grpc"
	gwhttp "gitlab.com/beneath-hq/beneath/gateway/http"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/ctxutil"
	"gitlab.com/beneath-hq/beneath/pkg/envutil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
)

type configSpecification struct {
	HTTPPort         int    `envconfig:"GATEWAY_PORT" default:"8080"`
	GRPCPort         int    `envconfig:"GATEWAY_PORT_GRPC" default:"9090"`
	MQDriver         string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LookupDriver     string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver  string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL         string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresDB       string `envconfig:"CONTROL_POSTGRES_DB" required:"true"`
	PostgresUser     string `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`
}

const (
	metricsCacheSize      = 2500
	metricsCommitInterval = 30 * time.Second
)

func main() {
	// Config for gateway
	var config configSpecification
	envutil.LoadConfig("beneath", &config)

	// Init logging
	log.InitLogger()

	// Init connections
	hub.InitPostgres(config.PostgresHost, config.PostgresDB, config.PostgresUser, config.PostgresPassword)
	hub.InitRedis(config.RedisURL)
	hub.InitEngine(config.MQDriver, config.LookupDriver, config.WarehouseDriver)

	// Init gateway globals
	gw.InitMetrics(metricsCacheSize, metricsCommitInterval)
	gw.InitSubscriptions(hub.Engine)

	// A ctx that we can cancel, and that also cancels on sigint, etc.
	ctx, cancel := context.WithCancel(ctxutil.WithCancelOnTerminate(context.Background()))

	// Run listenAndServeHTTP and listenAndServeGRPC in the background, cancel if any of them fails
	go func() {
		err := listenAndServeHTTP(config.HTTPPort)
		log.S.Errorw("http serve failed", "error", err)
		cancel()
	}()

	go func() {
		err := listenAndServeGRPC(config.GRPCPort)
		log.S.Errorw("grpc serve failed", "error", err)
		cancel()
	}()

	// Run metrics and subscriptions in the background, exit when ctx is cancelled and they've finished cleaning up
	group := new(errgroup.Group)

	group.Go(func() error {
		gw.Metrics.RunForever(ctx)
		return nil
	})

	group.Go(func() error {
		gw.Subscriptions.RunForever(ctx)
		return nil
	})

	// run
	err := group.Wait()
	if err != nil {
		log.S.Fatal(err)
	}
	os.Exit(0)
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
