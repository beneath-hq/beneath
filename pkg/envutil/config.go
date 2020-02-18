package envutil

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Env represents a runtime environment
type Env string

const (
	// Development is used for running locally
	Development Env = "development"

	// Production is used in normal operation
	Production Env = "production"

	// Test is used for running automated tests
	Test Env = "test"
)

// GetEnv reads the ENV environment variable. It panics if it's not set.
func GetEnv() Env {
	env := os.Getenv("ENV")
	switch env {
	case "production":
		return Production
	case "prod":
		return Production
	case "development":
		return Development
	case "dev":
		return Development
	case "test":
		return Test
	case "":
		panic(fmt.Errorf("ENV not set"))
	default:
		panic(fmt.Errorf("ENV <%s> not recognized", env))
	}
}

// LoadConfig loads config from env and panics on error
// For more details, see https://github.com/kelseyhightower/envconfig
func LoadConfig(prefix string, spec interface{}) {
	// load .env if present
	env := GetEnv()
	loadEnv(fmt.Sprintf("configs/.%s.env", string(env)))

	// parse env into config
	err := envconfig.Process(prefix, spec)
	if err != nil {
		panic(err)
	}
}

// loadEnv attempts to load .env files into the process. It panics on error.
func loadEnv(paths ...string) {
	for _, path := range paths {
		err := godotenv.Load(path)
		if err != nil && err.Error() != fmt.Sprintf("open %s: no such file or directory", path) {
			panic(err)
		}
	}
}
