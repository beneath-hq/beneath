package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/services/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/paymentsutil"
)

func (r *queryResolver) BillingInfo(ctx context.Context, organizationID uuid.UUID) (*entity.BillingInfo, error) {
	secret := middleware.GetSecret(ctx)

	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	billingInfo := entity.FindBillingInfo(ctx, organizationID)
	if billingInfo == nil {
		return nil, gqlerror.Errorf("Billing info for organization %s not found", organizationID)
	}

	return billingInfo, nil
}

func (r *mutationResolver) UpdateBillingInfo(ctx context.Context, organizationID uuid.UUID, billingMethodID *uuid.UUID, billingPlanID uuid.UUID, country string, region *string, companyName *string, taxNumber *string) (*entity.BillingInfo, error) {
	secret := middleware.GetSecret(ctx)

	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationID.String())
	}

	prevBillingInfo := entity.FindBillingInfo(ctx, organizationID)
	if prevBillingInfo == nil {
		return nil, gqlerror.Errorf("Existing billing info not found")
	}

	var billingMethod *entity.BillingMethod
	if billingMethodID != nil {
		billingMethod = entity.FindBillingMethod(ctx, *billingMethodID)
		if billingMethod == nil {
			return nil, gqlerror.Errorf("Billing method %s not found", billingMethodID)
		}
	}

	billingPlan := entity.FindBillingPlan(ctx, billingPlanID)
	if billingPlan == nil {
		return nil, gqlerror.Errorf("Billing plan %s not found", billingPlanID)
	}

	if billingMethodID == nil && !billingPlan.Default {
		return nil, gqlerror.Errorf("A valid billing method is required for that billing plan")
	}

	if prevBillingInfo.BillingPlan.MultipleUsers && !billingPlan.MultipleUsers {
		if !secret.IsMaster() {
			return nil, gqlerror.Errorf("Enterprise plans require a Beneath Payments Admin to cancel")
		}
	}

	if paymentsutil.IsBlacklisted(country) {
		return nil, gqlerror.Errorf("Beneath does not sell its services to %s, as it's sanctioned by the EU", country)
	}

	if billingMethod != nil && billingMethod.PaymentsDriver == entity.StripeWireDriver && !billingPlan.MultipleUsers {
		return nil, gqlerror.Errorf("Only Enterprise plans allow payment by wire")
	}

	billingInfo, err := prevBillingInfo.Update(ctx, billingMethodID, billingPlanID, country, region, companyName, taxNumber)
	if err != nil {
		return nil, gqlerror.Errorf("Unable to update the organization's billing plan")
	}

	return billingInfo, nil
}
