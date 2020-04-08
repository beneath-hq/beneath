package resolver

import (
	"context"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/internal/middleware"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

func (r *queryResolver) BillingInfo(ctx context.Context, organizationID uuid.UUID) (*entity.BillingInfo, error) {
	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	billingInfo := entity.FindBillingInfo(ctx, organizationID)
	if billingInfo == nil {
		return nil, gqlerror.Errorf("Billing info for organization %s not found", organizationID)
	}

	return billingInfo, nil
}

// TODO: this might have names as inputs not IDs
func (r *mutationResolver) UpdateBillingInfo(ctx context.Context, organizationID uuid.UUID, billingMethodID uuid.UUID, billingPlanID uuid.UUID) (*entity.BillingInfo, error) {
	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID)
	}

	secret := middleware.GetSecret(ctx)

	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationID.String())
	}

	billingInfo := entity.FindBillingInfo(ctx, organizationID)
	if billingInfo == nil {
		return nil, gqlerror.Errorf("Existing billing info not found")
	}

	newBillingMethod := entity.FindBillingMethod(ctx, billingMethodID)
	if newBillingMethod == nil {
		return nil, gqlerror.Errorf("Billing method %s not found", billingMethodID)
	}

	newBillingPlan := entity.FindBillingPlan(ctx, billingPlanID)
	if newBillingPlan == nil {
		return nil, gqlerror.Errorf("Billing plan %s not found", billingPlanID)
	}

	if !billingInfo.BillingPlan.Personal && newBillingPlan.Personal {

	}
	// check for eligibility to use X plan with Y billing method
	// entity.UpdateBillingInfo() ... and downstream effects
	// panic("not implemented")

	return billingInfo, nil
}
