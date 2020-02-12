package postgres

import (
	"sync"

	"github.com/beneath-core/beneath-go/core/envutil"
	"github.com/beneath-core/beneath-go/engine/driver"
)

// configSpecification defines the config variables to load from ENV
type configSpecification struct {
}

// Postgres implements beneath.LookupService and beneath.WarehouseService
type Postgres struct {
}

// Global
var global Postgres
var once sync.Once

func createGlobal() {
	// parse config from env
	var config configSpecification
	envutil.LoadConfig("beneath_engine_postgres", &config)

	// create instance
	global = Postgres{}
}

// GetLookupService returns a Postgres implementation of beneath.LookupService
func GetLookupService() driver.LookupService {
	once.Do(createGlobal)
	return global
}

// GetWarehouseService returns a Postgres implementation of beneath.LookupService
func GetWarehouseService() driver.WarehouseService {
	once.Do(createGlobal)
	return global
}
