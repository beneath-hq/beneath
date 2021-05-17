package resolver

import (
	"context"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/server/control/gql"
)

// StreamIndex returns the gql.StreamIndexResolver
func (r *Resolver) StreamIndex() gql.StreamIndexResolver {
	return &streamIndexResolver{r}
}

type streamIndexResolver struct{ *Resolver }

func (r *streamIndexResolver) IndexID(ctx context.Context, obj *models.StreamIndex) (string, error) {
	return obj.StreamIndexID.String(), nil
}
