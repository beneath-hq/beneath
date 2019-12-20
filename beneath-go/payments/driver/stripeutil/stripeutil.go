package stripeutil

import (
	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core/log"
	uuid "github.com/satori/go.uuid"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/invoice"
	"github.com/stripe/stripe-go/invoiceitem"
	"github.com/stripe/stripe-go/paymentmethod"
	"github.com/stripe/stripe-go/setupintent"
)

const (
	daysUntilInvoiceDue = 10
)

// InitStripe sets our Stripe API key
func InitStripe(stripeKey string) {
	stripe.Key = stripeKey
}

// GenerateSetupIntent gets ready for a customer to add credit card information
func GenerateSetupIntent(organizationID uuid.UUID, billingPlanID uuid.UUID) *stripe.SetupIntent {
	params := &stripe.SetupIntentParams{
		PaymentMethodTypes: stripe.StringSlice([]string{string(stripe.PaymentMethodTypeCard)}),
		Usage:              stripe.String(string(stripe.SetupIntentUsageOffSession)),
	}
	params.AddMetadata("OrganizationID", organizationID.String())
	params.AddMetadata("BillingPlanID", billingPlanID.String())

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

// NewInvoiceItem adds an invoice line item
// TODO: do we want to add any metadata to these line items?
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
		panic("stripe error when adding invoice item")
	}

	return invoiceItem
}

// CreateInvoice creates an invoice
// there must exist invoice line items for the given customer before this function is called
func CreateInvoice(customerID string, paymentsDriver entity.PaymentsDriver) *stripe.Invoice {
	params := &stripe.InvoiceParams{}
	if paymentsDriver == entity.StripeCardDriver {
		params = &stripe.InvoiceParams{
			Customer:         stripe.String(customerID),
			CollectionMethod: stripe.String(string(stripe.InvoiceCollectionMethodChargeAutomatically)),
			AutoAdvance:      stripe.Bool(true), // this tells Stripe to re-try failed payments using their "Smart Retries" feature; it will retry up to 4 times
		}
	} else if paymentsDriver == entity.StripeWireDriver {
		params = &stripe.InvoiceParams{
			Customer:         stripe.String(customerID),
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
				log.S.Infof("card declined when attempting to pay invoice for stripe customer %s", stripeErr.PaymentIntent.Customer.ID)
			case stripe.ErrorCodeExpiredCard:
				log.S.Infof("card declined when attempting to pay invoice for stripe customer %s", stripeErr.PaymentIntent.Customer.ID)
			case "invoice_payment_intent_requires_action":
				// not sure how common this will be
				// this will happen when a customer needs to manually authorize each payment
				// TODO: need to send email to customer, bring them to checkout page, and get manual authorization
				panic("stripe error when paying invoice")
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
func PrettyDescription(product entity.Product) string {
	// TODO: upgrade descriptions
	switch product {
	case entity.SeatProduct:
		return string(product)
	case entity.ReadOverageProduct:
		return string(product)
	case entity.WriteOverageProduct:
		return string(product)
	default:
		panic("unknown product")
	}
}
