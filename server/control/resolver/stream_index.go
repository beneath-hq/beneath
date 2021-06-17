package resolver

import (
	"context"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/server/control/gql"
)

// TableIndex returns the gql.TableIndexResolver
func (r *Resolver) TableIndex() gql.TableIndexResolver {
	return &tableIndexResolver{r}
}

type tableIndexResolver struct{ *Resolver }

func (r *tableIndexResolver) IndexID(ctx context.Context, obj *models.TableIndex) (string, error) {
	return obj.TableIndexID.String(), nil
}
