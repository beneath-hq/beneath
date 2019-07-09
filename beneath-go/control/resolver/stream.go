package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/model"
	uuid "github.com/satori/go.uuid"
)

// Stream returns the gql.StreamResolver
func (r *Resolver) Stream() gql.StreamResolver {
	return &streamResolver{r}
}

type streamResolver struct{ *Resolver }

func (r *streamResolver) StreamID(ctx context.Context, obj *model.Stream) (string, error) {
	panic("not implemented")
}

func (r *queryResolver) Stream(ctx context.Context, name string, projectName string) (*model.Stream, error) {
	panic("not implemented")
}

func (r *mutationResolver) CreateExternalStream(ctx context.Context, projectID uuid.UUID, name string, description *string, schema string, batch bool, manual bool) (*model.Stream, error) {
	panic("not implemented")
}

func (r *mutationResolver) UpdateStream(ctx context.Context, streamID uuid.UUID, description *string, manual *bool) (*model.Stream, error) {
	panic("not implemented")
}
