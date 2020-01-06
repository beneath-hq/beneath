package billing

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/payments"
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
	billingInfo := entity.FindBillingInfo(ctx, t.OrganizationID)
	if billingInfo == nil {
		panic("didn't find organization's billing info")
	}

	billedResources := entity.FindBilledResources(ctx, t.OrganizationID, t.BillingTime)
	if billedResources == nil {
		panic("didn't find any billed resources")
	}

	driver, err := payments.GetDriver(string(billingInfo.PaymentsDriver))
	if err != nil {
		panic("couldn't get payments driver")
	}

	err = driver.IssueInvoiceForResources(billingInfo, billedResources)
	if err != nil {
		panic("couldn't issue invoice")
	}

	return nil
}
