package entity

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/ee/infrastructure/payments/driver"
	"gitlab.com/beneath-hq/beneath/hub"
	"gitlab.com/beneath-hq/beneath/pkg/log"
)

// SendInvoiceTask sends the Stripe invoice to the customer
type SendInvoiceTask struct {
	BillingInfo *BillingInfo
	BillingTime time.Time
}

// Run triggers the task
func (t *SendInvoiceTask) Run(ctx context.Context) error {
	billedResources := FindBilledResources(ctx, t.BillingInfo.OrganizationID, t.BillingTime)
	if billedResources == nil {
		log.S.Infof("didn't find any billed resources for organization %s", t.BillingInfo.OrganizationID.String())
		return nil
	}

	if t.BillingInfo.BillingMethodID == nil {
		log.S.Infof("organization %s does not pay for its usage", t.BillingInfo.OrganizationID.String())
		return nil
	}

	paymentDriver := hub.PaymentDrivers[string(t.BillingInfo.BillingMethod.PaymentsDriver)]
	if paymentDriver == nil {
		panic("couldn't get payments driver")
	}

	// convert []T to []interface{}
	brs := make([]driver.BilledResource, len(billedResources))
	for i := range billedResources {
		brs[i] = billedResources[i]
	}

	err := paymentDriver.IssueInvoiceForResources(t.BillingInfo, brs)
	if err != nil {
		panic("couldn't issue invoice")
	}

	// Possibly use this in the process of downgrading someone from an Enterprise plan. Need to consider what happens to their outstanding projects.
	// delete organization if no more users
	// organization := FindOrganization(ctx, t.OrganizationID)
	// if organization == nil {
	// 	panic("didn't find organization")
	// }

	// if len(organization.Users) == 0 {
	// 	organization.Delete(ctx)
	// }

	return nil
}
