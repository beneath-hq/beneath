package beneath

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// ConfigSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type ConfigSpecification struct {
	Port              int    `envconfig:"PORT" default:"5000"`
	StreamsPlatform   string `envconfig:"STREAMS_PLATFORM" required:"true"`
	TablesPlatform    string `envconfig:"TABLES_PLATFORM" required:"true"`
	WarehousePlatform string `envconfig:"WAREHOUSE_PLATFORM" required:"true"`
	RedisURL          string `envconfig:"REDIS_URL" required:"true"`
	PostgresURL       string `envconfig:"POSTGRES_URL" required:"true"`
}

var (
	// Config values loaded from env
	Config ConfigSpecification
)

func init() {
	// load .env if present
	err := godotenv.Load()
	if err != nil && err.Error() != "open .env: no such file or directory" {
		log.Fatal(err.Error())
	}

	// parse env into config
	err = envconfig.Process("beneath", &Config)
	if err != nil {
		log.Fatal(err.Error())
	}
}
