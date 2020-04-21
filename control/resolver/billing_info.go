package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
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

func (r *mutationResolver) UpdateBillingInfo(ctx context.Context, organizationID uuid.UUID, billingMethodID uuid.UUID, billingPlanID uuid.UUID, country string, state *string, companyName *string, taxNumber *string) (*entity.BillingInfo, error) {
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

	// downgrading from enterprise plan requires a Beneath Admin
	if !billingInfo.BillingPlan.Personal && newBillingPlan.Personal {
		if !secret.IsMaster() {
			return nil, gqlerror.Errorf("Enterprise plans require a Beneath Payments Admin to cancel")
		}
	}

	// TODO: check for eligibility to use X plan with Y billing method with Z tax info
	// payments.CheckBlacklist()
	// payments.CheckWirePermission()
	// payments.Check...

	// TODO: update this!
	newBillingInfo, err := billingInfo.Update(ctx, billingMethodID, billingPlanID)
	if err != nil {
		return nil, gqlerror.Errorf("Unable to update the organization's billing plan")
	}

	return newBillingInfo, nil
}

// TODO: delete this in favor of full update
func (r *mutationResolver) UpdateBillingInfoBillingMethod(ctx context.Context, organizationID uuid.UUID, billingMethodID uuid.UUID) (*entity.BillingInfo, error) {
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

	// TODO: check for eligibility to use Y billing method with X plan

	newBillingInfo, err := billingInfo.UpdateBillingMethod(ctx, billingMethodID)
	if err != nil {
		return nil, gqlerror.Errorf("Unable to update the organization's billing method")
	}

	return newBillingInfo, nil
}
