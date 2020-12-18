package resolver

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/server/control/gql"
	"gitlab.com/beneath-hq/beneath/services/middleware"
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

	return service, nil
}

func (r *queryResolver) ServiceByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, serviceName string) (*models.Service, error) {
	service := r.Services.FindServiceByOrganizationProjectAndName(ctx, organizationName, projectName, serviceName)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s/%s/%s not found", organizationName, projectName, serviceName)
	}

	secret := middleware.GetSecret(ctx)
	selfFind := secret.IsService() && secret.GetOwnerID() == service.ServiceID
	if !selfFind {
		perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, service.Project.Public)
		if !perms.View {
			return nil, gqlerror.Errorf("You are not allowed to view organization resources")
		}
	}

	return service, nil
}
func (r *queryResolver) StreamPermissionsForService(ctx context.Context, serviceID uuid.UUID) ([]*models.PermissionsServicesStreams, error) {
	service := r.Services.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	projectPerms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, service.Project.Public)
	if !projectPerms.View {
		return nil, gqlerror.Errorf("You are not allowed to view project resources")
	}

	servicePerms := r.Services.FindStreamPermissionsForService(ctx, serviceID)

	return servicePerms, nil
}

func (r *mutationResolver) CreateService(ctx context.Context, input gql.CreateServiceInput) (*models.Service, error) {
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

	return service, nil
}

func (r *mutationResolver) UpdateService(ctx context.Context, input gql.UpdateServiceInput) (*models.Service, error) {
	project := r.Projects.FindProjectByOrganizationAndName(ctx, input.OrganizationName, input.ProjectName)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s/%s not found", input.OrganizationName, input.ProjectName)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, project.ProjectID, project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s/%s", input.OrganizationName, input.ProjectName)
	}

	service := r.Services.FindServiceByOrganizationProjectAndName(ctx, input.OrganizationName, input.ProjectName, input.ServiceName)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s/%s/%s not found", input.OrganizationName, input.ProjectName, input.ServiceName)
	}

	err := r.Services.Update(ctx, service, input.Description, input.SourceURL, IntToInt64(input.ReadQuota), IntToInt64(input.WriteQuota), IntToInt64(input.ScanQuota))
	if err != nil {
		return nil, gqlerror.Errorf("Error updating service: %s", err.Error())
	}

	return service, nil
}

func (r *mutationResolver) UpdateServiceStreamPermissions(ctx context.Context, serviceID uuid.UUID, streamID uuid.UUID, read *bool, write *bool) (*models.PermissionsServicesStreams, error) {
	service := r.Services.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	projPerms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, false)
	if !projPerms.Create {
		return nil, gqlerror.Errorf("Not allowed to edit the service")
	}

	stream := r.Streams.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	streamProjectPerms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, stream.ProjectID, false)
	if !streamProjectPerms.Create {
		return nil, gqlerror.Errorf("Not allowed to access stream")
	}

	pss := r.Permissions.FindPermissionsServicesStreams(ctx, serviceID, streamID)
	if pss == nil {
		pss = &models.PermissionsServicesStreams{
			ServiceID: serviceID,
			StreamID:  streamID,
		}
	}

	err := r.Permissions.UpdateServiceStreamPermission(ctx, pss, read, write)
	if err != nil {
		return nil, err
	}

	pss.Stream = stream

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
		return false, gqlerror.Errorf("Not allowed to edit organization resources")
	}

	err := r.Services.Delete(ctx, service)
	if err != nil {
		return false, gqlerror.Errorf("Failed deleting service: %s", err.Error())
	}

	return true, nil
}
