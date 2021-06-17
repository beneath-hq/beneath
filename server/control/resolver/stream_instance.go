package resolver

import (
	"context"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/server/control/gql"
)

// TableInstance returns the gql.TableInstanceResolver
func (r *Resolver) TableInstance() gql.TableInstanceResolver {
	return &tableInstanceResolver{r}
}

type tableInstanceResolver struct{ *Resolver }

func (r *tableInstanceResolver) TableInstanceID(ctx context.Context, obj *models.TableInstance) (string, error) {
	return obj.TableInstanceID.String(), nil
}
