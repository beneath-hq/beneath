package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"robpike.io/filter"

	"gitlab.com/beneath-org/beneath/engine/driver"
	"gitlab.com/beneath-org/beneath/pkg/log"
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
	err := b.createProject(ctx, p)
	if err != nil {
		if !isAlreadyExists(err) {
			return fmt.Errorf("error creating dataset for project '%s': %v", p.GetProjectName(), err)
		}
		return b.updateProject(ctx, p)
	}
	return nil
}

func (b BigQuery) createProject(ctx context.Context, p driver.Project) error {
	name := p.GetProjectName()
	return b.Client.Dataset(externalDatasetName(name)).Create(ctx, &bigquery.DatasetMetadata{
		Name: name,
		Labels: map[string]string{
			ProjectIDLabel: p.GetProjectID().String(),
		},
		Access: makeAccess(p.GetPublic(), nil),
	})
}

func (b BigQuery) updateProject(ctx context.Context, p driver.Project) error {
	name := p.GetProjectName()
	dataset := b.Client.Dataset(externalDatasetName(name))
	md, err := dataset.Metadata(ctx)
	if err != nil {
		return fmt.Errorf("error getting dataset to update '%s': %v", name, err)
	}

	_, err = dataset.Update(ctx, bigquery.DatasetMetadataToUpdate{
		Name:   name,
		Access: makeAccess(p.GetPublic(), md.Access),
	}, md.ETag)

	if err != nil && !isExpiredETag(err) {
		return fmt.Errorf("error updating dataset for project '%s': %v", name, err)
	}

	return nil
}

// RemoveProject implements beneath.Service
func (b BigQuery) RemoveProject(ctx context.Context, p driver.Project) error {
	name := p.GetProjectName()
	dataset := b.Client.Dataset(externalDatasetName(name))
	md, err := dataset.Metadata(ctx)
	if err != nil {
		return fmt.Errorf("error getting dataset to delete '%s': %v", name, err)
	}

	if md.Labels[ProjectIDLabel] != p.GetProjectID().String() {
		return fmt.Errorf("project ID label doesn't match project to delete for name '%s' and id '%s'", name, p.GetProjectID().String())
	}

	err = dataset.DeleteWithContents(ctx)
	if err != nil {
		return fmt.Errorf("error deleting dataset for project '%s': %v", name, err)
	}

	return nil
}

// RegisterInstance implements beneath.Service
func (b BigQuery) RegisterInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	// build schema object
	schema, err := bigquery.SchemaFromJSON([]byte(s.GetBigQuerySchema()))
	if err != nil {
		return err
	}

	// create external table
	dataset := b.Client.Dataset(externalDatasetName(p.GetProjectName()))
	table := dataset.Table(externalTableName(s.GetStreamName(), i.GetStreamInstanceID()))
	mdt, err := table.Metadata(ctx)
	if err != nil {
		if !isNotFound(err) {
			return err
		}

		// table doesn't exist -- create it and return
		// this is the expected scenario

		err = table.Create(ctx, &bigquery.TableMetadata{
			Schema: schema,
			TimePartitioning: &bigquery.TimePartitioning{
				Field: "__timestamp",
			},
			Clustering: &bigquery.Clustering{
				Fields: s.GetCodec().PrimaryIndex.GetFields(),
			},
			Labels: map[string]string{
				StreamIDLabel:   s.GetStreamID().String(),
				InstanceIDLabel: i.GetStreamInstanceID().String(),
			},
		})
		if err != nil && !isAlreadyExists(err) {
			return err
		}
		return nil
	}

	// table somehow exists -- RegisterInstance must be idempotent, so we'll execute an update

	// update
	_, err = table.Update(ctx, bigquery.TableMetadataToUpdate{
		Name:   s.GetStreamName(),
		Schema: schema,
	}, mdt.ETag)
	if err != nil {
		if isExpiredETag(err) {
			return nil
		}
		return err
	}

	// get view metadata (for ETag)
	view := dataset.Table(externalStreamViewName(s.GetStreamName()))
	mdv, err := view.Metadata(ctx)
	if err != nil {
		if isNotFound(err) {
			log.S.Infof("couldn't update instance '%s' because its view doesn't exist", i.GetStreamInstanceID().String())
			return nil
		}
		return err
	}

	// update view if same instance
	if mdv.Labels[InstanceIDLabel] == i.GetStreamInstanceID().String() {
		_, err = view.Update(ctx, bigquery.TableMetadataToUpdate{
			Name:   s.GetStreamName(),
			Schema: b.tableSchemaToViewSchema(schema),
		}, mdv.ETag)
		if err != nil {
			if isExpiredETag(err) {
				return nil
			}
			return err
		}
	}

	return nil
}

// PromoteInstance implements beneath.Service
func (b BigQuery) PromoteInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	// references
	dataset := b.Client.Dataset(externalDatasetName(p.GetProjectName()))
	underlying := dataset.Table(externalTableName(s.GetStreamName(), i.GetStreamInstanceID()))
	view := dataset.Table(externalStreamViewName(s.GetStreamName()))

	// view details
	viewQuery := fmt.Sprintf("select * except(`__key`, `__timestamp`) from `%s`", fullyQualifiedName(underlying))
	labels := map[string]string{
		StreamIDLabel:   s.GetStreamID().String(),
		InstanceIDLabel: i.GetStreamInstanceID().String(),
	}

	// check underlying exists
	md, err := underlying.Metadata(ctx)
	if err != nil {
		if isNotFound(err) {
			log.S.Infof("couldn't promote instance '%s' because its underlying table doesn't exist", i.GetStreamInstanceID().String())
			return nil
		}
		return err
	}

	// try to create view
	err = view.Create(ctx, &bigquery.TableMetadata{
		Name:      s.GetStreamName(),
		ViewQuery: viewQuery,
		Labels:    labels,
	})
	if err == nil {
		// trigger update to set view field descriptions (not possible with Create for stupid reasons)
		_, err = view.Update(ctx, bigquery.TableMetadataToUpdate{
			Schema: b.tableSchemaToViewSchema(md.Schema),
		}, "")
		if err != nil {
			return err
		}
	} else if err != nil {
		if !isAlreadyExists(err) {
			return err
		}

		// update view instead
		mdu := bigquery.TableMetadataToUpdate{
			Name:           s.GetStreamName(),
			ExpirationTime: bigquery.NeverExpire,
			ViewQuery:      viewQuery,
		}
		for l, v := range labels {
			mdu.SetLabel(l, v)
		}
		_, err = view.Update(ctx, mdu, "")
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveInstance implements beneath.Service
func (b BigQuery) RemoveInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	// delete external table
	dataset := b.Client.Dataset(externalDatasetName(p.GetProjectName()))
	table := dataset.Table(externalTableName(s.GetStreamName(), i.GetStreamInstanceID()))
	err := table.Delete(ctx)
	if err != nil && !isNotFound(err) {
		return err
	}

	// if view belongs to this instance, make it expire immediately
	// (cannot delete directly because delete doesn't accept etag)
	view := dataset.Table(externalStreamViewName(s.GetStreamName()))
	md, err := view.Metadata(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return err
	}
	if md.Labels[InstanceIDLabel] == i.GetStreamInstanceID().String() {
		_, err := view.Update(ctx, bigquery.TableMetadataToUpdate{
			ExpirationTime: time.Now(),
		}, md.ETag)
		if err != nil {
			return err
		}
	}

	return nil
}

// Reset implements beneath.Service
func (b BigQuery) Reset(ctx context.Context) error {
	datasets := b.Client.Datasets(ctx)
	for {
		ds, err := datasets.Next()
		if err == iterator.Done {
			break
		}

		err = ds.Delete(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BigQuery) tableSchemaToViewSchema(schema bigquery.Schema) bigquery.Schema {
	schemaI := filter.Choose(schema, func(f *bigquery.FieldSchema) bool {
		return f.Name != "__key" && f.Name != "__timestamp"
	})
	schema = schemaI.(bigquery.Schema)
	b.recursivelyMakeSchemaFieldsNullable(schema)
	return schema
}

func (b *BigQuery) recursivelyMakeSchemaFieldsNullable(schema bigquery.Schema) {
	for _, f := range schema {
		f.Required = false
		if f.Schema != nil {
			b.recursivelyMakeSchemaFieldsNullable(f.Schema)
		}
	}
}

func makeAccess(public bool, access []*bigquery.AccessEntry) []*bigquery.AccessEntry {
	if public {
		// check if exists
		found := false
		for _, a := range access {
			if a.Entity == "allAuthenticatedUsers" {
				found = true
				return access
			}
		}

		// add public auth if not exists
		if !found {
			access = append(access, &bigquery.AccessEntry{
				Role:       bigquery.ReaderRole,
				EntityType: bigquery.SpecialGroupEntity,
				Entity:     "allAuthenticatedUsers",
			})
		}
	} else {
		// remove all occurences
		filter.ChooseInPlace(&access, func(a *bigquery.AccessEntry) bool {
			return a.Entity != "allAuthenticatedUsers"
		})
	}

	return access
}
