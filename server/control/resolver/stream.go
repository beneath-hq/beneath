package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/server/control/gql"
	"github.com/beneath-hq/beneath/services/middleware"
)

// Table returns the gql.TableResolver
func (r *Resolver) Table() gql.TableResolver {
	return &tableResolver{r}
}

type tableResolver struct{ *Resolver }

func (r *tableResolver) TableID(ctx context.Context, obj *models.Table) (string, error) {
	return obj.TableID.String(), nil
}

func (r *queryResolver) TableByID(ctx context.Context, tableID uuid.UUID) (*models.Table, error) {
	table := r.Tables.FindTable(ctx, tableID)
	if table == nil {
		return nil, gqlerror.Errorf("Table with ID %s not found", tableID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, table.TableID, table.ProjectID, table.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read table with ID %s", tableID.String())
	}

	return tableWithPermissions(table, perms), nil
}

func (r *queryResolver) TableByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, tableName string) (*models.Table, error) {
	table := r.Tables.FindTableByOrganizationProjectAndName(ctx, organizationName, projectName, tableName)
	if table == nil {
		return nil, gqlerror.Errorf("Table %s/%s/%s not found", organizationName, projectName, tableName)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, table.TableID, table.ProjectID, table.Project.Public)
	if !perms.Read && !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to find table %s/%s/%s", organizationName, projectName, tableName)
	}

	return tableWithPermissions(table, perms), nil
}

func (r *queryResolver) TableInstanceByTableAndVersion(ctx context.Context, tableID uuid.UUID, version int) (*models.TableInstance, error) {
	instance := r.Tables.FindTableInstanceByVersion(ctx, tableID, version)
	if instance == nil {
		return nil, gqlerror.Errorf("Instance for table %s version %d not found", tableID.String(), version)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, tableID, instance.Table.ProjectID, instance.Table.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read instance version %d for table %s", version, tableID.String())
	}

	return instanceWithPermissions(instance, perms), nil
}

func (r *queryResolver) TableInstanceByOrganizationProjectTableAndVersion(ctx context.Context, organizationName string, projectName string, tableName string, version int) (*models.TableInstance, error) {
	instance := r.Tables.FindTableInstanceByOrganizationProjectTableAndVersion(ctx, organizationName, projectName, tableName, version)
	if instance == nil {
		return nil, gqlerror.Errorf("Table instance %s/%s/%s version %d not found", organizationName, projectName, tableName, version)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, instance.Table.TableID, instance.Table.ProjectID, instance.Table.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read table %s/%s/%s", organizationName, projectName, tableName)
	}

	return instanceWithPermissions(instance, perms), nil
}

func (r *queryResolver) TableInstancesForTable(ctx context.Context, tableID uuid.UUID) ([]*models.TableInstance, error) {
	table := r.Tables.FindTable(ctx, tableID)
	if table == nil {
		return nil, gqlerror.Errorf("Table with ID %s not found", tableID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, table.TableID, table.ProjectID, table.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read table with ID %s", tableID.String())
	}

	instances := r.Tables.FindTableInstances(ctx, tableID, nil, nil)
	return instances, nil
}

func (r *queryResolver) TableInstancesByOrganizationProjectAndTableName(ctx context.Context, organizationName string, projectName string, tableName string) ([]*models.TableInstance, error) {
	table := r.Tables.FindTableByOrganizationProjectAndName(ctx, organizationName, projectName, tableName)
	if table == nil {
		return nil, gqlerror.Errorf("Table %s/%s/%s not found", organizationName, projectName, tableName)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, table.TableID, table.ProjectID, table.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read table with ID %s", table.TableID.String())
	}

	instances := r.Tables.FindTableInstances(ctx, table.TableID, nil, nil)
	return instances, nil
}

func (r *queryResolver) TablesForUser(ctx context.Context, userID uuid.UUID) ([]*models.Table, error) {
	secret := middleware.GetSecret(ctx)
	if !(secret.IsUser() && secret.GetOwnerID() == userID) {
		return nil, gqlerror.Errorf("TablesForUser can only be called for the calling user")
	}
	return r.Tables.FindTablesForUser(ctx, userID), nil
}

func (r *queryResolver) CompileSchema(ctx context.Context, input gql.CompileSchemaInput) (*gql.CompileSchemaOutput, error) {
	table := &models.Table{}
	err := r.Tables.CompileToTable(table, input.SchemaKind, input.Schema, input.Indexes, nil)
	if err != nil {
		return nil, gqlerror.Errorf("Error compiling schema: %s", err.Error())
	}

	return &gql.CompileSchemaOutput{
		CanonicalAvroSchema: table.CanonicalAvroSchema,
		CanonicalIndexes:    table.CanonicalIndexes,
	}, nil
}

func (r *mutationResolver) CreateTable(ctx context.Context, input gql.CreateTableInput) (*models.Table, error) {
	// Handle UpdateIfExists (returns if exists)
	if input.UpdateIfExists != nil && *input.UpdateIfExists {
		table := r.Tables.FindTableByOrganizationProjectAndName(ctx, input.OrganizationName, input.ProjectName, input.TableName)
		if table != nil {
			return r.updateExistingFromCreateTable(ctx, table, input)
		}
	}

	project := r.Projects.FindProjectByOrganizationAndName(ctx, input.OrganizationName, input.ProjectName)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s/%s not found", input.OrganizationName, input.ProjectName)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, project.ProjectID, project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s/%s", input.OrganizationName, input.ProjectName)
	}

	table, err := r.Tables.CreateTable(ctx, &models.CreateTableCommand{
		Project:                   project,
		Name:                      input.TableName,
		SchemaKind:                input.SchemaKind,
		Schema:                    input.Schema,
		Indexes:                   input.Indexes,
		Description:               input.Description,
		Meta:                      input.Meta,
		AllowManualWrites:         input.AllowManualWrites,
		UseLog:                    input.UseLog,
		UseIndex:                  input.UseIndex,
		UseWarehouse:              input.UseWarehouse,
		LogRetentionSeconds:       input.LogRetentionSeconds,
		IndexRetentionSeconds:     input.IndexRetentionSeconds,
		WarehouseRetentionSeconds: input.WarehouseRetentionSeconds,
	})
	if err != nil {
		return nil, gqlerror.Errorf("Error creating table: %s", err.Error())
	}

	_, err = r.Tables.CreateTableInstance(ctx, table, nil, true)
	if err != nil {
		return nil, gqlerror.Errorf("Error creating first instance: %s", err.Error())
	}

	return tableWithProjectPermissions(table, perms), nil
}

func (r *mutationResolver) updateExistingFromCreateTable(ctx context.Context, table *models.Table, input gql.CreateTableInput) (*models.Table, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, table.ProjectID, table.Project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s", table.Project.Name)
	}

	// check use and retention unchanged (probably doesn't belong here, but tricky to move to service)
	if !checkUse(input.UseLog, table.UseLog) {
		return nil, gqlerror.Errorf("Cannot update useLog on existing table")
	}
	if !checkUse(input.UseIndex, table.UseIndex) {
		return nil, gqlerror.Errorf("Cannot update useIndex on existing table")
	}
	if !checkUse(input.UseWarehouse, table.UseWarehouse) {
		return nil, gqlerror.Errorf("Cannot update useWarehouse on existing table")
	}
	if !checkRetention(input.LogRetentionSeconds, table.LogRetentionSeconds) {
		return nil, gqlerror.Errorf("Cannot update logRetentionSeconds on existing table")
	}
	if !checkRetention(input.IndexRetentionSeconds, table.IndexRetentionSeconds) {
		return nil, gqlerror.Errorf("Cannot update indexRetentionSeconds on existing table")
	}
	if !checkRetention(input.WarehouseRetentionSeconds, table.WarehouseRetentionSeconds) {
		return nil, gqlerror.Errorf("Cannot update warehouseRetentionSeconds on existing table")
	}

	// attempt update
	err := r.Tables.UpdateTable(ctx, &models.UpdateTableCommand{
		Table:             table,
		SchemaKind:        &input.SchemaKind,
		Schema:            &input.Schema,
		Indexes:           input.Indexes,
		Description:       input.Description,
		Meta:              input.Meta,
		AllowManualWrites: input.AllowManualWrites,
	})
	if err != nil {
		return nil, gqlerror.Errorf("Error updating existing table: %s", err.Error())
	}

	return tableWithProjectPermissions(table, perms), nil
}

func checkUse(incoming *bool, current bool) bool {
	if incoming != nil {
		return *incoming == current
	}
	return current
}

func checkRetention(incoming *int, current int32) bool {
	if incoming != nil {
		return int32(*incoming) == current
	}
	return current == 0
}

func (r *mutationResolver) UpdateTable(ctx context.Context, input gql.UpdateTableInput) (*models.Table, error) {
	table := r.Tables.FindTable(ctx, input.TableID)
	if table == nil {
		return nil, gqlerror.Errorf("Table with ID %s not found", input.TableID)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, table.ProjectID, table.Project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s", table.Project.Name)
	}

	err := r.Tables.UpdateTable(ctx, &models.UpdateTableCommand{
		Table:             table,
		SchemaKind:        input.SchemaKind,
		Schema:            input.Schema,
		Indexes:           input.Indexes,
		Description:       input.Description,
		AllowManualWrites: input.AllowManualWrites,
	})
	if err != nil {
		return nil, gqlerror.Errorf("Error updating table: %s", err.Error())
	}

	return tableWithProjectPermissions(table, perms), nil
}

func (r *mutationResolver) DeleteTable(ctx context.Context, tableID uuid.UUID) (bool, error) {
	table := r.Tables.FindTable(ctx, tableID)
	if table == nil {
		return false, gqlerror.Errorf("Table %s not found", tableID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, table.ProjectID, false)
	if !perms.Create {
		return false, gqlerror.Errorf("Not allowed to perform admin functions in project %s", table.Project.Name)
	}

	err := r.Tables.DeleteTable(ctx, table)
	if err != nil {
		return false, gqlerror.Errorf("%s", err.Error())
	}

	return true, nil
}

// MaxInstancesPerTable sets a limit for the number of instances for a table at any given time
const MaxInstancesPerTable = 25

func (r *mutationResolver) CreateTableInstance(ctx context.Context, input gql.CreateTableInstanceInput) (*models.TableInstance, error) {
	if input.UpdateIfExists != nil && *input.UpdateIfExists {
		if input.Version == nil {
			return nil, gqlerror.Errorf("Cannot set updateIfExists=true without providing a version")
		}
		instance := r.Tables.FindTableInstanceByVersion(ctx, input.TableID, *input.Version)
		if instance != nil {
			return r.updateExistingFromCreateTableInstance(ctx, instance, input)
		}
	}

	table := r.Tables.FindTable(ctx, input.TableID)
	if table == nil {
		return nil, gqlerror.Errorf("Table %s not found", input.TableID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, table.TableID, table.ProjectID, table.Project.Public)
	if !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to write to table %s/%s/%s", table.Project.Organization.Name, table.Project.Name, table.Name)
	}

	// check not too many instances
	if table.InstancesCreatedCount-table.InstancesDeletedCount >= MaxInstancesPerTable {
		return nil, gqlerror.Errorf("You cannot have more than %d instances per table. Delete an existing instance to make room for more.", MaxInstancesPerTable)
	}

	makePrimary := false
	if input.MakePrimary != nil {
		makePrimary = *input.MakePrimary
	}

	si, err := r.Tables.CreateTableInstance(ctx, table, input.Version, makePrimary)
	if err != nil {
		return nil, gqlerror.Errorf("Error creating table instance: %s", err.Error())
	}

	return instanceWithPermissions(si, perms), nil
}

func (r *mutationResolver) updateExistingFromCreateTableInstance(ctx context.Context, instance *models.TableInstance, input gql.CreateTableInstanceInput) (*models.TableInstance, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, instance.Table.TableID, instance.Table.ProjectID, false)
	if !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to write to instance %s", instance.TableInstanceID)
	}

	makePrimary := false
	if input.MakePrimary != nil {
		makePrimary = *input.MakePrimary
	}

	err := r.Tables.UpdateTableInstance(ctx, instance.Table, instance, false, makePrimary)
	if err != nil {
		return nil, gqlerror.Errorf("Error updating table instance: %s", err.Error())
	}

	return instanceWithPermissions(instance, perms), nil
}

func (r *mutationResolver) UpdateTableInstance(ctx context.Context, input gql.UpdateTableInstanceInput) (*models.TableInstance, error) {
	instance := r.Tables.FindTableInstance(ctx, input.TableInstanceID)
	if instance == nil {
		return nil, gqlerror.Errorf("Table instance %s not found", input.TableInstanceID)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, instance.Table.TableID, instance.Table.ProjectID, false)
	if !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to write to instance %s", input.TableInstanceID)
	}

	makeFinal := false
	makePrimary := false
	if input.MakeFinal != nil {
		makeFinal = *input.MakeFinal
	}
	if input.MakePrimary != nil {
		makePrimary = *input.MakePrimary
	}

	err := r.Tables.UpdateTableInstance(ctx, instance.Table, instance, makeFinal, makePrimary)
	if err != nil {
		return nil, gqlerror.Errorf("Error updating table instance: %s", err.Error())
	}

	return instanceWithPermissions(instance, perms), nil
}

func (r *mutationResolver) DeleteTableInstance(ctx context.Context, instanceID uuid.UUID) (bool, error) {
	instance := r.Tables.FindTableInstance(ctx, instanceID)
	if instance == nil {
		return false, gqlerror.Errorf("Table instance '%s' not found", instanceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.TablePermissionsForSecret(ctx, secret, instance.Table.TableID, instance.Table.ProjectID, false)
	if !perms.Write {
		return false, gqlerror.Errorf("Not allowed to write to instance %s", instanceID.String())
	}

	err := r.Tables.DeleteTableInstance(ctx, instance.Table, instance)
	if err != nil {
		return false, gqlerror.Errorf("Couldn't delete table instance: %s", err.Error())
	}

	return true, nil
}

func instanceWithPermissions(si *models.TableInstance, perms models.TablePermissions) *models.TableInstance {
	if si.Table != nil {
		si.Table = tableWithPermissions(si.Table, perms)
	}
	return si
}

func tableWithPermissions(s *models.Table, perms models.TablePermissions) *models.Table {
	// Note: This is a pretty bad approximation, but will do for our current use case
	return tableWithProjectPermissions(s, models.ProjectPermissions{
		View:   perms.Read,
		Create: perms.Write,
		Admin:  perms.Write,
	})
}

func tableWithProjectPermissions(s *models.Table, perms models.ProjectPermissions) *models.Table {
	if s.Project != nil {
		s.Project.Permissions = &models.PermissionsUsersProjects{
			View:   perms.View,
			Create: perms.Create,
			Admin:  perms.Admin,
		}
	}
	return s
}
