package resolver

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/server/control/gql"
	"github.com/beneath-hq/beneath/services/middleware"
)

// Service returns gql.ServiceResolver implementation.
func (r *Resolver) Service() gql.ServiceResolver { return &serviceResolver{r} }

type serviceResolver struct{ *Resolver }

func (r *serviceResolver) ServiceID(ctx context.Context, obj *models.Service) (string, error) {
	return obj.ServiceID.String(), nil
}

func (r *serviceResolver) QuotaStartTime(ctx context.Context, obj *models.Service) (*time.Time, error) {
	t := r.Usage.GetQuotaPeriod(obj.QuotaEpoch).Floor(time.Now())
	return &t, nil
}

func (r *serviceResolver) QuotaEndTime(ctx context.Context, obj *models.Service) (*time.Time, error) {
	t := r.Usage.GetQuotaPeriod(obj.QuotaEpoch).Next(time.Now())
	return &t, nil
}

func (r *queryResolver) ServiceByID(ctx context.Context, serviceID uuid.UUID) (*models.Service, error) {
	service := r.Services.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, service.Project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("You are not allowed to view project resources")
	}

	return serviceWithProjectPermissions(service, perms), nil
}

func (r *queryResolver) ServiceByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, serviceName string) (*models.Service, error) {
	service := r.Services.FindServiceByOrganizationProjectAndName(ctx, organizationName, projectName, serviceName)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s/%s/%s not found", organizationName, projectName, serviceName)
	}

	secret := middleware.GetSecret(ctx)
	selfFind := secret.IsService() && secret.GetOwnerID() == service.ServiceID
	var perms models.ProjectPermissions
	if selfFind {
		perms = models.ProjectPermissions{View: true}
	} else {
		perms = r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, service.Project.Public)
	}

	if !perms.View {
		return nil, gqlerror.Errorf("You are not allowed to view project resources")
	}

	return serviceWithProjectPermissions(service, perms), nil
}
func (r *queryResolver) TablePermissionsForService(ctx context.Context, serviceID uuid.UUID) ([]*models.PermissionsServicesTables, error) {
	service := r.Services.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	projectPerms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, service.Project.Public)
	if !projectPerms.View {
		return nil, gqlerror.Errorf("You are not allowed to view project resources")
	}

	servicePerms := r.Services.FindTablePermissionsForService(ctx, serviceID)

	return servicePerms, nil
}

func (r *mutationResolver) CreateService(ctx context.Context, input gql.CreateServiceInput) (*models.Service, error) {
	// Handle UpdateIfExists (returns if exists)
	if input.UpdateIfExists != nil && *input.UpdateIfExists {
		service := r.Services.FindServiceByOrganizationProjectAndName(ctx, input.OrganizationName, input.ProjectName, input.ServiceName)
		if service != nil {
			return r.updateExistingFromCreateService(ctx, service, input)
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

	service := &models.Service{
		Name:       input.ServiceName,
		Project:    project,
		ProjectID:  project.ProjectID,
		QuotaEpoch: time.Now(),
	}

	err := r.Services.Create(ctx, service, input.Description, input.SourceURL, IntToInt64(input.ReadQuota), IntToInt64(input.WriteQuota), IntToInt64(input.ScanQuota))
	if err != nil {
		return nil, gqlerror.Errorf("Error creating service: %s", err.Error())
	}

	return serviceWithProjectPermissions(service, perms), nil
}

func (r *mutationResolver) updateExistingFromCreateService(ctx context.Context, service *models.Service, input gql.CreateServiceInput) (*models.Service, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, service.Project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s/%s", input.OrganizationName, input.ProjectName)
	}

	err := r.Services.Update(ctx, service, input.Description, input.SourceURL, IntToInt64(input.ReadQuota), IntToInt64(input.WriteQuota), IntToInt64(input.ScanQuota))
	if err != nil {
		return nil, gqlerror.Errorf("Error updating service: %s", err.Error())
	}

	return serviceWithProjectPermissions(service, perms), nil
}

func (r *mutationResolver) UpdateService(ctx context.Context, input gql.UpdateServiceInput) (*models.Service, error) {
	service := r.Services.FindServiceByOrganizationProjectAndName(ctx, input.OrganizationName, input.ProjectName, input.ServiceName)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s/%s/%s not found", input.OrganizationName, input.ProjectName, input.ServiceName)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, service.Project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s/%s", input.OrganizationName, input.ProjectName)
	}

	err := r.Services.Update(ctx, service, input.Description, input.SourceURL, IntToInt64(input.ReadQuota), IntToInt64(input.WriteQuota), IntToInt64(input.ScanQuota))
	if err != nil {
		return nil, gqlerror.Errorf("Error updating service: %s", err.Error())
	}

	return serviceWithProjectPermissions(service, perms), nil
}

func (r *mutationResolver) UpdateServiceTablePermissions(ctx context.Context, serviceID uuid.UUID, tableID uuid.UUID, read *bool, write *bool) (*models.PermissionsServicesTables, error) {
	service := r.Services.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	projPerms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, false)
	if !projPerms.Create {
		return nil, gqlerror.Errorf("Not allowed to edit the service")
	}

	table := r.Tables.FindTable(ctx, tableID)
	if table == nil {
		return nil, gqlerror.Errorf("Table %s not found", tableID.String())
	}

	tableProjectPerms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, table.ProjectID, false)
	if !tableProjectPerms.Create {
		return nil, gqlerror.Errorf("Not allowed to access table")
	}

	pss := r.Permissions.FindPermissionsServicesTables(ctx, serviceID, tableID)
	if pss == nil {
		pss = &models.PermissionsServicesTables{
			ServiceID: serviceID,
			TableID:   tableID,
		}
	}

	err := r.Permissions.UpdateServiceTablePermission(ctx, pss, read, write)
	if err != nil {
		return nil, err
	}

	pss.Table = table

	return pss, nil
}

func (r *mutationResolver) DeleteService(ctx context.Context, serviceID uuid.UUID) (bool, error) {
	service := r.Services.FindService(ctx, serviceID)
	if service == nil {
		return false, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, false)
	if !perms.Create {
		return false, gqlerror.Errorf("Not allowed to edit project resources")
	}

	err := r.Services.Delete(ctx, service)
	if err != nil {
		return false, gqlerror.Errorf("Failed deleting service: %s", err.Error())
	}

	return true, nil
}

func serviceWithProjectPermissions(s *models.Service, perms models.ProjectPermissions) *models.Service {
	if s.Project != nil {
		s.Project.Permissions = &models.PermissionsUsersProjects{
			View:   perms.View,
			Create: perms.Create,
			Admin:  perms.Admin,
		}
	}
	return s
}
