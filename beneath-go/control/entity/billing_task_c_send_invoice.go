package entity

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/payments/driver"
	"github.com/beneath-core/beneath-go/taskqueue"
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

	paymentDriver := db.PaymentDrivers[string(billingInfo.PaymentsDriver)]
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

	return nil
}
