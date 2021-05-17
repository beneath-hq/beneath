package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/beneath-hq/beneath/ee/models"
	"github.com/beneath-hq/beneath/ee/server/control/gql"
	"github.com/beneath-hq/beneath/services/middleware"
)

// BillingInfo returns gql.BillingInfoResolver
func (r *Resolver) BillingInfo() gql.BillingInfoResolver { return &billingInfoResolver{r} }

type billingInfoResolver struct{ *Resolver }

func (r *billingInfoResolver) OrganizationID(ctx context.Context, obj *models.BillingInfo) (string, error) {
	return obj.OrganizationID.String(), nil
}

func (r *queryResolver) BillingInfo(ctx context.Context, organizationID uuid.UUID) (*models.BillingInfo, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	billingInfo := r.Billing.FindBillingInfoByOrganization(ctx, organizationID)
	if billingInfo == nil {
		return nil, gqlerror.Errorf("Billing info for organization %s not found", organizationID)
	}

	return billingInfo, nil
}

func (r *mutationResolver) UpdateBillingDetails(ctx context.Context, organizationID uuid.UUID, country *string, region *string, companyName *string, taxNumber *string) (*models.BillingInfo, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	bi := r.Billing.FindBillingInfoByOrganization(ctx, organizationID)
	if bi == nil {
		return nil, gqlerror.Errorf("Existing billing info not found for organization %s", organizationID.String())
	}

	err := r.Billing.UpdateBillingDetails(ctx, bi, country, region, companyName, taxNumber)
	if err != nil {
		return nil, gqlerror.Errorf("Error updating billing details: %s", err.Error())
	}

	return bi, nil
}

func (r *mutationResolver) UpdateBillingMethod(ctx context.Context, organizationID uuid.UUID, billingMethodID *uuid.UUID) (*models.BillingInfo, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	bi := r.Billing.FindBillingInfoByOrganization(ctx, organizationID)
	if bi == nil {
		return nil, gqlerror.Errorf("Existing billing info not found for organization %s", organizationID.String())
	}

	var billingMethod *models.BillingMethod
	if billingMethodID != nil {
		billingMethod = r.Billing.FindBillingMethod(ctx, *billingMethodID)
		if billingMethod == nil {
			return nil, gqlerror.Errorf("Billing method %s not found", billingMethodID.String())
		}
	}

	err := r.Billing.UpdateBillingMethod(ctx, bi, billingMethod)
	if err != nil {
		return nil, gqlerror.Errorf("Error updating billing method: %s", err.Error())
	}

	return bi, nil
}

func (r *mutationResolver) UpdateBillingPlan(ctx context.Context, organizationID uuid.UUID, billingPlanID uuid.UUID) (*models.BillingInfo, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	bi := r.Billing.FindBillingInfoByOrganization(ctx, organizationID)
	if bi == nil {
		return nil, gqlerror.Errorf("Existing billing info not found for organization %s", organizationID.String())
	}

	billingPlan := r.Billing.FindBillingPlan(ctx, billingPlanID)
	if billingPlan == nil {
		return nil, gqlerror.Errorf("Billing plan %s not found", billingPlanID)
	}

	if billingPlan.UIRank == nil {
		if !secret.IsMaster() {
			return nil, gqlerror.Errorf("The selected plan is not available for self-service, contact us to upgrade")
		}
	}

	err := r.Billing.UpdateBillingPlan(ctx, bi, billingPlan)
	if err != nil {
		return nil, gqlerror.Errorf("Error updating billing plan: %s", err.Error())
	}

	return bi, nil
}
