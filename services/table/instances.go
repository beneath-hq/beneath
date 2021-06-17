package table

import (
	"context"
	"fmt"
	"time"

	"github.com/beneath-hq/beneath/infra/db"
	"github.com/beneath-hq/beneath/models"
	uuid "github.com/satori/go.uuid"
)

// FindTableInstance finds an instance and related table details
func (s *Service) FindTableInstance(ctx context.Context, instanceID uuid.UUID) *models.TableInstance {
	si := &models.TableInstance{TableInstanceID: instanceID}
	err := s.DB.GetDB(ctx).ModelContext(ctx, si).Column(
		"table_instance.*",
		"Table",
		"Table.TableIndexes",
		"Table.PrimaryTableInstance",
		"Table.Project",
		"Table.Project.Organization",
	).WherePK().Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return si
}

// FindTableInstanceByOrganizationProjectTableAndVersion finds an instance and related table details
func (s *Service) FindTableInstanceByOrganizationProjectTableAndVersion(ctx context.Context, organizationName string, projectName string, tableName string, version int) *models.TableInstance {
	si := &models.TableInstance{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, si).
		Column(
			"table_instance.*",
			"Table",
			"Table.TableIndexes",
			"Table.PrimaryTableInstance",
			"Table.Project",
			"Table.Project.Organization",
		).
		Where("lower(table__project__organization.name) = lower(?)", organizationName).
		Where("lower(table__project.name) = lower(?)", projectName).
		Where("lower(table.name) = lower(?)", tableName).
		Where("table_instance.version = ?", version).
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return si
}

// FindTableInstanceByVersion finds an instance by version in its parent table
func (s *Service) FindTableInstanceByVersion(ctx context.Context, tableID uuid.UUID, version int) *models.TableInstance {
	si := &models.TableInstance{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, si).
		Column(
			"table_instance.*",
			"Table",
			"Table.TableIndexes",
			"Table.PrimaryTableInstance",
			"Table.Project",
			"Table.Project.Organization",
		).
		Where("table_instance.table_id = ?", tableID).
		Where("table_instance.version = ?", version).
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return si
}

// FindTableInstances finds the instances associated with the given table
func (s *Service) FindTableInstances(ctx context.Context, tableID uuid.UUID, fromVersion *int, toVersion *int) []*models.TableInstance {
	var res []*models.TableInstance

	// optimization
	if fromVersion != nil && toVersion != nil && *fromVersion == *toVersion {
		return res
	}

	// build query
	query := s.DB.GetDB(ctx).ModelContext(ctx, &res).Where("table_id = ?", tableID)
	if fromVersion != nil {
		query = query.Where("version >= ?", *fromVersion)
	}
	if toVersion != nil {
		query = query.Where("version < ?", *toVersion)
	}

	// select
	err := query.Order("version DESC").Select()
	if err != nil {
		panic(fmt.Errorf("error fetching table instances: %s", err.Error()))
	}

	return res
}

// CreateTableInstance creates and registers a new table instance
func (s *Service) CreateTableInstance(ctx context.Context, table *models.Table, version *int, makePrimary bool) (*models.TableInstance, error) {
	if version == nil {
		version = &table.NextInstanceVersion
	}

	if *version < 0 {
		return nil, fmt.Errorf("Cannot create table instance with negative version (got %d)", version)
	}

	instance := &models.TableInstance{
		TableID: table.TableID,
		Table:   table,
		Version: *version,
	}

	if makePrimary {
		now := time.Now()
		instance.MadePrimaryOn = &now
	}

	// create in transaction
	err := s.DB.InTransaction(ctx, func(ctx context.Context) error {
		tx := s.DB.GetDB(ctx)

		_, err := tx.Model(instance).Insert()
		if err != nil {
			return err
		}

		tableQuery := tx.Model(table).WherePK()
		tableQuery.Set("next_instance_version = greatest(next_instance_version, ? + 1)", instance.Version)
		tableQuery.Set("instances_created_count = instances_created_count + 1")
		if makePrimary {
			tableQuery.Set("primary_table_instance_id = ?", instance.TableInstanceID)
			tableQuery.Set("instances_made_primary_count = instances_made_primary_count + 1")
		}
		_, err = tableQuery.Update()
		if err != nil {
			return err
		}

		// modify table to reflect changes without refetching
		if instance.Version >= table.NextInstanceVersion {
			table.NextInstanceVersion = instance.Version + 1
		}
		table.InstancesCreatedCount++
		table.InstancesMadePrimaryCount++
		table.PrimaryTableInstance = instance
		table.PrimaryTableInstanceID = &instance.TableInstanceID

		return nil
	})
	if err != nil {
		return nil, err
	}

	// publish event
	err = s.Bus.Publish(ctx, &models.TableInstanceCreatedEvent{
		Table:         table,
		TableInstance: instance,
		MakePrimary:   makePrimary,
	})
	if err != nil {
		return nil, err
	}

	return instance, nil
}

// UpdateTableInstance makes a table instance primary and/or final
func (s *Service) UpdateTableInstance(ctx context.Context, table *models.Table, instance *models.TableInstance, makeFinal bool, makePrimary bool) error {
	// makeFinal is one-way and can be done once
	// makePrimary can be done any number of times

	// bail if there's nothing to do
	if (!makeFinal || instance.MadeFinalOn != nil) && !makePrimary {
		return nil
	}

	// run in transaction
	now := time.Now()
	err := s.DB.InTransaction(ctx, func(ctx context.Context) error {
		tx := s.DB.GetDB(ctx)
		updateTableQuery := tx.Model(table).WherePK()
		updateInstanceQuery := tx.Model(instance).WherePK()

		// handle makeFinal
		if makeFinal && instance.MadeFinalOn == nil {
			updateTableQuery.Set("instances_made_final_count = instances_made_final_count + 1")
			updateInstanceQuery.Set("made_final_on = now()")
			instance.MadeFinalOn = &now
		}

		// handle makePrimary
		if makePrimary {
			updateInstanceQuery.Set("made_primary_on = now()")
			instance.MadePrimaryOn = &now
			updateTableQuery.Set("primary_table_instance_id = ?", instance.TableInstanceID)
			if instance.MadePrimaryOn == nil {
				updateTableQuery.Set("instances_made_primary_count = instances_made_primary_count + 1")
			}
		}

		// updated on timestamps
		updateTableQuery.Set("updated_on = now()")
		updateInstanceQuery.Set("updated_on = now()")
		instance.UpdatedOn = now

		// update
		_, err := updateTableQuery.Update()
		if err != nil {
			return err
		}
		_, err = updateInstanceQuery.Update()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// publish event
	err = s.Bus.Publish(ctx, &models.TableInstanceUpdatedEvent{
		Table:            table,
		TableInstance:    instance,
		MakeFinal:        makeFinal,
		MakePrimary:      makePrimary,
		OrganizationName: table.Project.Organization.Name,
		ProjectName:      table.Project.Name,
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteTableInstance deletes a table and its data
func (s *Service) DeleteTableInstance(ctx context.Context, table *models.Table, instance *models.TableInstance) error {
	err := s.DB.InTransaction(ctx, func(ctx context.Context) error {
		tx := s.DB.GetDB(ctx)

		err := tx.Delete(instance)
		if err != nil {
			return err
		}

		_, err = tx.Model(table).
			WherePK().
			Set("updated_on = ?", table.UpdatedOn).
			Set("instances_deleted_count = instances_deleted_count + 1").
			Update()
		if err != nil {
			return err
		}

		// table.PrimaryTableInstanceID has SET NULL constraint (so we don't have to handle it)

		return nil
	})
	if err != nil {
		return err
	}

	// publish event
	err = s.Bus.Publish(ctx, &models.TableInstanceDeletedEvent{
		Table:         table,
		TableInstance: instance,
	})
	if err != nil {
		return err
	}

	return nil
}
