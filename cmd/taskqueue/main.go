package main

import (
	"github.com/beneath-core/control/payments"
	"github.com/beneath-core/control/taskqueue/worker"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/pkg/envutil"
	"github.com/beneath-core/pkg/log"

	// import modules that register tasks in taskqueue
	_ "github.com/beneath-core/control/entity"
)

type configSpecification struct {
	MQDriver         string   `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LookupDriver     string   `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver  string   `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL         string   `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string   `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresDB       string   `envconfig:"CONTROL_POSTGRES_DB" required:"true"`
	PostgresUser     string   `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string   `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`
	PaymentsDrivers  []string `envconfig:"CONTROL_PAYMENTS_DRIVERS" required:"true"`
}

func main() {
	var config configSpecification
	envutil.LoadConfig("beneath", &config)

	log.InitLogger()

	hub.InitPostgres(config.PostgresHost, config.PostgresDB, config.PostgresUser, config.PostgresPassword)
	hub.InitRedis(config.RedisURL)
	hub.InitEngine(config.MQDriver, config.LookupDriver, config.WarehouseDriver)

	hub.SetPaymentDrivers(payments.InitDrivers(config.PaymentsDrivers))

	log.S.Fatal(worker.Work())
}
