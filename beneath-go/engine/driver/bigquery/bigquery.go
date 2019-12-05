package bigquery

import (
	"context"
	"sync"

	"github.com/beneath-core/beneath-go/engine/driver"

	"github.com/beneath-core/beneath-go/core"

	bq "cloud.google.com/go/bigquery"
)

// configSpecification defines the config variables to load from ENV
type configSpecification struct {
	ProjectID string `envconfig:"PROJECT_ID" required:"true"`
}

// BigQuery implements beneath.WarehouseService
type BigQuery struct {
	Client *bq.Client
}

// Global
var global BigQuery
var once sync.Once

func createGlobal() {
	// parse config from env
	var config configSpecification
	core.LoadConfig("beneath_engine_bigquery", &config)

	// create client
	client, err := bq.NewClient(context.Background(), config.ProjectID)
	if err != nil {
		panic(err)
	}

	// create instance
	global = BigQuery{
		Client: client,
	}
}

// GetWarehouseService todo
func GetWarehouseService() driver.WarehouseService {
	once.Do(createGlobal)
	return global
}
