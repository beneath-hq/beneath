package resolver

import (
	"context"

	"github.com/beneath-hq/beneath/ee/models"
	"github.com/beneath-hq/beneath/ee/server/control/gql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// BillingPlan returns the gql.BillingPlanResolver
func (r *Resolver) BillingPlan() gql.BillingPlanResolver {
	return &billingPlanResolver{r}
}

type billingPlanResolver struct{ *Resolver }

func (r *billingPlanResolver) BillingPlanID(ctx context.Context, obj *models.BillingPlan) (string, error) {
	return obj.BillingPlanID.String(), nil
}

func (r *billingPlanResolver) Currency(ctx context.Context, obj *models.BillingPlan) (string, error) {
	return string(obj.Currency), nil
}

func (r *billingPlanResolver) Period(ctx context.Context, obj *models.BillingPlan) (string, error) {
	return string(obj.Period), nil
}

func (r *queryResolver) BillingPlans(ctx context.Context) ([]*models.BillingPlan, error) {
	billingPlans := r.Billing.FindBillingPlansAvailableInUI(ctx)
	if billingPlans == nil {
		return nil, gqlerror.Errorf("Could not find billing plans")
	}

	return billingPlans, nil
}
