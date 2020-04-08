package entity

import (
	"context"
	"time"

	"github.com/beneath-core/control/payments/driver"
	"github.com/beneath-core/control/taskqueue"
	"github.com/beneath-core/internal/hub"
	uuid "github.com/satori/go.uuid"
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
	billingInfo := FindBillingInfo(ctx, t.OrganizationID)
	if billingInfo == nil {
		panic("didn't find organization's billing info")
	}

	billedResources := FindBilledResources(ctx, t.OrganizationID, t.BillingTime)
	if billedResources == nil && len(billingInfo.Organization.Users) > 0 {
		panic("didn't find any billed resources")
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

	// delete organization if no more users
	organization := FindOrganization(ctx, t.OrganizationID)
	if organization == nil {
		panic("didn't find organization")
	}

	if len(organization.Users) == 0 {
		organization.Delete(ctx)
	}

	return nil
}
