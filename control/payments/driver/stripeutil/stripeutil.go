package stripeutil

import (
	"time"

	uuid "github.com/satori/go.uuid"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/invoice"
	"github.com/stripe/stripe-go/invoiceitem"
	"github.com/stripe/stripe-go/paymentmethod"
	"github.com/stripe/stripe-go/setupintent"
	"github.com/stripe/stripe-go/taxrate"
	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/payments/driver"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/pkg/paymentsutil"
)

const (
	daysUntilInvoiceDue = 15
)

// InitStripe sets our Stripe API key
func InitStripe(stripeKey string) {
	stripe.Key = stripeKey
}

// GenerateSetupIntent gets ready for a customer to add credit card information
func GenerateSetupIntent(organizationID uuid.UUID) *stripe.SetupIntent {
	params := &stripe.SetupIntentParams{
		PaymentMethodTypes: stripe.StringSlice([]string{string(stripe.PaymentMethodTypeCard)}),
		Usage:              stripe.String(string(stripe.SetupIntentUsageOffSession)),
	}
	params.AddMetadata("OrganizationID", organizationID.String())

	setupIntent, err := setupintent.New(params)
	if err != nil {
		panic("stripe error when initiating setupintent")
	}

	return setupIntent
}

// AttachPaymentMethod attaches a payment method to customer; use this when updating a customer's payment method
func AttachPaymentMethod(customerID string, paymentMethodID string) *stripe.PaymentMethod {
	paymentMethod, err := paymentmethod.Attach(paymentMethodID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	})
	if err != nil {
		panic("stripe error when attaching a paymentmethod to customer")
	}

	return paymentMethod
}

// RetrieveCustomer retrieves a customer from Stripe
func RetrieveCustomer(customerID string) *stripe.Customer {
	customer, err := customer.Get(customerID, nil)
	if err != nil {
		panic("stripe error when retrieving a customer")
	}

	return customer
}

// RetrievePaymentMethod returns a payment method, with which we access billing_details
func RetrievePaymentMethod(paymentMethodID string) *stripe.PaymentMethod {
	paymentMethod, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		panic("stripe error when getting payment method")
	}

	return paymentMethod
}

// CreateCardCustomer registers a card-paying customer with Stripe
func CreateCardCustomer(organizationID uuid.UUID, orgName string, email string, pm *stripe.PaymentMethod) *stripe.Customer {
	params := &stripe.CustomerParams{
		Name:          stripe.String(orgName),
		Email:         stripe.String(email),
		PaymentMethod: stripe.String(pm.ID),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm.ID),
		},
	}
	params.AddMetadata("OrganizationID", organizationID.String())

	customer, err := customer.New(params)
	if err != nil {
		panic("stripe error when creating a customer")
	}

	return customer
}

// CreateWireCustomer registers a wire-paying customer with Stripe
func CreateWireCustomer(organizationID uuid.UUID, orgName string, email string) *stripe.Customer {
	params := &stripe.CustomerParams{
		Name:  stripe.String(orgName),
		Email: stripe.String(email),
	}
	params.AddMetadata("OrganizationID", organizationID.String())

	customer, err := customer.New(params)
	if err != nil {
		panic("stripe error when creating a customer")
	}

	return customer
}

// UpdateCardCustomer updates the customer with Stripe
func UpdateCardCustomer(customerID string, email string, paymentMethodID string) *stripe.Customer {
	customer, err := customer.Update(customerID, &stripe.CustomerParams{
		Email: stripe.String(email),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	})
	if err != nil {
		panic("stripe error when updating a customer")
	}

	return customer
}

// UpdateWireCustomer updates the customer with Stripe
func UpdateWireCustomer(customerID string, email string) *stripe.Customer {
	customer, err := customer.Update(customerID, &stripe.CustomerParams{
		Email: stripe.String(email),
	})
	if err != nil {
		panic("stripe error when updating a customer")
	}

	return customer
}

// NewInvoiceItemSeats adds an invoice line item for seats
func NewInvoiceItemSeats(customerID string, quantity int64, unitAmount int64, currency string, start int64, end int64, description string) *stripe.InvoiceItem {
	params := &stripe.InvoiceItemParams{
		Customer:    stripe.String(customerID),
		Quantity:    stripe.Int64(quantity),
		UnitAmount:  stripe.Int64(unitAmount),
		Currency:    stripe.String(string(currency)),
		Description: stripe.String(description),
		Period: &stripe.InvoiceItemPeriodParams{
			Start: stripe.Int64(start),
			End:   stripe.Int64(end),
		},
	}

	invoiceItem, err := invoiceitem.New(params)
	if err != nil {
		panic("stripe error when adding invoice item")
	}

	return invoiceItem
}

// NewInvoiceItemOther adds an invoice line item for non-seat products
func NewInvoiceItemOther(customerID string, amount int64, currency string, start int64, end int64, description string) *stripe.InvoiceItem {
	params := &stripe.InvoiceItemParams{
		Customer:    stripe.String(customerID),
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(string(currency)),
		Description: stripe.String(description),
		Period: &stripe.InvoiceItemPeriodParams{
			Start: stripe.Int64(start),
			End:   stripe.Int64(end),
		},
	}

	invoiceItem, err := invoiceitem.New(params)
	if err != nil {
		panic("stripe error when adding invoice item")
	}

	return invoiceItem
}

// CreateTaxRate creates a Stripe Tax Rate object
func CreateTaxRate(taxPercentage int, description string) *stripe.TaxRate {
	params := &stripe.TaxRateParams{
		DisplayName: stripe.String(description),
		Percentage:  stripe.Float64(float64(taxPercentage)),
		Inclusive:   stripe.Bool(false),
	}
	tr, err := taxrate.New(params)
	if err != nil {
		panic("stripe error when creating tax rate object")
	}

	return tr
}

// CreateInvoice creates an invoice
// there must exist invoice line items for the given customer before this function is called
func CreateInvoice(customerID string, paymentsDriver string, taxRate *stripe.TaxRate) *stripe.Invoice {
	params := &stripe.InvoiceParams{}
	if paymentsDriver == string(entity.StripeCardDriver) {
		params = &stripe.InvoiceParams{
			Customer:         stripe.String(customerID),
			DefaultTaxRates:  stripe.StringSlice([]string{taxRate.ID}),
			CollectionMethod: stripe.String(string(stripe.InvoiceCollectionMethodChargeAutomatically)),
			AutoAdvance:      stripe.Bool(true), // this tells Stripe to re-try failed payments using their "Smart Retries" feature; it will retry up to 4 times
		}
	} else if paymentsDriver == string(entity.StripeWireDriver) {
		params = &stripe.InvoiceParams{
			Customer:         stripe.String(customerID),
			DefaultTaxRates:  stripe.StringSlice([]string{taxRate.ID}),
			DaysUntilDue:     stripe.Int64(daysUntilInvoiceDue),
			CollectionMethod: stripe.String(string(stripe.InvoiceCollectionMethodSendInvoice)),
		}
	} else {
		panic("unrecognized payment driver, or payment driver does not use invoices")
	}

	invoice, err := invoice.New(params)
	if err != nil {
		panic("stripe error when creating invoice")
	}

	return invoice
}

// CreateStripeInvoice is a wrapper to create a Stripe invoice using the stripeutil functions
func CreateStripeInvoice(billingInfo driver.BillingInfo, billedResources []driver.BilledResource) *stripe.Invoice {
	if billingInfo.GetDriverPayload()["customer_id"] == nil {
		panic("stripe customer id is not set")
	}

	var seatCount int64
	var seatPrice int64
	var seatStartTime int64
	var seatEndTime int64

	for _, item := range billedResources {
		// only itemize the products that cost money (i.e. don't itemize the included Reads and Writes)
		if item.GetTotalPriceCents() != 0 {
			if item.GetProduct() == string(entity.SeatProduct) { // count seats for the batch line item
				seatCount++
				seatPrice = int64(item.GetTotalPriceCents())
				seatStartTime = item.GetStartTime().Unix()
				seatEndTime = item.GetEndTime().Unix()
			} else { // itemize everything else
				NewInvoiceItemOther(billingInfo.GetDriverPayload()["customer_id"].(string), int64(item.GetTotalPriceCents()), string(billingInfo.GetBillingPlanCurrency()), item.GetStartTime().Unix(), item.GetEndTime().Unix(), PrettyDescription(item.GetProduct()))
			}
		}
	}

	// batch seats
	if seatCount > 0 {
		NewInvoiceItemSeats(billingInfo.GetDriverPayload()["customer_id"].(string), seatCount, seatPrice, string(billingInfo.GetBillingPlanCurrency()), seatStartTime, seatEndTime, PrettyDescription(string(entity.SeatProduct), ""))
	}

	// add tax
	taxPercentage, description := paymentsutil.ComputeTaxPercentage(billingInfo.GetCountry(), billingInfo.GetRegion(), billingInfo.IsCompany(), time.Now())
	taxRate := CreateTaxRate(taxPercentage, description)

	inv := CreateInvoice(billingInfo.GetDriverPayload()["customer_id"].(string), billingInfo.GetPaymentsDriver(), taxRate)

	return inv
}

// ListCustomerInvoices lists a customer's historical invoices
func ListCustomerInvoices(customerID string) []*stripe.Invoice {
	var invoices []*stripe.Invoice
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(customerID),
	}
	i := invoice.List(params)
	idx := 0
	for i.Next() {
		invoices[idx] = i.Invoice()
		idx++
	}

	return invoices
}

// PayInvoice triggers Stripe to pay the invoice on behalf of the customer
// use this for customers who pay by card and are automatically charged
func PayInvoice(invoiceID string) {
	_, err := invoice.Pay(invoiceID, nil)
	if err != nil {
		// don't panic under these circumstances, just log
		if stripeErr, ok := err.(*stripe.Error); ok {
			switch stripeErr.Code {
			case stripe.ErrorCodeCardDeclined:
				log.S.Infof(stripeErr.Msg)
			case stripe.ErrorCodeExpiredCard:
				log.S.Infof(stripeErr.Msg)
			case "invoice_payment_intent_requires_action":
				// stripe will send the customer an email, bring them to a hosted checkout page, and get manual authorization
				// https://stripe.com/docs/error-codes/invoice-payment-intent-requires-action
				// https://stripe.com/docs/billing/subscriptions/payment#handling-action-required
				log.S.Infof(stripeErr.Msg)
			default:
				panic("stripe error when paying invoice")
			}
		} else {
			panic("stripe error when paying invoice")
		}
	}
}

// SendInvoice triggers Stripe to send the invoice to the customer
// use this for customers who pay by wire and are not automatically charged
func SendInvoice(invoiceID string) {
	_, err := invoice.SendInvoice(invoiceID, nil)
	if err != nil {
		panic("stripe error when sending invoice")
	}
}

// PrettyDescription makes the Stripe invoice more informative
func PrettyDescription(product string) string {
	switch product {
	case string(entity.SeatProduct):
		return "Seats"
	case string(entity.SeatProratedProduct):
		return "Seat (prorated)"
	case string(entity.SeatProratedCreditProduct):
		return "Seat credit (prorated)"
	case string(entity.ReadOverageProduct):
		return "Organization read overage (GB)"
	case string(entity.WriteOverageProduct):
		return "Organization write overage (GB)"
	default:
		panic("unknown product")
	}
}
