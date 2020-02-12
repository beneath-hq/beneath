package main

import (
	"github.com/beneath-core/control/taskqueue/worker"
	"github.com/beneath-core/pkg/envutil"
	"github.com/beneath-core/pkg/log"
	"github.com/beneath-core/db"
	"github.com/beneath-core/payments"

	// import modules that register tasks in taskqueue
	_ "github.com/beneath-core/control/entity"
)

type configSpecification struct {
	MQDriver         string   `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LookupDriver     string   `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver  string   `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL         string   `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string   `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresUser     string   `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string   `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`
	PaymentsDrivers  []string `envconfig:"PAYMENTS_DRIVERS" required:"true"`
}

func main() {
	var config configSpecification
	envutil.LoadConfig("beneath", &config)

	db.InitPostgres(config.PostgresHost, config.PostgresUser, config.PostgresPassword)
	db.InitRedis(config.RedisURL)
	db.InitEngine(config.MQDriver, config.LookupDriver, config.WarehouseDriver)
	db.SetPaymentDrivers(payments.InitDrivers(config.PaymentsDrivers))

	log.S.Fatal(worker.Work())
}
