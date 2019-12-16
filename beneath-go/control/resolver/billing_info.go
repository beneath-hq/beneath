package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
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
		BillingPlanID:  billingInfo.BillingPlanID,
		PaymentsDriver: string(billingInfo.PaymentsDriver),
	}

	return billingInfoGQL, nil
}

// This might be re-used later when updating an organization's billing info (billing plan, payment driver) from the front-end
// func (r *mutationResolver) UpdateBillingPlan(ctx context.Context, organizationID uuid.UUID, billingPlanID uuid.UUID) (*entity.BillingInfo, error) {
// 	organization := entity.FindOrganization(ctx, organizationID)
// 	if organization == nil {
// 		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
// 	}

// 	secret := middleware.GetSecret(ctx)
// 	perms := secret.OrganizationPermissions(ctx, organizationID)
// 	if !perms.Admin {
// 		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
// 	}

// 	billingInfo := entity.FindBillingInfo(ctx, organizationID)
// 	if billingInfo == nil {
// 		return nil, gqlerror.Errorf("Billing info for organization %s not found", organizationID.String())
// 	}

// 	billingPlan := entity.FindBillingPlan(ctx, billingPlanID)
// 	if billingPlan == nil {
// 		return nil, gqlerror.Errorf("Billing plan %s not found", billingPlanID.String())
// 	}

// 	billingInfo, err := billingInfo.UpdateBillingPlanID(ctx, billingPlanID)
// 	if err != nil {
// 		return nil, gqlerror.Errorf("Failed to update the organization's billing plan")
// 	}

// 	return billingInfo, nil
// }
