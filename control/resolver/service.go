package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
)

func (r *queryResolver) ServiceByID(ctx context.Context, serviceID uuid.UUID) (*entity.Service, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, service.ProjectID, service.Project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("You are not allowed to view organization resources")
	}

	return service, nil
}

func (r *queryResolver) ServiceByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, serviceName string) (*entity.Service, error) {
	service := entity.FindServiceByOrganizationProjectAndName(ctx, organizationName, projectName, serviceName)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s/%s/%s not found", organizationName, projectName, serviceName)
	}

	secret := middleware.GetSecret(ctx)
	selfFind := secret.IsService() && secret.GetOwnerID() == service.ServiceID
	if !selfFind {
		perms := secret.ProjectPermissions(ctx, service.ProjectID, service.Project.Public)
		if !perms.View {
			return nil, gqlerror.Errorf("You are not allowed to view organization resources")
		}
	}

	return service, nil
}

func (r *mutationResolver) StageService(ctx context.Context, organizationName string, projectName string, serviceName string, description *string, sourceURL *string, readQuota *int, writeQuota *int) (*entity.Service, error) {
	var project *entity.Project
	var service *entity.Service

	service = entity.FindServiceByOrganizationProjectAndName(ctx, organizationName, projectName, serviceName)
	if service == nil {
		project = entity.FindProjectByOrganizationAndName(ctx, organizationName, projectName)
		if project == nil {
			return nil, gqlerror.Errorf("Project %s/%s not found", organizationName, projectName)
		}
		service = &entity.Service{
			Name:      serviceName,
			ProjectID: project.ProjectID,
		}
	} else {
		project = service.Project
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, project.ProjectID, project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s/%s", organizationName, projectName)
	}

	err := service.Stage(ctx, description, sourceURL, IntToInt64(readQuota), IntToInt64(writeQuota))
	if err != nil {
		return nil, gqlerror.Errorf("Error staging service: %s", err.Error())
	}

	return service, nil
}

func (r *mutationResolver) UpdateServiceStreamPermissions(ctx context.Context, serviceID uuid.UUID, streamID uuid.UUID, read *bool, write *bool) (*entity.PermissionsServicesStreams, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	projPerms := secret.ProjectPermissions(ctx, service.ProjectID, false)
	if !projPerms.Create {
		return nil, gqlerror.Errorf("Not allowed to edit the service")
	}

	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	streamProjectPerms := secret.ProjectPermissions(ctx, stream.ProjectID, false)
	if !streamProjectPerms.Create {
		return nil, gqlerror.Errorf("Not allowed to access stream")
	}

	pss := entity.FindPermissionsServicesStreams(ctx, serviceID, streamID)
	if pss == nil {
		pss = &entity.PermissionsServicesStreams{
			ServiceID: serviceID,
			StreamID:  streamID,
		}
	}

	err := pss.Update(ctx, read, write)
	if err != nil {
		return nil, err
	}

	return pss, nil
}

func (r *mutationResolver) DeleteService(ctx context.Context, serviceID uuid.UUID) (bool, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return false, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, service.ProjectID, false)
	if !perms.Create {
		return false, gqlerror.Errorf("Not allowed to edit organization resources")
	}

	err := service.Delete(ctx)
	if err != nil {
		return false, gqlerror.Errorf("Failed deleting service: %s", err.Error())
	}

	return true, nil
}
