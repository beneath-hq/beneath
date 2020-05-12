package stripecard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	uuid "github.com/satori/go.uuid"
	stripe "github.com/stripe/stripe-go"
	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/payments/driver"
	"gitlab.com/beneath-hq/beneath/control/payments/driver/stripeutil"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/envutil"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/jsonutil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
)

// StripeCard implements beneath.PaymentsDriver
type StripeCard struct {
	config configSpecification
}

type card struct {
	Brand    string
	Last4    string
	ExpMonth int
	ExpYear  int
}

type address struct {
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    string
}

type billingDetails struct {
	Name    string
	Email   string
	Address *address
}

const (
	maxCardRetries = 4
)

type configSpecification struct {
	StripeSecret string `envconfig:"CONTROL_STRIPE_SECRET" required:"true"`
}

// New initializes a StripeCard object
func New() StripeCard {
	var config configSpecification
	envutil.LoadConfig("beneath", &config)
	stripeutil.InitStripe(config.StripeSecret)
	return StripeCard{
		config: config,
	}
}

// GetHTTPHandlers returns the necessary handlers to implement Stripe card payments
func (c *StripeCard) GetHTTPHandlers() map[string]httputil.AppHandler {
	return map[string]httputil.AppHandler{
		"generate_setup_intent": c.handleGenerateSetupIntent, // this generates a secret that is used in the front-end; then using that secret, front-end sends card info straight to stripe
		"webhook":               c.handleStripeWebhook,       // stripe sends the card setup outcome to this webhook
	}
}

func (c *StripeCard) handleGenerateSetupIntent(w http.ResponseWriter, req *http.Request) error {
	organizationID, err := uuid.FromString(req.URL.Query().Get("organizationID"))
	if err != nil {
		return httputil.NewError(400, "Unable to get organizationID from the request")
	}

	organization := entity.FindOrganization(req.Context(), organizationID)
	if organization == nil {
		return httputil.NewError(400, "Organization not found")
	}

	secret := middleware.GetSecret(req.Context())
	perms := secret.OrganizationPermissions(req.Context(), organizationID)
	if !perms.Admin {
		return httputil.NewError(403, fmt.Sprintf("You are not allowed to perform admin functions in organization %s", organizationID.String()))
	}

	setupIntent := stripeutil.GenerateSetupIntent(organizationID)

	// create json response
	json, err := jsonutil.Marshal(map[string]interface{}{
		"client_secret": setupIntent.ClientSecret,
	})
	if err != nil {
		return httputil.NewError(500, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)

	return nil
}

// Q: what's the right way to return errors for the webhook?
func (c *StripeCard) handleStripeWebhook(w http.ResponseWriter, req *http.Request) error {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.S.Errorf("Error reading request body: %v\\n", err)
		return err
	}

	event := stripe.Event{}

	if err := json.Unmarshal(payload, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.S.Errorf("Failed to parse webhook body json: %v\\n", err.Error())
		return err
	}

	switch event.Type {

	case "setup_intent.succeeded":
		var setupIntent stripe.SetupIntent
		err := json.Unmarshal(event.Data.Raw, &setupIntent)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.S.Errorf("Error parsing webhook JSON: %v\\n", err)
			return err
		}

		organization := entity.FindOrganization(req.Context(), uuid.FromStringOrNil(setupIntent.Metadata["OrganizationID"]))
		if organization == nil {
			panic("organization not found")
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

		// get full paymentMethod object
		paymentMethod := stripeutil.RetrievePaymentMethod(setupIntent.PaymentMethod.ID)

		// Our requests to Stripe differ whether or not the customer is already registered in Stripe
		var customer *stripe.Customer
		if customerID != "" {
			paymentMethod := stripeutil.AttachPaymentMethod(customerID, paymentMethod.ID)
			customer = stripeutil.UpdateCardCustomer(customerID, paymentMethod.BillingDetails.Email, paymentMethod.ID)
		} else {
			customer = stripeutil.CreateCardCustomer(organization.OrganizationID, organization.Name, paymentMethod.BillingDetails.Email, setupIntent.PaymentMethod)
			customerID = customer.ID
		}

		driverPayload := make(map[string]interface{})
		driverPayload["customer_id"] = customerID
		driverPayload["payment_method_id"] = paymentMethod.ID
		driverPayload["brand"] = paymentMethod.Card.Brand
		driverPayload["last4"] = paymentMethod.Card.Last4
		driverPayload["expMonth"] = paymentMethod.Card.ExpMonth
		driverPayload["expYear"] = paymentMethod.Card.ExpYear

		_, err = entity.CreateBillingMethod(req.Context(), organization.OrganizationID, entity.StripeCardDriver, driverPayload)
		if err != nil {
			log.S.Errorf("Error creating billing method: %v\\n", err)
			return httputil.NewError(500, "error creating billing method: %v\\n", err)
		}

	case "invoice.payment_failed":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.S.Errorf("Error parsing webhook JSON: %v\\n", err)
			return err
		}

		customer := stripeutil.RetrieveCustomer(invoice.Customer.ID)
		organizationID := uuid.FromStringOrNil(customer.Metadata["OrganizationID"])
		log.S.Infof("Invoice payment failed on attempt #%d for organization %s", invoice.AttemptCount, organizationID.String())

		// after X failed card retries, if the customer is not an Enterprise customer, downgrade the customer to a Free plan, which will update the users' quotas
		if (*invoice.CollectionMethod == stripe.InvoiceCollectionMethodChargeAutomatically) && (invoice.AttemptCount == maxCardRetries) && (invoice.Paid == false) {
			defaultBillingPlan := entity.FindDefaultBillingPlan(req.Context())

			billingInfo := entity.FindBillingInfo(req.Context(), organizationID)
			if billingInfo == nil {
				panic("could not find organization's billing info")
			}

			if !billingInfo.BillingPlan.MultipleUsers {
				// Q: should we take them off the faulty billing method? mark the billing method as faulty?
				billingInfo.Update(req.Context(), billingInfo.BillingMethodID, defaultBillingPlan.BillingPlanID, billingInfo.Country, &billingInfo.Region, &billingInfo.CompanyName, &billingInfo.TaxNumber) // only changing the billing plan
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.S.Errorf("Error updating Billing Info: %v\\n", err)
					return err
				}
				log.S.Infof("Organization %s downgraded to the Free plan due to payment failure", organizationID.String())
			}
		}

	case "invoice.updated":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.S.Errorf("Error parsing webhook JSON: %v\\n", err)
			return err
		}

		// when invoice goes "past_due" for pay-by-wire customers, log it
		// TODO: waiting on stripe to fix their bug: this "invoice.updated" (nor "invoice.payment_failed") webhook doesn't currently trigger for one-off invoices that are paid by wire when they go past_due
		if (*invoice.CollectionMethod == stripe.InvoiceCollectionMethodSendInvoice) && (invoice.Status == "past_due") && (invoice.Paid == false) {
			customer := stripeutil.RetrieveCustomer(invoice.Customer.ID)
			organizationID := uuid.FromStringOrNil(customer.Metadata["OrganizationID"])

			log.S.Infof("Invoice is past due for organization %s", organizationID)
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		log.S.Errorf("Unexpected event type: %s", event.Type)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// IssueInvoiceForResources implements Payments interface
func (c *StripeCard) IssueInvoiceForResources(billingInfo driver.BillingInfo, billedResources []driver.BilledResource) error {
	inv := stripeutil.CreateStripeInvoice(billingInfo, billedResources)
	stripeutil.PayInvoice(inv.ID)

	return nil
}
