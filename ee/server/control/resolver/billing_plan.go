package resolver

import (
	"context"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
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

func (r *queryResolver) BillingPlans(ctx context.Context) ([]*entity.BillingPlan, error) {
	billingPlans := entity.FindBillingPlansAvailableInUI(ctx)
	if billingPlans == nil {
		return nil, gqlerror.Errorf("Could not find billing plans")
	}

	return billingPlans, nil
}
