package payments

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/beneath-hq/beneath/ee/models"
)

// HandleShouldInvoiceBilledResourcesEvent triggers invoicing after new billed resources are committed
func (s *Service) HandleShouldInvoiceBilledResourcesEvent(ctx context.Context, msg *models.ShouldInvoiceBilledResourcesEvent) error {
	bi := s.Billing.FindBillingInfoByOrganization(ctx, msg.OrganizationID)
	if bi == nil {
		return fmt.Errorf("billing info not found for organizationID=%s", msg.OrganizationID.String())
	}

	err := s.invoiceOutstanding(ctx, bi)
	if err != nil {
		return err
	}

	return nil
}

// invoiceOutstanding sends an invoice for every billed resource between bi.LastInvoiceTime (exclusive) and now (inclusive)
func (s *Service) invoiceOutstanding(ctx context.Context, bi *models.BillingInfo) error {
	invoiceFromTime := bi.LastInvoiceTime
	invoiceToTime := time.Now()
	if invoiceFromTime.After(invoiceToTime) {
		return nil
	}

	resources := s.Billing.FindBilledResources(ctx, bi.OrganizationID, invoiceFromTime, invoiceToTime)
	if len(resources) == 0 {
		s.Logger.Infof("didn't find any billed resources for organization %s", bi.OrganizationID.String())
		return nil
	}

	if bi.BillingMethodID == nil {
		s.Logger.Infof("organization %s does not have billing method, skipping", bi.OrganizationID.String())
		return nil
	}

	err := s.IssueInvoiceForResources(ctx, bi, resources)
	if err != nil {
		return err
	}

	err = s.Billing.UpdateLastInvoiceTime(ctx, bi, invoiceToTime)
	if err != nil {
		return err
	}

	return nil
}
