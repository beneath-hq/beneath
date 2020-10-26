package envutil

import (
	"fmt"
	"os"
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
