package resolver

import (
	"context"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/control/gql"
	"github.com/beneath-core/core/middleware"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

// Service returns the gql.ServiceResolver
func (r *Resolver) Service() gql.ServiceResolver {
	return &serviceResolver{r}
}

type serviceResolver struct{ *Resolver }

func (r *serviceResolver) ServiceID(ctx context.Context, obj *entity.Service) (string, error) {
	return obj.ServiceID.String(), nil
}

func (r *serviceResolver) Kind(ctx context.Context, obj *entity.Service) (string, error) {
	return string(obj.Kind), nil
}

func (r *queryResolver) Service(ctx context.Context, serviceID uuid.UUID) (*entity.Service, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("You are not allowed to view organization resources")
	}

	return service, nil
}

func (r *queryResolver) ServiceByNameAndOrganization(ctx context.Context, name string, organizationName string) (*entity.Service, error) {
	service := entity.FindServiceByNameAndOrganization(ctx, name, organizationName)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s/%s not found", organizationName, name)
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("You are not allowed to view organization resources")
	}

	return service, nil
}

func (r *mutationResolver) CreateService(ctx context.Context, name string, organizationID uuid.UUID, readQuota int, writeQuota int) (*entity.Service, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to administrate organization resources")
	}

	// note: you can only create External services (not Model services) via GraphQL
	service, err := entity.CreateService(ctx, name, entity.ServiceKindExternal, organizationID, readQuota, writeQuota)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (r *mutationResolver) UpdateService(ctx context.Context, serviceID uuid.UUID, name *string, readQuota *int, writeQuota *int) (*entity.Service, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to administrate organization resources")
	}

	if service.Kind != entity.ServiceKindExternal {
		return nil, gqlerror.Errorf("Can only directly edit external services")
	}

	service, err := service.UpdateDetails(ctx, name, readQuota, writeQuota)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (r *mutationResolver) UpdateServiceOrganization(ctx context.Context, serviceID uuid.UUID, organizationID uuid.UUID) (*entity.Service, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	if service.OrganizationID == organizationID {
		return nil, gqlerror.Errorf("The %s service is already owned by the %s organization", service.Name, organization.Name)
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("You are not authorized for organization %s", organizationID.String())
	}

	service, err := service.UpdateOrganization(ctx, organizationID)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return service, nil
}

func (r *mutationResolver) DeleteService(ctx context.Context, serviceID uuid.UUID) (bool, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return false, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.Admin {
		return false, gqlerror.Errorf("Not allowed to administrate organization resources")
	}

	if service.Kind != entity.ServiceKindExternal {
		return false, gqlerror.Errorf("Can only directly edit external services")
	}

	err := service.Delete(ctx)
	if err != nil {
		return false, gqlerror.Errorf("Failed deleting service: %s", err.Error())
	}

	return true, nil
}

func (r *mutationResolver) UpdateServicePermissions(ctx context.Context, serviceID uuid.UUID, streamID uuid.UUID, read *bool, write *bool) (*entity.PermissionsServicesStreams, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to administrate organization resources")
	}

	if service.Kind != entity.ServiceKindExternal {
		return nil, gqlerror.Errorf("Can only directly edit external services")
	}

	pss, err := service.UpdatePermissions(ctx, streamID, read, write)
	if err != nil {
		return nil, gqlerror.Errorf("%s", err.Error())
	}

	return pss, nil
}
