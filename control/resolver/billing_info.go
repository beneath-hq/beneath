package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
)

func (r *queryResolver) BillingInfo(ctx context.Context, organizationID uuid.UUID) (*gql.BillingInfo, error) {
	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	billingInfo := entity.FindBillingInfo(ctx, organizationID)
	if billingInfo == nil {
		return nil, gqlerror.Errorf("Billing info for organization %s not found", organizationID)
	}

	billingInfoGQL := &gql.BillingInfo{
		OrganizationID: billingInfo.OrganizationID,
		BillingPlan:    billingInfo.BillingPlan,
		PaymentsDriver: string(billingInfo.PaymentsDriver),
	}

	return billingInfoGQL, nil
}
