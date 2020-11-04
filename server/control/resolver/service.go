package resolver

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/services/middleware"
)

func (r *queryResolver) ServiceByID(ctx context.Context, serviceID uuid.UUID) (*models.Service, error) {
	service := r.Services.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, service.Project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("You are not allowed to view organization resources")
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

func (r *mutationResolver) StageService(ctx context.Context, organizationName string, projectName string, serviceName string, description *string, sourceURL *string, readQuota *int, writeQuota *int, scanQuota *int) (*models.Service, error) {
	var project *models.Project
	var service *models.Service

	service = r.Services.FindServiceByOrganizationProjectAndName(ctx, organizationName, projectName, serviceName)
	if service == nil {
		project = r.Projects.FindProjectByOrganizationAndName(ctx, organizationName, projectName)
		if project == nil {
			return nil, gqlerror.Errorf("Project %s/%s not found", organizationName, projectName)
		}
		service = &models.Service{
			Name:       serviceName,
			Project:    project,
			ProjectID:  project.ProjectID,
			QuotaEpoch: time.Now(),
		}
	} else {
		project = service.Project
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, project.ProjectID, project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s/%s", organizationName, projectName)
	}

	err := r.Services.Stage(ctx, service, description, sourceURL, IntToInt64(readQuota), IntToInt64(writeQuota), IntToInt64(scanQuota))
	if err != nil {
		return nil, gqlerror.Errorf("Error staging service: %s", err.Error())
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
