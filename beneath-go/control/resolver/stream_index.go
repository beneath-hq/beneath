package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
)

// StreamIndex returns the gql.StreamIndexResolver
func (r *Resolver) StreamIndex() gql.StreamIndexResolver {
	return &streamIndexResolver{r}
}

type streamIndexResolver struct{ *Resolver }

func (r *streamIndexResolver) IndexID(ctx context.Context, obj *entity.StreamIndex) (string, error) {
	return obj.StreamIndexID.String(), nil
}
