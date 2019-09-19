package bigquery

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	bq "cloud.google.com/go/bigquery"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-core/beneath-go/core"
)

// configSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type configSpecification struct {
	ProjectID string `envconfig:"PROJECT_ID" required:"true"`
}

// BigQuery implements beneath.WarehouseDriver
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

// New returns a new
func New() *BigQuery {
	// parse config from env
	var config configSpecification
	core.LoadConfig("beneath_engine_bigquery", &config)

	// create client
	client, err := bq.NewClient(context.Background(), config.ProjectID)
	if err != nil {
		panic(err)
	}

	// create internal dataset if doesn't exist
	err = client.Dataset(internalDatasetName()).Create(context.Background(), &bq.DatasetMetadata{})
	if err != nil && !isAlreadyExists(err) {
		panic(err)
	}

	// create instance
	return &BigQuery{
		Client: client,
	}
}

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	status, _ := status.FromError(err)
	return status.Code() == codes.AlreadyExists || strings.Contains(err.Error(), "Error 409: Already Exists:")
}

func isExpiredETag(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Error 400: Precondition check failed")
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Error 404: Not found")
}

func internalDatasetName() string {
	return InternalDatasetName
}

func internalTableName(instanceID uuid.UUID) string {
	return strings.ReplaceAll(instanceID.String(), "-", "_")
}

func externalDatasetName(projectName string) string {
	return strings.ReplaceAll(projectName, "-", "_")
}

func externalTableName(streamName string, instanceID uuid.UUID) string {
	name := strings.ReplaceAll(streamName, "-", "_")
	return fmt.Sprintf("%s_%s", name, hex.EncodeToString(instanceID[0:4]))
}

func externalStreamViewName(streamName string) string {
	return strings.ReplaceAll(streamName, "-", "_")
}

func fullyQualifiedName(table *bq.Table) string {
	return strings.ReplaceAll(table.FullyQualifiedName(), ":", ".")
}
