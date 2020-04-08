package resolver

import (
	"context"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/control/gql"
)

// BillingMethod returns the gql.BillingMethodResolver
func (r *Resolver) BillingMethod() gql.BillingMethodResolver {
	return &billingMethodResolver{r}
}

type billingMethodResolver struct{ *Resolver }

func (r *billingMethodResolver) PaymentsDriver(ctx context.Context, obj *entity.BillingMethod) (string, error) {
	return string(obj.PaymentsDriver), nil
}
