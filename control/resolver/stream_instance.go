package resolver

import (
	"context"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
)

// StreamInstance returns the gql.StreamInstanceResolver
func (r *Resolver) StreamInstance() gql.StreamInstanceResolver {
	return &streamInstanceResolver{r}
}

type streamInstanceResolver struct{ *Resolver }

func (r *streamInstanceResolver) StreamInstanceID(ctx context.Context, obj *entity.StreamInstance) (string, error) {
	return obj.StreamInstanceID.String(), nil
}
