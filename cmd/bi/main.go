package main

import (
	"context"
	"os"

	"gitlab.com/beneath-hq/beneath/bi/worker"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/ctxutil"
	"gitlab.com/beneath-hq/beneath/pkg/envutil"
	"gitlab.com/beneath-hq/beneath/pkg/log"

	// import modules that register tasks in taskqueue
	_ "gitlab.com/beneath-hq/beneath/control/entity"
)

type configSpecification struct {
	MQDriver         string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LookupDriver     string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver  string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL         string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresDB       string `envconfig:"CONTROL_POSTGRES_DB" required:"true"`
	PostgresUser     string `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`
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

	// Run forever (until failure)
	ctx := ctxutil.WithCancelOnTerminate(context.Background())
	err := worker.Work(ctx)
	if err != nil {
		log.S.Fatal(err)
	}
	os.Exit(0)
}
