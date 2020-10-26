package stripewire

import (
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"
	stripe "github.com/stripe/stripe-go"
	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/payments/driver/stripeutil"
	"gitlab.com/beneath-hq/beneath/ee/infrastructure/payments/driver"
	"gitlab.com/beneath-hq/beneath/pkg/envutil"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/services/middleware"
)

// StripeWire implements beneath.PaymentsDriver
type StripeWire struct {
	config configSpecification
}

type configSpecification struct {
	StripeSecret string `envconfig:"CONTROL_STRIPE_SECRET" required:"true"`
}

// New initializes a StripeWire object
func New() StripeWire {
	var config configSpecification
	envutil.LoadConfig("beneath", &config)
	stripeutil.InitStripe(config.StripeSecret)
	return StripeWire{
		config: config,
	}
}

// GetHTTPHandlers returns the necessary handlers to implement Stripe card payments
func (s *StripeWire) GetHTTPHandlers() map[string]httputil.AppHandler {
	return map[string]httputil.AppHandler{
		"initialize_customer": s.handleInitializeCustomer,
		// "webhook":               handleStripeWebhook,       // TODO: when a customer pays by wire, check to see if any important Stripe events are emitted via webhook
	}
}

// create/update a customer's billing info and Stripe registration
func (s *StripeWire) handleInitializeCustomer(w http.ResponseWriter, req *http.Request) error {
	organizationID, err := uuid.FromString(req.URL.Query().Get("organizationID"))
	if err != nil {
		return httputil.NewError(400, "couldn't get organizationID from the request")
	}

	organization := entity.FindOrganization(req.Context(), organizationID)
	if organization == nil {
		return httputil.NewError(400, "organization not found")
	}

	emailAddress := req.URL.Query().Get("emailAddress")
	if emailAddress == "" {
		return httputil.NewError(400, "couldn't get emailAddress from the request")
	}

	// Beneath will call the function from an admin panel (after a customer discussion)
	secret := middleware.GetSecret(req.Context())
	if !secret.IsMaster() {
		return httputil.NewError(403, fmt.Sprintf("enabling payment by wire requires a Beneath Payments Admin"))
	}

	billingMethods := entity.FindBillingMethodsByOrganization(req.Context(), organization.OrganizationID)

	// if customer has already been registered with stripe, get customerID from driver_payload
	customerID := ""
	for _, bm := range billingMethods {
		if bm.PaymentsDriver == entity.StripeCardDriver || bm.PaymentsDriver == entity.StripeWireDriver {
			customerID = bm.DriverPayload["customer_id"].(string)
			break
		}
	}

	// Our requests to Stripe differ whether or not the customer is already registered in Stripe
	var customer *stripe.Customer
	if customerID != "" {
		// customer is already registered with Stripe
		stripeutil.UpdateWireCustomer(customerID, emailAddress)
	} else {
		// customer needs to be registered with stripe
		customer = stripeutil.CreateWireCustomer(organization.OrganizationID, organization.Name, emailAddress)
		customerID = customer.ID
	}

	driverPayload := make(map[string]interface{})
	driverPayload["customer_id"] = customerID

	_, err = entity.CreateBillingMethod(req.Context(), organization.OrganizationID, entity.StripeWireDriver, driverPayload)
	if err != nil {
		log.S.Errorf("Error creating billing method: %v\\n", err)
		return httputil.NewError(500, "error creating billing method: %v\\n", err)
	}

	return nil
}

// IssueInvoiceForResources implements Payments interface
func (s *StripeWire) IssueInvoiceForResources(billingInfo driver.BillingInfo, billedResources []driver.BilledResource) error {
	inv := stripeutil.CreateStripeInvoice(billingInfo, billedResources)
	stripeutil.SendInvoice(inv.ID)

	return nil
}
