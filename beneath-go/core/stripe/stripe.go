package stripe

import (
	"github.com/beneath-core/beneath-go/core/log"
	uuid "github.com/satori/go.uuid"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/invoice"
	"github.com/stripe/stripe-go/invoiceitem"
	"github.com/stripe/stripe-go/paymentmethod"
	"github.com/stripe/stripe-go/setupintent"
)

// Q: stripe.Key needs to be set. Should I set it before each action? Or assume it's set globally (per InitClient), didn't expire, and we're good.
const (
	daysUntilInvoiceDue = 30
)

// InitClient sets our Stripe API key
func InitClient(stripeKey string) {
	stripe.Key = stripeKey
}

// CreateCustomer registers the customer with Stripe
// TODO when we collect more customer data like billing address: func CreateCustomer(customerData CustomerData paymentMethod string) *stripe.Customer {
func CreateCustomer(organizationName string, emailAddress string, paymentMethodID string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Name:          stripe.String(organizationName),
		Email:         stripe.String(emailAddress),
		PaymentMethod: stripe.String(paymentMethodID),
	}
	customer, err := customer.New(params)
	if err != nil {
		log.S.Errorf("Stripe error: %s", err.Error())
		return customer, err
	}

	return customer, nil
}

// CreateSetupIntent gets ready for a customer to add credit card information
func CreateSetupIntent(organizationID uuid.UUID, billingPlanID uuid.UUID) *stripe.SetupIntent {
	params := &stripe.SetupIntentParams{
		PaymentMethodTypes: []*string{stripe.String("card")},
		Usage:              stripe.String(string(stripe.SetupIntentUsageOffSession)),
	}
	params.AddMetadata("OrganizationID", organizationID.String())
	params.AddMetadata("BillingPlanID", billingPlanID.String())
	setupIntent, err := setupintent.New(params)
	if err != nil {
		log.S.Errorf("Stripe error: %s", err.Error())
	}

	return setupIntent
}

// RetrievePaymentMethod returns a payment method, with which we access billing_details
func RetrievePaymentMethod(paymentMethodID string) *stripe.PaymentMethod {
	paymentMethod, _ := paymentmethod.Get(paymentMethodID, nil)

	return paymentMethod
}

// NewInvoiceItem adds an invoice line item
// Q: do we want to add any metadata to these line items?
// other descriptors to add: period, quantity, tax rates
func NewInvoiceItem(customerID string, amount int64, currency string, description string) *stripe.InvoiceItem {
	params := &stripe.InvoiceItemParams{
		Customer:    stripe.String(customerID),
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(string(currency)),
		Description: stripe.String(description),
	}
	invoiceItem, err := invoiceitem.New(params)
	if err != nil {
		log.S.Errorf("Stripe error: %s", err.Error())
	}

	return invoiceItem
}

// CreateInvoice creates an invoice
// there must exist invoice line items for the given customer before this function is called
func CreateInvoice(customerID string, paymentMethodType string) *stripe.Invoice {
	params := &stripe.InvoiceParams{}
	if paymentMethodType == "card" {
		paymentMethodID := GetCustomerRecentPaymentMethodID(customerID, paymentMethodType)
		params = &stripe.InvoiceParams{
			Customer:             stripe.String(customerID),
			CollectionMethod:     stripe.String(string(stripe.InvoiceCollectionMethodChargeAutomatically)),
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		}
	} else if paymentMethodType == "wire" {
		params = &stripe.InvoiceParams{
			Customer:         stripe.String(customerID),
			DaysUntilDue:     stripe.Int64(daysUntilInvoiceDue),
			CollectionMethod: stripe.String(string(stripe.InvoiceCollectionMethodSendInvoice)),
		}
	} else {
		panic("unrecognized payment method")
	}

	invoice, err := invoice.New(params)
	if err != nil {
		log.S.Errorf("Stripe error: %s", err.Error())
	}

	return invoice
}

// IssueInvoice calls the relevant "SendInvoice" or "PayInvoice" function
func IssueInvoice(invoiceID string, paymentMethodType string) {
	if paymentMethodType == "wire" {
		SendInvoice(invoiceID)
	} else if paymentMethodType == "card" {
		PayInvoice(invoiceID)
	} else {
		panic("the organization does not have a payment method set")
	}
}

// SendInvoice triggers Stripe to send the invoice to the customer
// use this for customers who pay by wire and are not automatically charged
func SendInvoice(invoiceID string) *stripe.Invoice {
	invoice, err := invoice.SendInvoice(invoiceID, nil)
	if err != nil {
		log.S.Errorf("Stripe error: %s", err.Error())
	}

	return invoice
}

// PayInvoice triggers Stripe to pay the invoice on behalf of the customer
// use this for customers who pay by card and are automatically charged
func PayInvoice(invoiceID string) *stripe.Invoice {
	invoice, err := invoice.Pay(invoiceID, nil)
	if err != nil {
		log.S.Errorf("Stripe error: %s", err.Error())
	}

	return invoice
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

// GetCustomerRecentPaymentMethodID returns the customer's most recent payment method ID for a given payment method type
func GetCustomerRecentPaymentMethodID(customerID string, paymentMethodType string) string {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String(paymentMethodType),
	}

	i := paymentmethod.List(params)
	created := int64(0)
	recentPaymentMethodID := ""

	for i.Next() {
		paymentMethod := i.PaymentMethod()

		if paymentMethod.Created > created {
			created = paymentMethod.Created
			recentPaymentMethodID = paymentMethod.ID
		}
	}

	return recentPaymentMethodID
}
