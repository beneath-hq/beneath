package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
)

// StreamInstance returns the gql.StreamInstanceResolver
func (r *Resolver) StreamInstance() gql.StreamInstanceResolver {
	return &streamInstanceResolver{r}
}

type streamInstanceResolver struct{ *Resolver }

func (r *streamInstanceResolver) InstanceID(ctx context.Context, obj *entity.StreamInstance) (string, error) {
	return obj.StreamInstanceID.String(), nil
}
