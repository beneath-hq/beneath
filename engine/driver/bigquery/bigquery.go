package bigquery

import (
	"context"
	"sync"

	bq "cloud.google.com/go/bigquery"

	"gitlab.com/beneath-hq/beneath/pkg/envutil"
	"gitlab.com/beneath-hq/beneath/engine/driver"
)

// configSpecification defines the config variables to load from ENV
type configSpecification struct {
	ProjectID string `envconfig:"PROJECT_ID" required:"true"`
}

// BigQuery implements beneath.WarehouseService
type BigQuery struct {
	Client *bq.Client
}

const (
	// ProjectIDLabel is a bigquery label key for the project ID
	ProjectIDLabel = "project_id"

	// StreamIDLabel is a bigquery label key for a stream ID
	StreamIDLabel = "stream_id"

	// InstanceIDLabel is a bigquery label key for an instance ID
	InstanceIDLabel = "instance_id"

	// InternalDatasetName is the dataset that stores all raw records
	InternalDatasetName = "__internal"
)

// Global
var global BigQuery
var once sync.Once

func createGlobal() {
	// parse config from env
	var config configSpecification
	envutil.LoadConfig("beneath_engine_bigquery", &config)

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
