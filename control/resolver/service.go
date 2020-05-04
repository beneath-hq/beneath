package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
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
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create organization resources")
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
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to edit organization resources")
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

func (r *mutationResolver) UpdateServiceStreamPermissions(ctx context.Context, serviceID uuid.UUID, streamID uuid.UUID, read *bool, write *bool) (*entity.PermissionsServicesStreams, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	if service.Kind != entity.ServiceKindExternal {
		return nil, gqlerror.Errorf("Can only directly edit external services")
	}

	secret := middleware.GetSecret(ctx)
	orgPerms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !orgPerms.Create {
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
	perms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.Create {
		return false, gqlerror.Errorf("Not allowed to edit organization resources")
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
