package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/beneath-hq/beneath/ee/models"
	"github.com/beneath-hq/beneath/ee/server/control/gql"
	"github.com/beneath-hq/beneath/pkg/jsonutil"
	"github.com/beneath-hq/beneath/services/middleware"
)

// BillingMethod returns the gql.BillingMethodResolver
func (r *Resolver) BillingMethod() gql.BillingMethodResolver {
	return &billingMethodResolver{r}
}

type billingMethodResolver struct{ *Resolver }

func (r *billingMethodResolver) BillingMethodID(ctx context.Context, obj *models.BillingMethod) (string, error) {
	return obj.BillingMethodID.String(), nil
}

func (r *billingMethodResolver) PaymentsDriver(ctx context.Context, obj *models.BillingMethod) (string, error) {
	return string(obj.PaymentsDriver), nil
}

func (r *billingMethodResolver) DriverPayload(ctx context.Context, obj *models.BillingMethod) (string, error) {
	json, err := jsonutil.Marshal(obj.DriverPayload)
	return string(json), err
}

func (r *queryResolver) BillingMethods(ctx context.Context, organizationID uuid.UUID) ([]*models.BillingMethod, error) {
	secret := middleware.GetSecret(ctx)

	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationID.String())
	}

	billingMethods := r.Billing.FindBillingMethodsByOrganization(ctx, organizationID)

	// billingMethods may be empty
	return billingMethods, nil
}
