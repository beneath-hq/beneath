package entity

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/stripe"
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

type configSpecificationSendInvoiceTask struct {
	StripeSecret string `envconfig:"CONTROL_STRIPE_SECRET" required:"true"`
}

// Run triggers the task
func (t *SendInvoiceTask) Run(ctx context.Context) error {
	var config configSpecificationSendInvoiceTask
	core.LoadConfig("beneath", &config)

	stripe.InitClient(config.StripeSecret)

	organization := FindOrganization(ctx, t.OrganizationID)
	if organization == nil {
		panic("didn't find organization")
	}

	billedResources := FindBilledResources(ctx, t.OrganizationID, t.BillingTime)
	if billedResources == nil {
		panic("didn't find any billed resources")
	}

	for _, item := range billedResources {
		// only itemize the products that cost money (i.e. don't itemize the included Reads and Writes)
		if item.TotalPriceCents > 0 {
			stripe.NewInvoiceItem(organization.StripeCustomerID, int64(item.TotalPriceCents), string(organization.BillingPlan.Currency), prettyDescription(item.Product))
		}
	}

	invoice := stripe.CreateInvoice(organization.StripeCustomerID, string(organization.PaymentMethod))

	stripe.SendInvoice(invoice.ID)

	return nil
}

// prettyDescription ...
func prettyDescription(product Product) string {
	// TODO: upgrade descriptions
	switch product {
	case SeatProduct:
		return string(product)
	case ReadOverageProduct:
		return string(product)
	case WriteOverageProduct:
		return string(product)
	default:
		panic("unknown product")
	}
}
