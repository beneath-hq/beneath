package main

import (
	"github.com/beneath-core/pkg/envutil"
	"github.com/beneath-core/pkg/log"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/pipeline"
)

type configSpecification struct {
	MQDriver         string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LookupDriver     string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver  string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL         string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresUser     string `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`
}

func main() {
	var config configSpecification
	envutil.LoadConfig("beneath", &config)

	hub.InitPostgres(config.PostgresHost, config.PostgresUser, config.PostgresPassword)
	hub.InitRedis(config.RedisURL)
	hub.InitEngine(config.MQDriver, config.LookupDriver, config.WarehouseDriver)

	log.S.Info("pipeline started")
	log.S.Fatal(pipeline.Run())
}
