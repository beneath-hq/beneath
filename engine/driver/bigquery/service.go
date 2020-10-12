package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"

	"gitlab.com/beneath-hq/beneath/engine/driver"
	"gitlab.com/beneath-hq/beneath/pkg/schemalang/transpilers"
)

// MaxKeySize implements beneath.Service
func (b BigQuery) MaxKeySize() int {
	return 1024 // 1 kb
}

// MaxRecordSize implements beneath.Service
func (b BigQuery) MaxRecordSize() int {
	return 1048576 // 1 mb
}

// MaxRecordsInBatch implements beneath.Service
func (b BigQuery) MaxRecordsInBatch() int {
	return 10000
}

// RegisterProject implements beneath.Service
func (b BigQuery) RegisterProject(ctx context.Context, p driver.Project) error {
	return nil
}

// RemoveProject implements beneath.Service
func (b BigQuery) RemoveProject(ctx context.Context, p driver.Project) error {
	return nil
}

// RegisterInstance implements beneath.Service
func (b BigQuery) RegisterInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	// get bigquery schema
	schema, err := transpilers.FromAvro(s.GetCodec().AvroSchema)
	if err != nil {
		panic(err)
	}
	bqSchema := transpilers.ToBigQuery(schema, true)

	// inject internal fields
	bqSchema = append(bqSchema, &bigquery.FieldSchema{
		Name:     "__key",
		Type:     bigquery.BytesFieldType,
		Required: true,
	})
	bqSchema = append(bqSchema, &bigquery.FieldSchema{
		Name:     "__timestamp",
		Type:     bigquery.TimestampFieldType,
		Required: true,
	})

	// create time partitioning config
	timePartitioning := &bigquery.TimePartitioning{
		Field:      "__timestamp",
		Expiration: s.GetWarehouseRetention(),
	}

	// create table
	streamPath := fmt.Sprintf("%s-%s-%s", p.GetOrganizationName(), p.GetProjectName(), s.GetStreamName())
	table := b.InstancesDataset.Table(instanceTableName(i.GetStreamInstanceID()))
	err = table.Create(ctx, &bigquery.TableMetadata{
		Schema:           bqSchema,
		TimePartitioning: timePartitioning,
		Clustering: &bigquery.Clustering{
			Fields: s.GetCodec().PrimaryIndex.GetFields(),
		},
		Labels: map[string]string{
			OriginalStreamPathLabel: streamPath,
		},
	})
	if err != nil && !isAlreadyExists(err) { // for idempotency, we don't care if it exists
		return err
	}

	return nil
}

// PromoteInstance implements beneath.Service
func (b BigQuery) PromoteInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	return nil
}

// RemoveInstance implements beneath.Service
func (b BigQuery) RemoveInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	// delete instance table
	table := b.InstancesDataset.Table(instanceTableName(i.GetStreamInstanceID()))
	err := table.Delete(ctx)
	if err != nil && !isNotFound(err) {
		return err
	}

	return nil
}

// Reset implements beneath.Service
func (b BigQuery) Reset(ctx context.Context) error {
	err := b.InstancesDataset.Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}
