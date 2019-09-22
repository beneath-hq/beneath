package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
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
	return service, nil
}
