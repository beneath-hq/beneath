package bigquery

import (
	"context"
	"fmt"

	bq "cloud.google.com/go/bigquery"
	"github.com/mitchellh/mapstructure"

	"gitlab.com/beneath-hq/beneath/infrastructure/engine/driver"
)

// BigQuery implements beneath.WarehouseService
type BigQuery struct {
	ProjectID        string
	Client           *bq.Client
	InstancesDataset *bq.Dataset
}

// Options for connecting to Postgres
type Options struct {
	ProjectID          string `mapstructure:"project_id"`
	InstancesDatasetID string `mapstructure:"instances_dataset_id"`
}

func init() {
	driver.AddDriver("bigquery", newBigQuery)
}

func newBigQuery(optsMap map[string]interface{}) (driver.Service, error) {
	// load options
	var opts Options
	err := mapstructure.Decode(optsMap, &opts)
	if err != nil {
		return nil, fmt.Errorf("error decoding bigquery options: %s", err.Error())
	}

	// create client
	client, err := bq.NewClient(context.Background(), opts.ProjectID)
	if err != nil {
		return nil, err
	}

	// create instances dataset if it doesn't exist
	instancesDataset := client.Dataset(opts.InstancesDatasetID)
	err = instancesDataset.Create(context.Background(), nil)
	if err != nil && !isAlreadyExists(err) {
		return nil, err
	}

	return &BigQuery{
		ProjectID:        opts.ProjectID,
		Client:           client,
		InstancesDataset: instancesDataset,
	}, nil
}

// AsLookupService implements Service
func (b *BigQuery) AsLookupService() driver.LookupService {
	return nil
}

// AsWarehouseService implements Service
func (b *BigQuery) AsWarehouseService() driver.WarehouseService {
	return b
}

// AsUsageService implements Service
func (b *BigQuery) AsUsageService() driver.UsageService {
	return nil
}
