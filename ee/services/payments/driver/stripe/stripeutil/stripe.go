package stripeutil

import (
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/invoice"
	"github.com/stripe/stripe-go/invoiceitem"
	"github.com/stripe/stripe-go/paymentmethod"
	"github.com/stripe/stripe-go/setupintent"
	"github.com/stripe/stripe-go/taxrate"
	"go.uber.org/zap"

	"gitlab.com/beneath-hq/beneath/ee/models"
	"gitlab.com/beneath-hq/beneath/ee/pkg/paymentsutil"
	"gitlab.com/beneath-hq/beneath/ee/services/payments/driver"
)

const (
	daysUntilInvoiceDue = 15
)

// InitStripe sets our Stripe API key
func InitStripe(stripeKey string) {
	if stripe.Key != "" && stripe.Key != stripeKey {
		panic(fmt.Errorf("currently doesn't support multiple different Stripe clients"))
	}
	stripe.Key = stripeKey
}

// GenerateSetupIntent gets ready for a customer to add credit card information
func GenerateSetupIntent(organizationID uuid.UUID) (*stripe.SetupIntent, error) {
	params := &stripe.SetupIntentParams{
		PaymentMethodTypes: stripe.StringSlice([]string{string(stripe.PaymentMethodTypeCard)}),
		Usage:              stripe.String(string(stripe.SetupIntentUsageOffSession)),
	}
	params.AddMetadata("OrganizationID", organizationID.String())

	setupIntent, err := setupintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe error when initiating setupintent: %s", err.Error())
	}

	return setupIntent, nil
}

// AttachPaymentMethod attaches a payment method to customer; use this when updating a customer's payment method
func AttachPaymentMethod(customerID string, paymentMethodID string) (*stripe.PaymentMethod, error) {
	paymentMethod, err := paymentmethod.Attach(paymentMethodID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	})
	if err != nil {
		return nil, fmt.Errorf("stripe error when attaching a paymentmethod to customer: %s", err.Error())
	}

	return paymentMethod, nil
}

// RetrieveCustomer retrieves a customer from Stripe
func RetrieveCustomer(customerID string) (*stripe.Customer, error) {
	customer, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe error when retrieving a customer: %s", err.Error())
	}

	return customer, nil
}

// RetrievePaymentMethod returns a payment method, with which we access billing_details
func RetrievePaymentMethod(paymentMethodID string) (*stripe.PaymentMethod, error) {
	paymentMethod, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe error when getting payment method: %s", err.Error())
	}

	return paymentMethod, nil
}

// CreateCardCustomer registers a card-paying customer with Stripe
func CreateCardCustomer(organizationID uuid.UUID, orgName string, email string, pm *stripe.PaymentMethod) (*stripe.Customer, error) {
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
		return nil, fmt.Errorf("stripe error when creating a customer: %s", err.Error())
	}

	return customer, nil
}

// CreateWireCustomer registers a wire-paying customer with Stripe
func CreateWireCustomer(organizationID uuid.UUID, orgName string, email string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Name:  stripe.String(orgName),
		Email: stripe.String(email),
	}
	params.AddMetadata("OrganizationID", organizationID.String())

	customer, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe error when creating a customer: %s", err.Error())
	}

	return customer, nil
}

// UpdateCardCustomer updates the customer with Stripe
func UpdateCardCustomer(customerID string, email string, paymentMethodID string) (*stripe.Customer, error) {
	customer, err := customer.Update(customerID, &stripe.CustomerParams{
		Email: stripe.String(email),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("stripe error when updating a customer: %s", err.Error())
	}

	return customer, nil
}

// UpdateWireCustomer updates the customer with Stripe
func UpdateWireCustomer(customerID string, email string) (*stripe.Customer, error) {
	customer, err := customer.Update(customerID, &stripe.CustomerParams{
		Email: stripe.String(email),
	})
	if err != nil {
		return nil, fmt.Errorf("stripe error when updating a customer: %s", err.Error())
	}

	return customer, nil
}

// NewInvoiceItemSeats adds an invoice line item for seats
func NewInvoiceItemSeats(customerID string, quantity int64, unitAmount int64, currency string, start time.Time, end time.Time, description string) (*stripe.InvoiceItem, error) {
	params := &stripe.InvoiceItemParams{
		Customer:    stripe.String(customerID),
		Quantity:    stripe.Int64(quantity),
		UnitAmount:  stripe.Int64(unitAmount),
		Currency:    stripe.String(string(currency)),
		Description: stripe.String(description),
		Period: &stripe.InvoiceItemPeriodParams{
			Start: stripe.Int64(start.Unix()),
			End:   stripe.Int64(end.Unix()),
		},
	}

	invoiceItem, err := invoiceitem.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe error when adding invoice item: %s", err.Error())
	}

	return invoiceItem, nil
}

// NewInvoiceItemOther adds an invoice line item for non-seat products
func NewInvoiceItemOther(customerID string, amount int64, currency models.Currency, start time.Time, end time.Time, description string) (*stripe.InvoiceItem, error) {
	params := &stripe.InvoiceItemParams{
		Customer:    stripe.String(customerID),
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(string(currency)),
		Description: stripe.String(description),
		Period: &stripe.InvoiceItemPeriodParams{
			Start: stripe.Int64(start.Unix()),
			End:   stripe.Int64(end.Unix()),
		},
	}

	invoiceItem, err := invoiceitem.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe error when adding invoice item: %s", err.Error())
	}

	return invoiceItem, nil
}

// CreateTaxRate creates a Stripe Tax Rate object
func CreateTaxRate(taxPercentage int, description string) (*stripe.TaxRate, error) {
	params := &stripe.TaxRateParams{
		DisplayName: stripe.String(description),
		Percentage:  stripe.Float64(float64(taxPercentage)),
		Inclusive:   stripe.Bool(false),
	}
	tr, err := taxrate.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe error when creating tax rate object: %s", err.Error())
	}

	return tr, nil
}

// CreateInvoice creates an invoice
// there must exist invoice line items for the given customer before this function is called
func CreateInvoice(customerID string, paymentsDriver string, taxRate *stripe.TaxRate) (*stripe.Invoice, error) {
	params := &stripe.InvoiceParams{
		Customer:        stripe.String(customerID),
		DefaultTaxRates: stripe.StringSlice([]string{taxRate.ID}),
	}

	switch paymentsDriver {
	case driver.StripeCard:
		params.CollectionMethod = stripe.String(string(stripe.InvoiceCollectionMethodChargeAutomatically))
		params.AutoAdvance = stripe.Bool(true) // // this tells Stripe to re-try failed payments using their "Smart Retries" feature; it will retry up to 4 times
	case driver.StripeWire:
		params.DaysUntilDue = stripe.Int64(daysUntilInvoiceDue)
		params.CollectionMethod = stripe.String(string(stripe.InvoiceCollectionMethodSendInvoice))
	default:
		panic(fmt.Errorf("unrecognized payment driver, or payment driver does not use invoices"))
	}

	invoice, err := invoice.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe error when creating invoice: %s", err.Error())
	}

	return invoice, nil
}

// CreateStripeInvoice is a wrapper to create a Stripe invoice using the stripeutil functions
func CreateStripeInvoice(billingInfo *models.BillingInfo, billedResources []*models.BilledResource) (*stripe.Invoice, error) {
	customerID, ok := billingInfo.BillingMethod.DriverPayload["customer_id"].(string)
	if !ok || customerID == "" {
		return nil, fmt.Errorf("stripe customer id is not set")
	}

	var seatCount int64
	var seatPrice int64
	var seatStartTime time.Time
	var seatEndTime time.Time

	for _, item := range billedResources {
		// only itemize the products that cost money
		if item.TotalPriceCents == 0 {
			continue
		}

		// count seats for a single (batched) line item; itemize everything else
		if item.Product == models.SeatProduct {
			seatCount++
			seatPrice = item.TotalPriceCents
			seatStartTime = item.StartTime
			seatEndTime = item.EndTime
		} else {
			NewInvoiceItemOther(customerID, item.TotalPriceCents, billingInfo.BillingPlan.Currency, item.StartTime, item.EndTime, PrettyDescription(billingInfo.BillingPlan, item.Product))
		}
	}

	// batch seats
	if seatCount > 0 {
		NewInvoiceItemSeats(customerID, seatCount, seatPrice, string(billingInfo.BillingPlan.Currency), seatStartTime, seatEndTime, PrettyDescription(billingInfo.BillingPlan, models.SeatProduct))
	}

	// add tax
	taxPercentage, description := paymentsutil.ComputeTaxPercentage(billingInfo.Country, billingInfo.Region, billingInfo.IsCompany(), time.Now())
	taxRate, err := CreateTaxRate(taxPercentage, description)
	if err != nil {
		return nil, err
	}

	inv, err := CreateInvoice(customerID, billingInfo.BillingMethod.PaymentsDriver, taxRate)
	if err != nil {
		return nil, err
	}

	return inv, nil
}

// FindCustomerInvoices lists a customer's historical invoices
func FindCustomerInvoices(customerID string) []*stripe.Invoice {
	var invoices []*stripe.Invoice
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(customerID),
	}
	it := invoice.List(params)
	for it.Next() {
		invoices = append(invoices, it.Invoice())
	}
	return invoices
}

// PayInvoice triggers Stripe to pay the invoice on behalf of the customer
// use this for customers who pay by card and are automatically charged
func PayInvoice(logger *zap.SugaredLogger, invoiceID string) error {
	_, err := invoice.Pay(invoiceID, nil)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			switch stripeErr.Code {
			case stripe.ErrorCodeCardDeclined:
				logger.Infow("card declined", "invoice_id", invoiceID, "error_msg", stripeErr.Msg)
				return nil
			case stripe.ErrorCodeExpiredCard:
				logger.Infow("card expired", "invoice_id", invoiceID, "error_msg", stripeErr.Msg)
				return nil
			case "invoice_payment_intent_requires_action":
				// stripe will send the customer an email, bring them to a hosted checkout page, and get manual authorization
				// https://stripe.com/docs/error-codes/invoice-payment-intent-requires-action
				// https://stripe.com/docs/billing/subscriptions/payment#handling-action-required
				logger.Infow("card payment requires action", "invoice_id", invoiceID, "error_msg", stripeErr.Msg)
				return nil
			}
		}
		return fmt.Errorf("stripe error when paying invoice: %s", err.Error())
	}
	return nil
}

// SendInvoice triggers Stripe to send the invoice to the customer
// use this for customers who pay by wire and are not automatically charged
func SendInvoice(invoiceID string) error {
	_, err := invoice.SendInvoice(invoiceID, nil)
	if err != nil {
		return fmt.Errorf("stripe error when sending invoice: %s", err.Error())
	}
	return nil
}

// PrettyDescription makes the Stripe invoice more informative
func PrettyDescription(plan *models.BillingPlan, product models.Product) string {
	switch product {
	case models.BaseProduct:
		return fmt.Sprintf("%s (base)", plan.Name)
	case models.SeatProduct:
		return fmt.Sprintf("%s (seats)", plan.Name)
	case models.ProratedSeatProduct:
		return fmt.Sprintf("Seats added mid-period (prorated)")
	case models.ReadOverageProduct:
		return "Read overage (GB)"
	case models.WriteOverageProduct:
		return "Write overage (GB)"
	case models.ScanOverageProduct:
		return "Scan overage (GB)"
	default:
		panic(fmt.Errorf("unknown product '%s'", product))
	}
}
