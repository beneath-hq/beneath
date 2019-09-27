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
	if !secret.IsPersonal() {
		return nil, gqlerror.Errorf("Not allowed to create service")
	}

	// note: you can only create External services (not Model services) via GraphQL
	service, err := entity.CreateService(ctx, name, entity.ServiceKindExternal, organizationID, readBytesQuota, writeBytesQuota)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (r *mutationResolver) UpdateService(ctx context.Context, serviceID uuid.UUID, name *string, organizationID *uuid.UUID, readBytesQuota *int, writeBytesQuota *int) (*entity.Service, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsPersonal() {
		return nil, gqlerror.Errorf("Not allowed to update service")
	}

	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	// note: you can only create External services (not Model services) via GraphQL
	service, err := service.UpdateDetails(ctx, name, organizationID, readBytesQuota, writeBytesQuota)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (r *mutationResolver) DeleteService(ctx context.Context, serviceID uuid.UUID) (bool, error) {
	s := entity.FindService(ctx, serviceID)
	if s == nil {
		return false, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, s.OrganizationID)
	if !perms.Admin {
		return false, gqlerror.Errorf("Not allowed to delete services in organization %s", s.OrganizationID.String())
	}

	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return false, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	err := service.Delete(ctx)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	return true, nil
}

func (r *mutationResolver) UpdateServicePermissions(ctx context.Context, serviceID uuid.UUID, streamID uuid.UUID, read bool, write bool) (*entity.PermissionsServicesStreams, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsPersonal() {
		return nil, gqlerror.Errorf("Not allowed to update service permissions")
	}

	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	perms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to modify things in organization %s", service.OrganizationID.String())
	}

	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	pss, err := service.UpdatePermissions(ctx, streamID, read, write)
	if err != nil {
		return nil, err
	}

	return pss, nil
}
