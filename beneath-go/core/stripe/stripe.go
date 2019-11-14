package stripe

import (
	"github.com/beneath-core/beneath-go/core/log"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/invoice"
	"github.com/stripe/stripe-go/invoiceitem"
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
func CreateCustomer(emailAddress string, paymentMethod string) *stripe.Customer {
	params := &stripe.CustomerParams{
		Email:         stripe.String(emailAddress),
		PaymentMethod: stripe.String(paymentMethod),
	}
	customer, err := customer.New(params)
	if err != nil {
		log.S.Errorf("Stripe error: %s", err.Error())
	}

	return customer
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
func CreateInvoice(customerID string, paymentMethod string) *stripe.Invoice {
	collectionMethod := ""
	if paymentMethod == "card" {
		collectionMethod = string(stripe.InvoiceCollectionMethodChargeAutomatically)
	} else if paymentMethod == "wire" {
		collectionMethod = string(stripe.InvoiceCollectionMethodSendInvoice)
	} else {
		panic("unrecognized payment method")
	}

	params := &stripe.InvoiceParams{
		Customer:         stripe.String(customerID),
		CollectionMethod: stripe.String(collectionMethod),
		DaysUntilDue:     stripe.Int64(daysUntilInvoiceDue),
	}

	invoice, err := invoice.New(params)
	if err != nil {
		log.S.Errorf("Stripe error: %s", err.Error())
	}

	return invoice
}

// SendInvoice triggers Stripe to send the invoice to the customer
func SendInvoice(invoiceID string) *stripe.Invoice {
	invoice, err := invoice.SendInvoice(invoiceID, nil)
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
