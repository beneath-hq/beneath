package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

// BillingPlan returns the gql.BillingPlanResolver
func (r *Resolver) BillingPlan() gql.BillingPlanResolver {
	return &billingPlanResolver{r}
}

type billingPlanResolver struct{ *Resolver }

func (r *billingPlanResolver) Currency(ctx context.Context, obj *entity.BillingPlan) (string, error) {
	return string(obj.Currency), nil
}

func (r *billingPlanResolver) Period(ctx context.Context, obj *entity.BillingPlan) (string, error) {
	return string(obj.Period), nil
}

func (r *queryResolver) BillingPlan(ctx context.Context, billingPlanID uuid.UUID) (*entity.BillingPlan, error) {
	// Q: should we check admin privileges before returning billing plan?
	// secret := middleware.GetSecret(ctx)

	// perms := secret.OrganizationPermissions(ctx, organizationID)
	// if !perms.Admin {
	// 	return nil, gqlerror.Errorf("Not allowed to see billing plan without admin rights")
	// }

	billingPlan := entity.FindBillingPlan(ctx, billingPlanID)
	if billingPlan == nil {
		return nil, gqlerror.Errorf("Billing plan %s not found", billingPlanID)
	}

	return billingPlan, nil
}
