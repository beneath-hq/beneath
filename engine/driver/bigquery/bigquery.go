package bigquery

import (
	"context"
	"sync"

	bq "cloud.google.com/go/bigquery"

	"gitlab.com/beneath-hq/beneath/engine/driver"
	"gitlab.com/beneath-hq/beneath/pkg/envutil"
)

// configSpecification defines the config variables to load from ENV
type configSpecification struct {
	ProjectID          string `envconfig:"PROJECT_ID" required:"true"`
	InstancesDatasetID string `envconfig:"INSTANCES_DATASET_ID" required:"true"`
}

// BigQuery implements beneath.WarehouseService
type BigQuery struct {
	ProjectID        string
	Client           *bq.Client
	InstancesDataset *bq.Dataset
}

const (
	// OriginalStreamPathLabel is a label that gives the original path of the instance
	OriginalStreamPathLabel = "original_stream_path"
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

	// create instances dataset if it doesn't exist
	instancesDataset := client.Dataset(config.InstancesDatasetID)
	err = instancesDataset.Create(context.Background(), nil)
	if err != nil && !isAlreadyExists(err) {
		panic(err)
	}

	// create instance
	global = BigQuery{
		ProjectID:        config.ProjectID,
		Client:           client,
		InstancesDataset: instancesDataset,
	}
}

// GetWarehouseService todo
func GetWarehouseService() driver.WarehouseService {
	once.Do(createGlobal)
	return global
}
