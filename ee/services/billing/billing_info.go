package billing

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/ee/models"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
)

// FindBillingInfoByOrganization finds an organization's billing info
func (s *Service) FindBillingInfoByOrganization(ctx context.Context, organizationID uuid.UUID) *models.BillingInfo {
	billingInfo := &models.BillingInfo{
		OrganizationID: organizationID,
	}
	err := s.DB.GetDB(ctx).ModelContext(ctx, billingInfo).
		Column("billing_info.*", "BillingPlan", "BillingMethod", "Organization.Users").
		Where("billing_info.organization_id = ?", organizationID).
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return billingInfo
}

// FindAllPayingBillingInfos returns all billing infos where the organization is on a paid plan
func (s *Service) FindAllPayingBillingInfos(ctx context.Context) []*models.BillingInfo {
	var billingInfos []*models.BillingInfo
	err := s.DB.GetDB(ctx).ModelContext(ctx, &billingInfos).
		Column("billing_info.*", "BillingPlan", "BillingMethod", "Organization.Users").
		Where("billing_method_id IS NOT NULL").
		Select()
	if err != nil {
		panic(err)
	}
	return billingInfos
}

// UpdateBillingDetails updates billing details like country and tax number
func (s *Service) UpdateBillingDetails(ctx context.Context, bi *models.BillingInfo, country *string, region *string, companyName *string, taxNumber *string) error {
	if country != nil {
		bi.Country = *country
	}
	if region != nil {
		bi.Region = *region
	}
	if companyName != nil {
		bi.CompanyName = *companyName
	}
	if taxNumber != nil {
		bi.TaxNumber = *taxNumber
	}

	bi.UpdatedOn = time.Now()
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, bi).
		Column(
			"country",
			"region",
			"company_name",
			"tax_number",
			"updated_on",
		).WherePK().Update()
	if err != nil {
		return err
	}

	return nil
}

// UpdateBillingMethod updates the selected billing method for the organization
func (s *Service) UpdateBillingMethod(ctx context.Context, bi *models.BillingInfo, billingMethod *models.BillingMethod) error {
	if billingMethod == nil && !bi.BillingPlan.IsFree() {
		return fmt.Errorf("cannot remove billing method when not on a free plan")
	}

	if billingMethod == nil {
		bi.BillingMethod = nil
		bi.BillingMethodID = nil
	} else {
		bi.BillingMethod = billingMethod
		bi.BillingMethodID = &billingMethod.BillingMethodID
	}

	bi.UpdatedOn = time.Now()
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, bi).Column("billing_method_id", "updated_on").WherePK().Update()
	if err != nil {
		return err
	}

	return nil
}

// UpdateBillingPlan updates an organization's billing plan, probably triggering a bill and updating allowed quotas.
func (s *Service) UpdateBillingPlan(ctx context.Context, bi *models.BillingInfo, billingPlan *models.BillingPlan) error {
	// prepare
	oldBillingPlan := bi.BillingPlan
	newBillingPlan := billingPlan

	// done if we're already on billingPlan
	if oldBillingPlan.BillingPlanID == newBillingPlan.BillingPlanID {
		return nil
	}

	// cannot change to paid plan if no billing method set
	if !newBillingPlan.IsFree() && bi.BillingMethodID == nil {
		return fmt.Errorf("cannot upgrade to paid billing plan without a registered billing method")
	}

	// get all members, for seat-related calculations
	users, err := s.Organizations.FindOrganizationMembers(ctx, bi.OrganizationID)
	if err != nil {
		return err
	}

	// run transaction that updates billing info, and also commits billed resources and updates quotas
	err = s.DB.InTransaction(ctx, func(ctx context.Context) error {
		// add current overage to billed resources (under old plan)
		// commitOverageToBill

		// update billing info
		bi.BillingPlan = newBillingPlan
		bi.BillingPlanID = newBillingPlan.BillingPlanID
		bi.UpdatedOn = time.Now()
		_, err := s.DB.GetDB(ctx).ModelContext(ctx, bi).Column("billing_plan_id", "updated_on").WherePK().Update()
		if err != nil {
			return err
		}
		// TODO: also set new billing base time

		// add base price of the new plan to billed resources

		// TODO (future): credit unused resources from old plan (time-constrained)

		// update organization quotas based on new plan
		err = s.ComputeAndUpdateOrganizationQuotas(ctx, bi, len(users))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// trigger invoicing on org (SendInvoiceTask)

	return nil
}

// ComputeAndUpdateOrganizationQuotas derives quotas for the organization (based on its billing plan), and updates them
func (s *Service) ComputeAndUpdateOrganizationQuotas(ctx context.Context, bi *models.BillingInfo, numSeats int) error {
	org := bi.Organization
	plan := bi.BillingPlan

	// update prepaid quota field (only used when showing billing info)
	org.PrepaidReadQuota = int64ToPtr(plan.BaseReadQuota + plan.SeatReadQuota*int64(numSeats))
	org.PrepaidWriteQuota = int64ToPtr(plan.BaseWriteQuota + plan.SeatWriteQuota*int64(numSeats))
	org.PrepaidScanQuota = int64ToPtr(plan.BaseScanQuota + plan.SeatScanQuota*int64(numSeats))
	org.UpdatedOn = time.Now()
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, org).
		Column("prepaid_read_quota", "prepaid_write_quota", "prepaid_scan_quota", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	// update normal quotas (important, sets limits on activity)
	err = s.Organizations.UpdateQuotas(ctx, org, &plan.ReadQuota, &plan.WriteQuota, &plan.ScanQuota)
	if err != nil {
		return nil
	}

	return nil
}

func int64ToPtr(val int64) *int64 {
	return &val
}
