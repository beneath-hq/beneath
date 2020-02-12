package hub

import (
	"github.com/beneath-core/control/payments/driver"
	"github.com/beneath-core/engine"
	"github.com/go-pg/pg/v9"
	"github.com/go-redis/redis/v7"
)

var (
	// DB is the postgres connection
	DB *pg.DB

	// Redis connection
	Redis *redis.Client

	// Engine is the data plane
	Engine *engine.Engine

	// PaymentDrivers handle payment methods
	PaymentDrivers map[string]driver.PaymentsDriver
)

// InitPostgres sets up Postgres connection
func InitPostgres(host string, username string, password string) {
	DB = newDatabase(host, username, password)
}

// InitRedis sets up Redis connection
func InitRedis(redisURL string) {
	Redis = newRedis(redisURL)
}

// InitEngine sets up the engine connection
func InitEngine(mqDriver, lookupDriver, warehouseDriver string) {
	Engine = engine.NewEngine(mqDriver, lookupDriver, warehouseDriver)
}

// SetPaymentDrivers injects the payment drivers into the hub.PaymentDrivers object
func SetPaymentDrivers(paymentDrivers map[string]driver.PaymentsDriver) {
	PaymentDrivers = paymentDrivers
}

// Healthy returns true if connections are live
func Healthy() bool {
	// check postgres
	_, err := DB.Exec("SELECT 1")
	pg := err == nil

	// check redis
	_, err = Redis.Ping().Result()
	redis := err == nil

	return pg && redis && Engine.Healthy()
}
