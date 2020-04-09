package resolver

import (
	"context"

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
