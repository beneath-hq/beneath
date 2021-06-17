package table

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/beneath-hq/beneath/infra/db"
	"github.com/beneath-hq/beneath/models"
	uuid "github.com/satori/go.uuid"
)

// FindTable finds a table
func (s *Service) FindTable(ctx context.Context, tableID uuid.UUID) *models.Table {
	table := &models.Table{TableID: tableID}
	err := s.DB.GetDB(ctx).ModelContext(ctx, table).
		WherePK().
		Column(
			"table.*",
			"Project",
			"Project.Organization",
			"PrimaryTableInstance",
			"TableIndexes",
		).
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return table
}

// FindTableByOrganizationProjectAndName finds a table
func (s *Service) FindTableByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, tableName string) *models.Table {
	table := &models.Table{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, table).
		Column(
			"table.*",
			"Project",
			"Project.Organization",
			"PrimaryTableInstance",
			"TableIndexes",
		).
		Where("lower(project__organization.name) = lower(?)", organizationName).
		Where("lower(project.name) = lower(?)", projectName).
		Where("lower(table.name) = lower(?)", tableName).
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return table
}

// FindTablesForUser finds the tables that the user has access to
func (s *Service) FindTablesForUser(ctx context.Context, userID uuid.UUID) []*models.Table {
	var tables []*models.Table
	err := s.DB.GetDB(ctx).ModelContext(ctx, &tables).
		Column("table.*", "Project.project_id", "Project.name", "Project.Organization.organization_id", "Project.Organization.name").
		Join("JOIN permissions_users_projects AS pup ON pup.project_id = table.project_id").
		Where("pup.user_id = ?", userID).
		Order("project__organization.name", "project.name", "table.name").
		Limit(200).
		Select()
	if err != nil {
		panic(err)
	}
	return tables
}

// CreateTable compiles and creates a new table
func (s *Service) CreateTable(ctx context.Context, msg *models.CreateTableCommand) (*models.Table, error) {
	table := &models.Table{
		Name:      msg.Name,
		Project:   msg.Project,
		ProjectID: msg.Project.ProjectID,
	}

	// compile table
	err := s.CompileToTable(table, msg.SchemaKind, msg.Schema, msg.Indexes, msg.Description)
	if err != nil {
		return nil, err
	}
	table.SchemaMD5 = s.ComputeSchemaMD5(msg.Schema, msg.Indexes)

	// set other values
	table.Meta = derefBool(msg.Meta, false)
	table.AllowManualWrites = derefBool(msg.AllowManualWrites, false)
	table.UseLog = derefBool(msg.UseLog, true)
	table.UseIndex = derefBool(msg.UseIndex, true)
	table.UseWarehouse = derefBool(msg.UseWarehouse, true)

	if msg.LogRetentionSeconds != nil {
		if !table.UseLog {
			return nil, fmt.Errorf("Cannot set logRetentionSeconds on table that doesn't have useLog=true")
		}
		table.LogRetentionSeconds = int32(*msg.LogRetentionSeconds)
	}

	if msg.IndexRetentionSeconds != nil {
		if !table.UseIndex {
			return nil, fmt.Errorf("Cannot set indexRetentionSeconds on table that doesn't have useIndex=true")
		}
		table.IndexRetentionSeconds = int32(*msg.IndexRetentionSeconds)
	}

	if msg.WarehouseRetentionSeconds != nil {
		if !table.UseWarehouse {
			return nil, fmt.Errorf("Cannot set warehouseRetentionSeconds on table that doesn't have useWarehouse=true")
		}
		table.WarehouseRetentionSeconds = int32(*msg.WarehouseRetentionSeconds)
	}

	// if there's a normalized index, make sure we use a non-expiring log.
	// note that this is currently a little fictional, as normalized index support is pretty shaky.
	for _, index := range table.TableIndexes {
		if index.Normalize && (!table.UseLog || table.LogRetentionSeconds != 0) {
			return nil, fmt.Errorf("Cannot use normalized indexes when useLog=false or logRetentionSeconds is set")
		}
	}

	// temporarily, we're going to enforce UseLog
	if !table.UseLog {
		return nil, fmt.Errorf("Currently doesn't support tables with useLog=false")
	}

	// validate
	err = table.Validate()
	if err != nil {
		return nil, err
	}

	// insert table and its indexes transactionally
	err = s.DB.InTransaction(ctx, func(ctx context.Context) error {
		tx := s.DB.GetDB(ctx)

		_, err := tx.Model(table).Insert()
		if err != nil {
			return err
		}

		for _, index := range table.TableIndexes {
			index.TableID = table.TableID
			_, err := tx.Model(index).Insert()
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// publish events
	err = s.Bus.Publish(ctx, &models.TableCreatedEvent{
		Table: table,
	})
	if err != nil {
		return nil, err
	}

	return table, nil
}

// UpdateTable updates a tables schema (within limits) and more
func (s *Service) UpdateTable(ctx context.Context, msg *models.UpdateTableCommand) error {
	table := msg.Table
	updateFields := []string{}
	modifiedSchema := false

	// re-compile and set updates if schema changed
	if msg.SchemaKind != nil && msg.Schema != nil {
		schemaMD5 := s.ComputeSchemaMD5(*msg.Schema, msg.Indexes)
		if !bytes.Equal(schemaMD5, table.SchemaMD5) {
			err := s.CompileToTable(table, *msg.SchemaKind, *msg.Schema, msg.Indexes, msg.Description)
			if err != nil {
				return err
			}
			table.SchemaMD5 = schemaMD5
			updateFields = append(
				updateFields,
				"schema_kind",
				"schema",
				"schema_md5",
				"avro_schema",
				"canonical_avro_schema",
				"canonical_indexes",
				"description",
			)
			modifiedSchema = true
		}
	}

	// update description (you can change a description without changing the schema)
	// (this steps on the former clause's toes, but the if-check here makes them mutually exclusive)
	if msg.Description != nil && *msg.Description != table.Description {
		table.Description = *msg.Description
		updateFields = append(updateFields, "description")
	}

	// update meta
	if msg.Meta != nil && *msg.Meta != table.Meta {
		table.Meta = *msg.Meta
		updateFields = append(updateFields, "meta")
	}

	// update manual writes
	if msg.AllowManualWrites != nil && *msg.AllowManualWrites != table.AllowManualWrites {
		table.AllowManualWrites = *msg.AllowManualWrites
		updateFields = append(updateFields, "allow_manual_writes")
	}

	// quit if no changes
	if len(updateFields) == 0 {
		return nil
	}

	// validate
	err := table.Validate()
	if err != nil {
		return err
	}

	// update
	table.UpdatedOn = time.Now()
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, table).WherePK().Update()
	if err != nil {
		return err
	}

	// publish events
	err = s.Bus.Publish(ctx, &models.TableUpdatedEvent{
		Table:          table,
		ModifiedSchema: modifiedSchema,
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteTable deletes a table and all its related instances
func (s *Service) DeleteTable(ctx context.Context, table *models.Table) error {
	// delete instances
	instances := s.FindTableInstances(ctx, table.TableID, nil, nil)
	for _, inst := range instances {
		err := s.DeleteTableInstance(ctx, table, inst)
		if err != nil {
			return err
		}
	}

	// delete table
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, table).WherePK().Delete()
	if err != nil {
		return err
	}

	// publish event
	err = s.Bus.Publish(ctx, &models.TableDeletedEvent{TableID: table.TableID})
	if err != nil {
		return err
	}

	return nil
}
