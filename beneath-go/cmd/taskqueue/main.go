package main

import (
	"github.com/beneath-core/beneath-go/control/taskqueue/worker"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/db"

	// import modules that register tasks in taskqueue
	_ "github.com/beneath-core/beneath-go/control/entity"
)

type configSpecification struct {
	MQDriver         string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LogDriver        string `envconfig:"ENGINE_LOG_DRIVER" required:"true"`
	LookupDriver     string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver  string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL         string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresUser     string `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`
}

func main() {
	var config configSpecification
	core.LoadConfig("beneath", &config)

	db.InitPostgres(config.PostgresHost, config.PostgresUser, config.PostgresPassword)
	db.InitRedis(config.RedisURL)
	db.InitEngine(Config.MQDriver, Config.LogDriver, Config.LookupDriver, Config.WarehouseDriver)

	log.S.Fatal(worker.Work())
}
