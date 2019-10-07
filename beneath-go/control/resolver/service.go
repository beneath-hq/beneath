package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/core/middleware"
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

func (r *serviceResolver) ReadBytesQuota(ctx context.Context, obj *entity.Service) (int, error) {
	return int(obj.ReadQuota), nil
}
func (r *serviceResolver) WriteBytesQuota(ctx context.Context, obj *entity.Service) (int, error) {
	return int(obj.WriteQuota), nil
}

func (r *queryResolver) Service(ctx context.Context, serviceID uuid.UUID) (*entity.Service, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}
	return service, nil
}

func (r *mutationResolver) CreateService(ctx context.Context, name string, organizationID uuid.UUID, readBytesQuota int, writeBytesQuota int) (*entity.Service, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to administrate organization resources")
	}

	// note: you can only create External services (not Model services) via GraphQL
	service, err := entity.CreateService(ctx, name, entity.ServiceKindExternal, organizationID, readBytesQuota, writeBytesQuota)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (r *mutationResolver) UpdateService(ctx context.Context, serviceID uuid.UUID, name *string, readBytesQuota *int, writeBytesQuota *int) (*entity.Service, error) {
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

	service, err := service.UpdateDetails(ctx, name, readBytesQuota, writeBytesQuota)
	if err != nil {
		return nil, err
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

func (r *mutationResolver) UpdateServicePermissions(ctx context.Context, serviceID uuid.UUID, streamID uuid.UUID, read bool, write bool) (*entity.PermissionsServicesStreams, error) {
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
