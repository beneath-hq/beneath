package resolver

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

// BilledResource returns the gql.BilledResourceResolver
func (r *Resolver) BilledResource() gql.BilledResourceResolver {
	return &billedResourceResolver{r}
}

type billedResourceResolver struct{ *Resolver }

func (r *billedResourceResolver) EntityKind(ctx context.Context, obj *entity.BilledResource) (string, error) {
	return string(obj.EntityKind), nil
}

func (r *billedResourceResolver) Product(ctx context.Context, obj *entity.BilledResource) (string, error) {
	return string(obj.Product), nil
}

func (r *billedResourceResolver) Currency(ctx context.Context, obj *entity.BilledResource) (string, error) {
	return string(obj.Currency), nil
}

func (r *queryResolver) BilledResources(ctx context.Context, organizationID uuid.UUID, billingTime time.Time) ([]*entity.BilledResource, error) {
	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationID.String())
	}

	billedResources := entity.FindBilledResources(ctx, organizationID, billingTime)
	if billedResources == nil {
		return nil, gqlerror.Errorf("Billed resources for organization %s and time %s not found", organizationID.String(), billingTime)
	}

	return billedResources, nil
}
