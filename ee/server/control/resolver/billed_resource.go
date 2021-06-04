package resolver

import (
	"context"
	"time"

	"github.com/beneath-hq/beneath/ee/models"
	"github.com/beneath-hq/beneath/ee/server/control/gql"
	"github.com/beneath-hq/beneath/services/middleware"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// BilledResource returns the gql.BilledResourceResolver
func (r *Resolver) BilledResource() gql.BilledResourceResolver {
	return &billedResourceResolver{r}
}

type billedResourceResolver struct{ *Resolver }

func (r *billedResourceResolver) BilledResourceID(ctx context.Context, obj *models.BilledResource) (string, error) {
	return obj.BilledResourceID.String(), nil
}

func (r *billedResourceResolver) EntityKind(ctx context.Context, obj *models.BilledResource) (string, error) {
	return string(obj.EntityKind), nil
}

func (r *billedResourceResolver) Product(ctx context.Context, obj *models.BilledResource) (string, error) {
	return string(obj.Product), nil
}

func (r *billedResourceResolver) Quantity(ctx context.Context, obj *models.BilledResource) (float64, error) {
	return float64(obj.Quantity), nil
}

func (r *billedResourceResolver) Currency(ctx context.Context, obj *models.BilledResource) (string, error) {
	return string(obj.Currency), nil
}

func (r *queryResolver) BilledResources(ctx context.Context, organizationID uuid.UUID, fromBillingTime time.Time, toBillingTime time.Time) ([]*models.BilledResource, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationID.String())
	}

	billedResources := r.Billing.FindBilledResources(ctx, organizationID, fromBillingTime, toBillingTime)
	return billedResources, nil
}
