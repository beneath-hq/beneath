package bigquery

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
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

// GetMaxDataSize implements engine.WarehouseDriver
func (b *BigQuery) GetMaxDataSize() int {
	return 1000000
}

// RegisterProject  implements engine.WarehouseDriver
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

// RegisterStreamInstance implements engine.WarehouseDriver
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

	// create view
	viewMeta := &bq.TableMetadata{
		Description: meta.Description,
		Labels:      meta.Labels,
		ViewQuery:   fmt.Sprintf("select * except(`__key`, `__data`, `__timestamp`) from `%s`", fullyQualifiedName(table)),
	}
	table = dataset.Table(makeViewName(streamName))
	err = table.Create(context.Background(), viewMeta)
	if err != nil {
		return err
	}

	// done
	return nil
}

// Row represents a record and implements bigquery.ValueSaver for use in WriteRecord
type Row struct {
	Data     map[string]interface{}
	InsertID string
}

// Save implements bigquery.ValueSaver
func (r *Row) Save() (row map[string]bq.Value, insertID string, err error) {
	data := make(map[string]bq.Value, len(r.Data))
	for k, v := range r.Data {
		data[k] = r.recursiveSerialize(v)
	}
	return data, r.InsertID, nil
}

// the bigquery client serializes every type correctly except big numbers and byte arrays;
// we handle those by a recursive search
// note: overrides in place
func (r *Row) recursiveSerialize(valT interface{}) bq.Value {
	switch val := valT.(type) {
	case *big.Int:
		return val.String()
	case *big.Rat:
		return val.FloatString(0)
	case []byte:
		return "0x" + hex.EncodeToString(val)
	case map[string]interface{}:
		for k, v := range val {
			val[k] = r.recursiveSerialize(v)
		}
	case []interface{}:
		for i, v := range val {
			val[i] = r.recursiveSerialize(v)
		}
	}
	return valT
}

// WriteRecords implements engine.WarehouseDriver
func (b *BigQuery) WriteRecords(projectName string, streamName string, instanceID uuid.UUID, keys [][]byte, avros [][]byte, records []map[string]interface{}, timestamps []time.Time) error {
	// ensure all WriteRequest objects the same length
	if !(len(keys) == len(records) && len(keys) == len(avros) && len(keys) == len(timestamps)) {
		return fmt.Errorf("error: keys, avros, data, and timestamps do not all have the same length")
	}

	// create bigquery uploader
	dataset := makeDatasetName(projectName)
	table := makeTableName(streamName, instanceID)
	u := b.Client.Dataset(dataset).Table(table).Inserter()

	// create a BigQuery Row out of each of the records in the WriteRequest
	rows := make([]*Row, len(keys))
	for i, key := range keys {
		// add meta fields to be uploaded
		records[i]["__key"] = key
		records[i]["__data"] = avros[i]
		records[i]["__timestamp"] = timestamps[i]

		// data to be uploaded
		timestampBytes, err := timestamps[i].MarshalBinary()
		if err != nil {
			log.Panic(err.Error())
		}
		insertIDBytes := append(key, timestampBytes...)
		insertID := base64.StdEncoding.EncodeToString(insertIDBytes)
		rows[i] = &Row{
			Data:     records[i],
			InsertID: insertID,
		}
	}

	// upload all the rows at once
	err := u.Put(context.Background(), rows)
	if err != nil {
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

func makeViewName(streamName string) string {
	return strings.ReplaceAll(streamName, "-", "_")
}

func fullyQualifiedName(table *bq.Table) string {
	return strings.ReplaceAll(table.FullyQualifiedName(), ":", ".")
}
