package billing

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/hub"
	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// ADD TO WillCreateProjectEvent
// // check billing info if project is private
// if public != nil && !*public {
// 	bi := models.FindBillingInfo(ctx, organization.OrganizationID)
// 	if bi == nil {
// 		return nil, gqlerror.Errorf("Could not find billing info for organization %s", organizationName)
// 	}

// 	if !bi.BillingPlan.PrivateProjects {
// 		return nil, gqlerror.Errorf("Your organization's billing plan does not permit private projects")
// 	}
// }

type Service struct{}

func (s *Service) AddBilling(ctx context.Context, msg *models.OrganizationCreatedEvent) error {

	defaultBillingPlan := entity.FindDefaultBillingPlan(ctx)

	// set the organization's default quotas
	org.PrepaidReadQuota = &defaultBillingPlan.ReadQuota
	org.PrepaidWriteQuota = &defaultBillingPlan.WriteQuota
	org.PrepaidScanQuota = &defaultBillingPlan.ScanQuota
	org.ReadQuota = &defaultBillingPlan.ReadQuota
	org.WriteQuota = &defaultBillingPlan.WriteQuota
	org.ScanQuota = &defaultBillingPlan.ScanQuota

	// create billing info
	bi := &BillingInfo{
		OrganizationID: org.OrganizationID,
		BillingPlanID:  defaultBillingPlan.BillingPlanID,
	}
	_, err = tx.Model(bi).Insert()
	if err != nil {
		return err
	}
}

func (s *Service) UpdatexXX(ctx context.Context, msg *models.OrganizationTransferredUserEvent) error {

	// TODO (and what about deducting from fromOrg??? )

	// find target billing info
	targetBillingInfo := FindBillingInfo(ctx, targetOrg.OrganizationID)
	if targetBillingInfo == nil {
		return fmt.Errorf("Couldn't find billing info for target organization")
	}

	// add prorated seat to the target organization's next month's bill
	billingTime := timeutil.Next(time.Now(), targetBillingInfo.BillingPlan.Period)
	err = commitProratedSeatsToBill(ctx, targetBillingInfo, billingTime, []*User{user})
	if err != nil {
		panic("unable to commit prorated seat to bill")
	}

	// increment the target organization's prepaid quota by the seat quota
	// we do this now because we want to show the new usage capacity in the UI as soon as possible
	if targetBillingInfo.BillingPlan.SeatReadQuota > 0 || targetBillingInfo.BillingPlan.SeatWriteQuota > 0 || targetBillingInfo.BillingPlan.SeatScanQuota > 0 {
		newPrepaidReadQuota := *targetOrg.PrepaidReadQuota + targetBillingInfo.BillingPlan.SeatReadQuota
		newPrepaidWriteQuota := *targetOrg.PrepaidWriteQuota + targetBillingInfo.BillingPlan.SeatWriteQuota
		newPrepaidScanQuota := *targetOrg.PrepaidScanQuota + targetBillingInfo.BillingPlan.SeatScanQuota

		targetOrg.PrepaidReadQuota = &newPrepaidReadQuota
		targetOrg.PrepaidWriteQuota = &newPrepaidWriteQuota
		targetOrg.PrepaidScanQuota = &newPrepaidScanQuota

		targetOrg.UpdatedOn = time.Now()

		_, err = hub.DB.WithContext(ctx).Model(targetOrg).
			Column("prepaid_read_quota", "prepaid_write_quota", "prepaid_scan_quota", "updated_on").
			WherePK().
			Update()
		if err != nil {
			return err
		}
	}

	return nil

}

func (s *Service) UpdatePrepaidQuotas(ctx context.Context, o *models.Organization, bp *models.BillingPlan) error {
	numSeats := int64(len(o.Users))
	prepaidReadQuota := billingPlan.BaseReadQuota + billingPlan.SeatReadQuota*numSeats
	prepaidWriteQuota := billingPlan.BaseWriteQuota + billingPlan.SeatWriteQuota*numSeats
	prepaidScanQuota := billingPlan.BaseScanQuota + billingPlan.SeatScanQuota*numSeats

	// set fields
	o.PrepaidReadQuota = &prepaidReadQuota
	o.PrepaidWriteQuota = &prepaidWriteQuota
	o.PrepaidScanQuota = &prepaidScanQuota
	o.UpdatedOn = time.Now()

	// update
	_, err := hub.DB.WithContext(ctx).Model(o).
		Column("prepaid_read_quota", "prepaid_write_quota", "prepaid_scan_quota", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	return err
}
