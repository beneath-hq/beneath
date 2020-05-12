package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/jsonutil"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
)

// BillingMethod returns the gql.BillingMethodResolver
func (r *Resolver) BillingMethod() gql.BillingMethodResolver {
	return &billingMethodResolver{r}
}

type billingMethodResolver struct{ *Resolver }

func (r *billingMethodResolver) PaymentsDriver(ctx context.Context, obj *entity.BillingMethod) (string, error) {
	return string(obj.PaymentsDriver), nil
}

func (r *billingMethodResolver) DriverPayload(ctx context.Context, obj *entity.BillingMethod) (string, error) {
	json, err := jsonutil.Marshal(obj.DriverPayload)
	return string(json), err
}

func (r *queryResolver) BillingMethods(ctx context.Context, organizationID uuid.UUID) ([]*entity.BillingMethod, error) {
	secret := middleware.GetSecret(ctx)

	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationID.String())
	}

	billingMethods := entity.FindBillingMethodsByOrganization(ctx, organizationID)

	// billingMethods may be empty
	return billingMethods, nil
}
