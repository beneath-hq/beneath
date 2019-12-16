package main

import (
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/payments"
	"github.com/beneath-core/beneath-go/taskqueue"

	// import modules that register tasks in taskqueue
	_ "github.com/beneath-core/beneath-go/control/billing"
	_ "github.com/beneath-core/beneath-go/control/entity"
)

type configSpecification struct {
	StreamsDriver    string   `envconfig:"ENGINE_STREAMS_DRIVER" required:"true"`
	TablesDriver     string   `envconfig:"ENGINE_TABLES_DRIVER" required:"true"`
	WarehouseDriver  string   `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL         string   `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string   `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresUser     string   `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string   `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`
	PaymentsDrivers  []string `envconfig:"PAYMENTS_DRIVERS" required:"true"`
}

func main() {
	var config configSpecification
	core.LoadConfig("beneath", &config)

	db.InitPostgres(config.PostgresHost, config.PostgresUser, config.PostgresPassword)
	db.InitRedis(config.RedisURL)
	db.InitEngine(config.StreamsDriver, config.TablesDriver, config.WarehouseDriver)

	payments.InitDrivers(config.PaymentsDrivers)

	log.S.Fatal(taskqueue.Work())
}
