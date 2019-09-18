package bigquery

import (
	"context"
	"fmt"
	"time"

	"github.com/beneath-core/beneath-go/core/log"

	bq "cloud.google.com/go/bigquery"
	uuid "github.com/satori/go.uuid"
	"robpike.io/filter"
)

// RegisterProject  implements engine.WarehouseDriver
func (b *BigQuery) RegisterProject(ctx context.Context, projectID uuid.UUID, public bool, name, displayName, description string) error {
	err := b.Client.Dataset(externalDatasetName(name)).Create(ctx, &bq.DatasetMetadata{
		Name:        displayName,
		Description: description,
		Labels: map[string]string{
			ProjectIDLabel: projectID.String(),
		},
		Access: makeAccess(public, nil),
	})

	if err != nil {
		if isAlreadyExists(err) {
			return b.UpdateProject(ctx, projectID, public, name, displayName, description)
		}
		return fmt.Errorf("error creating dataset for project '%s': %v", name, err)
	}

	return nil
}

// UpdateProject implements engine.WarehouseDriver
func (b *BigQuery) UpdateProject(ctx context.Context, projectID uuid.UUID, public bool, name, displayName, description string) error {
	dataset := b.Client.Dataset(externalDatasetName(name))
	md, err := dataset.Metadata(ctx)
	if err != nil {
		return fmt.Errorf("error getting dataset to update '%s': %v", name, err)
	}

	_, err = dataset.Update(ctx, bq.DatasetMetadataToUpdate{
		Name:        displayName,
		Description: description,
		Access:      makeAccess(public, md.Access),
	}, md.ETag)

	if err != nil && !isExpiredETag(err) {
		return fmt.Errorf("error updating dataset for project '%s': %v", name, err)
	}

	return nil
}

// DeregisterProject implements engine.WarehouseDriver
func (b *BigQuery) DeregisterProject(ctx context.Context, projectID uuid.UUID, name string) error {
	dataset := b.Client.Dataset(externalDatasetName(name))
	md, err := dataset.Metadata(ctx)
	if err != nil {
		return fmt.Errorf("error getting dataset to delete '%s': %v", name, err)
	}

	if md.Labels[ProjectIDLabel] != projectID.String() {
		return fmt.Errorf("project ID label doesn't match project to delete for name '%s' and id '%s'", name, projectID.String())
	}

	err = dataset.DeleteWithContents(ctx)
	if err != nil {
		return fmt.Errorf("error deleting dataset for project '%s': %v", name, err)
	}

	return nil
}

// RegisterStreamInstance implements engine.WarehouseDriver
func (b *BigQuery) RegisterStreamInstance(ctx context.Context, projectName string, streamID uuid.UUID, streamName string, streamDescription string, schemaJSON string, keyFields []string, instanceID uuid.UUID) error {
	// build schema object
	schema, err := bq.SchemaFromJSON([]byte(schemaJSON))
	if err != nil {
		return err
	}

	// create internal table
	table := b.Client.Dataset(internalDatasetName()).Table(internalTableName(instanceID))
	err = table.Create(ctx, &bq.TableMetadata{
		Schema: internalRowSchema,
		TimePartitioning: &bq.TimePartitioning{
			Field: "timestamp",
		},
	})
	if err != nil && !isAlreadyExists(err) {
		return err
	}

	// create external table
	dataset := b.Client.Dataset(externalDatasetName(projectName))
	table = dataset.Table(externalTableName(streamName, instanceID))
	err = table.Create(ctx, &bq.TableMetadata{
		Description: streamDescription,
		Schema:      schema,
		TimePartitioning: &bq.TimePartitioning{
			Field: "__timestamp",
		},
		Clustering: &bq.Clustering{
			Fields: keyFields,
		},
		Labels: map[string]string{
			StreamIDLabel:   streamID.String(),
			InstanceIDLabel: instanceID.String(),
		},
	})
	if err != nil && !isAlreadyExists(err) {
		return err
	}

	return nil
}

// PromoteStreamInstance implements engine.WarehouseDriver
func (b *BigQuery) PromoteStreamInstance(ctx context.Context, projectName string, streamID uuid.UUID, streamName string, streamDescription string, instanceID uuid.UUID) error {
	// references
	dataset := b.Client.Dataset(externalDatasetName(projectName))
	underlying := dataset.Table(externalTableName(streamName, instanceID))
	view := dataset.Table(externalStreamViewName(streamName))

	// view details
	viewQuery := fmt.Sprintf("select * except(`__key`, `__timestamp`) from `%s`", fullyQualifiedName(underlying))
	labels := map[string]string{
		StreamIDLabel:   streamID.String(),
		InstanceIDLabel: instanceID.String(),
	}

	// check underlying exists
	_, err := underlying.Metadata(ctx)
	if err != nil {
		if isNotFound(err) {
			log.S.Infof("couldn't promote instance '%s' because its underlying table doesn't exist", instanceID.String())
			return nil
		}
		return err
	}

	// try to create view
	err = view.Create(ctx, &bq.TableMetadata{
		Name:        streamName,
		Description: streamDescription,
		ViewQuery:   viewQuery,
		Labels:      labels,
	})
	if err != nil {
		if !isAlreadyExists(err) {
			return err
		}

		// update view instead
		mdu := bq.TableMetadataToUpdate{
			Name:           streamName,
			Description:    streamDescription,
			ExpirationTime: bq.NeverExpire,
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

// UpdateStreamInstance implements engine.WarehouseDriver
func (b *BigQuery) UpdateStreamInstance(ctx context.Context, projectName string, streamName string, streamDescription string, schemaJSON string, instanceID uuid.UUID) error {
	// build schema object
	schema, err := bq.SchemaFromJSON([]byte(schemaJSON))
	if err != nil {
		return err
	}

	// get current metadata (for Etag)
	dataset := b.Client.Dataset(externalDatasetName(projectName))
	table := dataset.Table(externalTableName(streamName, instanceID))
	md, err := table.Metadata(ctx)
	if err != nil {
		if isNotFound(err) {
			log.S.Infof("couldn't update instance '%s' because its table doesn't exist", instanceID.String())
			return nil
		}
		return err
	}

	// update
	_, err = table.Update(ctx, bq.TableMetadataToUpdate{
		Name:        streamName,
		Description: streamDescription,
		Schema:      schema,
	}, md.ETag)
	if err != nil {
		if isExpiredETag(err) {
			return nil
		}
		return err
	}

	// get view metadata (for ETag)
	view := dataset.Table(externalStreamViewName(streamName))
	md, err = view.Metadata(ctx)
	if err != nil {
		if isNotFound(err) {
			log.S.Infof("couldn't update instance '%s' because its view doesn't exist", instanceID.String())
			return nil
		}
		return err
	}

	// update view if same instance
	if md.Labels[InstanceIDLabel] == instanceID.String() {
		_, err = view.Update(ctx, bq.TableMetadataToUpdate{
			Name:        streamName,
			Description: streamDescription,
		}, md.ETag)
		if err != nil {
			if isExpiredETag(err) {
				return nil
			}
			return err
		}
	}

	return nil
}

// DeregisterStreamInstance implements engine.WarehouseDriver
func (b *BigQuery) DeregisterStreamInstance(ctx context.Context, projectID uuid.UUID, projectName string, streamID uuid.UUID, streamName string, instanceID uuid.UUID) error {
	// delete internal table
	table := b.Client.Dataset(internalDatasetName()).Table(internalTableName(instanceID))
	err := table.Delete(ctx)
	if err != nil && !isNotFound(err) {
		return err
	}

	// delete external table
	dataset := b.Client.Dataset(externalDatasetName(projectName))
	table = dataset.Table(externalTableName(streamName, instanceID))
	err = table.Delete(ctx)
	if err != nil && !isNotFound(err) {
		return err
	}

	// if view belongs to this instance, make it expire immediately
	// (cannot delete directly because delete doesn't accept etag)
	view := dataset.Table(externalStreamViewName(streamName))
	md, err := view.Metadata(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return err
	}
	if md.Labels[InstanceIDLabel] == instanceID.String() {
		_, err := view.Update(ctx, bq.TableMetadataToUpdate{
			ExpirationTime: time.Now(),
		}, md.ETag)
		if err != nil {
			return err
		}
	}

	return nil
}

func makeAccess(public bool, access []*bq.AccessEntry) []*bq.AccessEntry {
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
			access = append(access, &bq.AccessEntry{
				Role:       bq.ReaderRole,
				EntityType: bq.SpecialGroupEntity,
				Entity:     "allAuthenticatedUsers",
			})
		}
	} else {
		// remove all occurences
		filter.ChooseInPlace(&access, func(a *bq.AccessEntry) bool {
			return a.Entity != "allAuthenticatedUsers"
		})
	}

	return access
}
