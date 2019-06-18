package core

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// LoadConfig loads config from env and panics on error
// For more details, see https://github.com/kelseyhightower/envconfig
func LoadConfig(prefix string, spec interface{}) {
	// load .env if present
	err := godotenv.Load()
	if err != nil && err.Error() != "open .env: no such file or directory" {
		log.Fatal(err.Error())
	}

	// parse env into config
	err = envconfig.Process(prefix, spec)
	if err != nil {
		log.Fatalf("%v: %v", prefix, err.Error())
	}
}
