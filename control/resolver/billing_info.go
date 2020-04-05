package resolver

import (
	"context"

	"gitlab.com/beneath-org/beneath/control/entity"
	"gitlab.com/beneath-org/beneath/control/gql"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
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
