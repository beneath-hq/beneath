package entity

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/control/payments/driver"
	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/log"
)

// SendInvoiceTask sends the Stripe invoice to the customer
type SendInvoiceTask struct {
	OrganizationID uuid.UUID
	BillingTime    time.Time
}

// register task
func init() {
	taskqueue.RegisterTask(&SendInvoiceTask{})
}

// Run triggers the task
func (t *SendInvoiceTask) Run(ctx context.Context) error {
	billedResources := FindBilledResources(ctx, t.OrganizationID, t.BillingTime)
	if billedResources == nil {
		log.S.Infof("didn't find any billed resources for organization %s", t.OrganizationID.String())
		return nil
	}

	billingInfo := FindBillingInfo(ctx, t.OrganizationID)
	if billingInfo == nil {
		panic("didn't find organization's billing info")
	}

	if *billingInfo.BillingMethodID == uuid.Nil {
		log.S.Infof("organization %s does not pay for its usage", t.OrganizationID.String())
		return nil
	}

	paymentDriver := hub.PaymentDrivers[string(billingInfo.BillingMethod.PaymentsDriver)]
	if paymentDriver == nil {
		panic("couldn't get payments driver")
	}

	// convert []T to []interface{}
	brs := make([]driver.BilledResource, len(billedResources))
	for i := range billedResources {
		brs[i] = billedResources[i]
	}

	err := paymentDriver.IssueInvoiceForResources(billingInfo, brs)
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
