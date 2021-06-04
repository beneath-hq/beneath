package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/beneath-hq/beneath/ee/models"
	nee_models "github.com/beneath-hq/beneath/models"
)

// HandleOrganizationCreatedEvent puts a newly created organization on the default plan.
// It is the *only* place where a BillingInfo row is created.
func (s *Service) HandleOrganizationCreatedEvent(ctx context.Context, msg *nee_models.OrganizationCreatedEvent) error {
	quotaPeriod := s.Usage.GetQuotaPeriod(msg.Organization.QuotaEpoch)

	defaultPlan := s.GetDefaultBillingPlan(ctx)
	bi := &models.BillingInfo{
		Organization:    msg.Organization,
		OrganizationID:  msg.Organization.OrganizationID,
		BillingPlan:     defaultPlan,
		BillingPlanID:   defaultPlan.BillingPlanID,
		NextBillingTime: quotaPeriod.Next(time.Now()),
		LastInvoiceTime: time.Time{},
	}

	return s.DB.InTransaction(ctx, func(ctx context.Context) error {
		// create billing info
		_, err := s.DB.GetDB(ctx).ModelContext(ctx, bi).Insert()
		if err != nil {
			return err
		}

		// set quotas on organization
		return s.ComputeAndUpdateOrganizationQuotas(ctx, bi, 1) // assumption that new organizations have 1 member
	})
}

// HandleOrganizationTransferredUserEvent performs seat-related billing for the user on its source/destination organization
func (s *Service) HandleOrganizationTransferredUserEvent(ctx context.Context, msg *nee_models.OrganizationTransferredUserEvent) error {
	// Desired behaviour:
	// 1. Seat removed from source, but they keep the user's prepaid quota to avoid refunding money.
	// --> Done, organization quotas are recalculated when billing runs (ComputeAndUpdateOrganizationQuotas).
	// 2. Add prorated seat bill to destination organization, add prorated prepaid quota.
	//    (We don't really have to worry about the users previous usage/overage since that's tracked directly on the billing org for billing purposes)

	bi := s.FindBillingInfoByOrganization(ctx, msg.Target.OrganizationID)
	if bi == nil {
		return fmt.Errorf("target organization %s is nil", msg.Target.OrganizationID)
	}

	return s.AddProratedUser(ctx, bi, msg.User, time.Now())
}
