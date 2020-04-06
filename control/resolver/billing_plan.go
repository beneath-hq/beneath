package resolver

import (
	"context"

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
