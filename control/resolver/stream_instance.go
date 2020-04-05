package resolver

import (
	"context"

	"gitlab.com/beneath-org/beneath/control/entity"
	"gitlab.com/beneath-org/beneath/control/gql"
)

// StreamInstance returns the gql.StreamInstanceResolver
func (r *Resolver) StreamInstance() gql.StreamInstanceResolver {
	return &streamInstanceResolver{r}
}

type streamInstanceResolver struct{ *Resolver }

func (r *streamInstanceResolver) InstanceID(ctx context.Context, obj *entity.StreamInstance) (string, error) {
	return obj.StreamInstanceID.String(), nil
}
