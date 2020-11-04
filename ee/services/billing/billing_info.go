package billing

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/ee/models"
	"gitlab.com/beneath-hq/beneath/ee/pkg/paymentsutil"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
)

// FindBillingInfoByOrganization finds an organization's billing info
func (s *Service) FindBillingInfoByOrganization(ctx context.Context, organizationID uuid.UUID) *models.BillingInfo {
	billingInfo := &models.BillingInfo{
		OrganizationID: organizationID,
	}
	err := s.DB.GetDB(ctx).ModelContext(ctx, billingInfo).
		Column("billing_info.*", "BillingPlan", "BillingMethod", "Organization").
		Where("billing_info.organization_id = ?", organizationID).
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return billingInfo
}

// FindBillingInfosToRun returns all billing infos where next_billing_time < now
func (s *Service) FindBillingInfosToRun(ctx context.Context) []*models.BillingInfo {
	var billingInfos []*models.BillingInfo
	err := s.DB.GetDB(ctx).ModelContext(ctx, &billingInfos).
		Column("billing_info.*", "BillingPlan", "BillingMethod", "Organization").
		Where("next_billing_time < now()").
		Select()
	if err != nil {
		panic(err)
	}
	return billingInfos
}

// UpdateBillingDetails updates billing details like country and tax number
func (s *Service) UpdateBillingDetails(ctx context.Context, bi *models.BillingInfo, country *string, region *string, companyName *string, taxNumber *string) error {
	if country != nil {
		if paymentsutil.IsBlacklisted(*country) {
			return fmt.Errorf("Cannot service customers in %s as it is under sanctions", *country)
		}
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

// UpdateLastInvoiceTime updates the billing infos LastInvoiceTime (actual invoicing must be handled upstream, this call has no side effects)
func (s *Service) UpdateLastInvoiceTime(ctx context.Context, bi *models.BillingInfo, lastInvoiceTime time.Time) error {
	bi.LastInvoiceTime = lastInvoiceTime
	bi.UpdatedOn = time.Now()
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, bi).
		WherePK().
		Column("last_invoice_time", "updated_on").
		Update()
	if err != nil {
		return err
	}
	return nil
}
