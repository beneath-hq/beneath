package bigquery

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	bq "cloud.google.com/go/bigquery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-core/beneath-go/core"
	uuid "github.com/satori/go.uuid"
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

// Row type gets fed into BigQuery
type Row struct {
	Data     map[string]interface{}
	InsertID string
}

// Save implements ValueSaver for the Row type
func (r *Row) Save() (row map[string]bq.Value, insertID string, err error) {
	data := make(map[string]bq.Value, len(r.Data))
	for k, v := range r.Data {
		data[k] = bq.Value(v)
	}
	return data, r.InsertID, nil
}

const (
	// ProjectIDLabel is a bigquery label key for the project ID
	ProjectIDLabel = "project_id"

	// StreamIDLabel is a bigquery label key for a stream ID
	StreamIDLabel = "stream_id"

	// InstanceIDLabel is a bigquery label key for an instance ID
	InstanceIDLabel = "instance_id"
)

// New returns a new
func New() *BigQuery {
	// parse config from env
	var config configSpecification
	core.LoadConfig("beneath_engine_bigquery", &config)

	// create client
	client, err := bq.NewClient(context.Background(), config.ProjectID)
	if err != nil {
		log.Fatalf("could not create bigquery client: %v", err)
	}

	// create instance
	return &BigQuery{
		Client: client,
	}
}

// GetMaxDataSize implements beneath.WarehouseDriver
func (b *BigQuery) GetMaxDataSize() int {
	return 1000000
}

// RegisterProject prepares a bigquery dataset for a project with the given name
func (b *BigQuery) RegisterProject(projectID uuid.UUID, public bool, name, displayName, description string) error {
	// prepare access entries if public (otherwise leaving as default)
	var access []*bq.AccessEntry
	if public {
		access = append(access, &bq.AccessEntry{
			Role:       bq.ReaderRole,
			EntityType: bq.SpecialGroupEntity,
			Entity:     "allAuthenticatedUsers",
		})
	}

	// prepare dataset metadata
	meta := &bq.DatasetMetadata{
		Name:        displayName,
		Description: description,
		Labels: map[string]string{
			ProjectIDLabel: projectID.String(),
		},
		Access: access,
	}

	// create dataset for project
	err := b.Client.Dataset(makeDatasetName(name)).Create(context.Background(), meta)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Panicf("error creating dataset for project '%s': %v", name, err)
		} else {
			log.Printf("trying to create dataset that already exists for project '%s'", name)
		}
	}

	// done
	return nil
}

// RegisterStreamInstance prepares a bigquery table for a stream instance
func (b *BigQuery) RegisterStreamInstance(projectID uuid.UUID, projectName string, streamID uuid.UUID, streamName string, streamDescription string, schemaJSON string, keyFields []string, instanceID uuid.UUID) error {
	// build schema object
	schema, err := bq.SchemaFromJSON([]byte(schemaJSON))
	if err != nil {
		return err
	}

	// build meta
	meta := &bq.TableMetadata{
		Description:      streamDescription,
		Schema:           schema,
		TimePartitioning: &bq.TimePartitioning{
			// TODO:
		},
		Clustering: &bq.Clustering{
			Fields: keyFields,
		},
		Labels: map[string]string{
			ProjectIDLabel:  projectID.String(),
			StreamIDLabel:   streamID.String(),
			InstanceIDLabel: instanceID.String(),
		},
	}

	// create table
	dataset := b.Client.Dataset(makeDatasetName(projectName))
	table := dataset.Table(makeTableName(streamName, instanceID))
	err = table.Create(context.Background(), meta)
	if err != nil {
		// TODO: very unlikely, but should probably delete old table and create new (to ensure correct schema)
		return err
	}

	// done
	return nil
}

func makeDatasetName(projectName string) string {
	return strings.ReplaceAll(projectName, "-", "_")
}

func makeTableName(streamName string, instanceID uuid.UUID) string {
	name := strings.ReplaceAll(streamName, "-", "_")
	return fmt.Sprintf("%s_%s", name, hex.EncodeToString(instanceID[0:4]))
}

// WriteRecordToWarehouse ...
func (b *BigQuery) WriteRecordToWarehouse(projectName string, streamName string, instanceID uuid.UUID, encodedKey []byte, decodedMap map[string]interface{}, sequenceNumber int64) error {
	ctx := context.Background()

	// compute dataset + table names
	datasetName := makeDatasetName(projectName)
	tableName := makeTableName(streamName, instanceID)

	// create an uploader in order to upload data to the table
	u := b.Client.Dataset(datasetName).Table(tableName).Uploader()

	// add metafields to be uploaded
	decodedMap["_insert_time"] = time.Now()
	decodedMap["_key"] = encodedKey
	decodedMap["_sequence_number"] = sequenceNumber

	// data to be uploaded
	insertID := string(append(encodedKey, byte(sequenceNumber)))
	rows := []*Row{
		{Data: decodedMap,
			InsertID: insertID},
	}

	log.Print("writing to warehouse...")
	if err := u.Put(ctx, rows); err != nil {
		if multiError, ok := err.(bq.PutMultiError); ok {
			for _, err1 := range multiError {
				for _, err2 := range err1.Errors {
					fmt.Println(err2)
				}
			}
		} else {
			fmt.Println(err)
		}
		return err
	}
	return nil
}
