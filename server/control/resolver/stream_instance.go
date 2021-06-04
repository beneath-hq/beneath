package resolver

import (
	"context"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/server/control/gql"
)

// StreamInstance returns the gql.StreamInstanceResolver
func (r *Resolver) StreamInstance() gql.StreamInstanceResolver {
	return &streamInstanceResolver{r}
}

type streamInstanceResolver struct{ *Resolver }

func (r *streamInstanceResolver) StreamInstanceID(ctx context.Context, obj *models.StreamInstance) (string, error) {
	return obj.StreamInstanceID.String(), nil
}
