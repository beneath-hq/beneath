package resolver

import (
	"context"

	"gitlab.com/beneath-org/beneath/control/entity"
	"gitlab.com/beneath-org/beneath/control/gql"
)

// StreamIndex returns the gql.StreamIndexResolver
func (r *Resolver) StreamIndex() gql.StreamIndexResolver {
	return &streamIndexResolver{r}
}

type streamIndexResolver struct{ *Resolver }

func (r *streamIndexResolver) IndexID(ctx context.Context, obj *entity.StreamIndex) (string, error) {
	return obj.StreamIndexID.String(), nil
}
