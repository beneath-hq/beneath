package billing

import (
	"context"

	"gitlab.com/beneath-hq/beneath/ee/models"
	nee_models "gitlab.com/beneath-hq/beneath/models"
)

// AddBillingInfoOnOrganizationCreated puts a newly created organization on the default plan.
// It is the *only* place where a BillingInfo row is created.
func (s *Service) AddBillingInfoOnOrganizationCreated(ctx context.Context, msg *nee_models.OrganizationCreatedEvent) error {
	defaultPlan := s.GetDefaultBillingPlan(ctx)
	bi := &models.BillingInfo{
		OrganizationID: msg.Organization.OrganizationID,
		BillingPlanID:  defaultPlan.BillingPlanID,
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

// func (s *Service) UpdatexXX(ctx context.Context, msg *models.OrganizationTransferredUserEvent) error {

// 	// TODO (and what about deducting from fromOrg??? )

// 	// find target billing info
// 	targetBillingInfo := FindBillingInfoByOrganization(ctx, targetOrg.OrganizationID)
// 	if targetBillingInfo == nil {
// 		return fmt.Errorf("Couldn't find billing info for target organization")
// 	}

// 	// add prorated seat to the target organization's next month's bill
// 	billingTime := timeutil.Next(time.Now(), targetBillingInfo.BillingPlan.Period)
// 	err = commitProratedSeatsToBill(ctx, targetBillingInfo, billingTime, []*User{user})
// 	if err != nil {
// 		panic("unable to commit prorated seat to bill")
// 	}

// 	// increment the target organization's prepaid quota by the seat quota
// 	// we do this now because we want to show the new usage capacity in the UI as soon as possible
// 	if targetBillingInfo.BillingPlan.SeatReadQuota > 0 || targetBillingInfo.BillingPlan.SeatWriteQuota > 0 || targetBillingInfo.BillingPlan.SeatScanQuota > 0 {
// 		newPrepaidReadQuota := *targetOrg.PrepaidReadQuota + targetBillingInfo.BillingPlan.SeatReadQuota
// 		newPrepaidWriteQuota := *targetOrg.PrepaidWriteQuota + targetBillingInfo.BillingPlan.SeatWriteQuota
// 		newPrepaidScanQuota := *targetOrg.PrepaidScanQuota + targetBillingInfo.BillingPlan.SeatScanQuota

// 		targetOrg.PrepaidReadQuota = &newPrepaidReadQuota
// 		targetOrg.PrepaidWriteQuota = &newPrepaidWriteQuota
// 		targetOrg.PrepaidScanQuota = &newPrepaidScanQuota

// 		targetOrg.UpdatedOn = time.Now()

// 		_, err = hub.DB.WithContext(ctx).Model(targetOrg).
// 			Column("prepaid_read_quota", "prepaid_write_quota", "prepaid_scan_quota", "updated_on").
// 			WherePK().
// 			Update()
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil

// }
