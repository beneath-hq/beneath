package stripecard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlab.com/beneath-org/beneath/control/entity"
	"gitlab.com/beneath-org/beneath/control/payments/driver"
	"gitlab.com/beneath-org/beneath/control/payments/driver/stripeutil"
	"gitlab.com/beneath-org/beneath/internal/middleware"
	"gitlab.com/beneath-org/beneath/pkg/envutil"
	"gitlab.com/beneath-org/beneath/pkg/httputil"
	"gitlab.com/beneath-org/beneath/pkg/jsonutil"
	"gitlab.com/beneath-org/beneath/pkg/log"
	uuid "github.com/satori/go.uuid"
	stripe "github.com/stripe/stripe-go"
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
		"get_payment_details":   c.handleGetPaymentDetails,
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

	billingPlanID, err := uuid.FromString(req.URL.Query().Get("billingPlanID"))
	if err != nil {
		return httputil.NewError(400, "Unable to get billingPlanID from the request")
	}

	billingPlan := entity.FindBillingPlan(req.Context(), billingPlanID)
	if billingPlan == nil {
		return httputil.NewError(400, "Billing plan not found")
	}

	if !billingPlan.Personal { // checks for enterprise plans
		return httputil.NewError(403, "Signing up for Enterprise billing plans requires contacting Beneath")
	}

	secret := middleware.GetSecret(req.Context())
	perms := secret.OrganizationPermissions(req.Context(), organizationID)
	if !perms.Admin {
		return httputil.NewError(403, fmt.Sprintf("You are not allowed to perform admin functions in organization %s", organizationID.String()))
	}

	setupIntent := stripeutil.GenerateSetupIntent(organizationID, billingPlanID)

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

		billingPlan := entity.FindBillingPlan(req.Context(), uuid.FromStringOrNil(setupIntent.Metadata["BillingPlanID"]))
		if billingPlan == nil {
			panic("billing plan not found")
		}

		billingInfo := entity.FindBillingInfo(req.Context(), organization.OrganizationID) // existing billing info (to check for existing stripe customer_id below)
		if billingInfo == nil {
			panic("billing info not found")
		}

		paymentMethod := stripeutil.RetrievePaymentMethod(setupIntent.PaymentMethod.ID)

		// Our requests to Stripe differ whether or not the customer is already registered in Stripe
		var customer *stripe.Customer
		if billingInfo.DriverPayload["customer_id"] != nil {
			customerID := billingInfo.DriverPayload["customer_id"].(string)
			paymentMethod := stripeutil.AttachPaymentMethod(customerID, paymentMethod.ID)
			customer = stripeutil.UpdateCardCustomer(customerID, paymentMethod.BillingDetails.Email, paymentMethod.ID)
		} else {
			customer = stripeutil.CreateCardCustomer(organization.OrganizationID, organization.Name, paymentMethod.BillingDetails.Email, setupIntent.PaymentMethod)
		}

		driverPayload := make(map[string]interface{})
		driverPayload["customer_id"] = customer.ID

		_, err = entity.UpdateBillingInfo(req.Context(), organization.OrganizationID, billingPlan.BillingPlanID, entity.StripeCardDriver, driverPayload)
		if err != nil {
			log.S.Errorf("Error updating billing info: %v\\n", err)
			return httputil.NewError(500, "error updating billing info: %v\\n", err)
		}

	case "invoice.payment_failed":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.S.Errorf("Error parsing webhook JSON: %v\\n", err)
			return err
		}

		// after X card retries, shut off the customer's service (aka switch them to the Free billing plan, which will update their users' quotas)
		if (*invoice.CollectionMethod == stripe.InvoiceCollectionMethodChargeAutomatically) && (invoice.Paid == false) && (invoice.AttemptCount == maxCardRetries) {
			organizationID := uuid.FromStringOrNil(invoice.Customer.Metadata["OrganizationID"])
			defaultBillingPlan := entity.FindDefaultBillingPlan(req.Context())

			driverPayload := make(map[string]interface{})
			driverPayload["customer_id"] = invoice.Customer.ID

			_, err = entity.UpdateBillingInfo(req.Context(), organizationID, defaultBillingPlan.BillingPlanID, entity.StripeCardDriver, driverPayload)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.S.Errorf("Error updating Billing Info: %v\\n", err)
				return err
			}
		}

	// using this as a hack for stripe wire failures for one-off invoices
	// TODO: waiting on stripe to fix their bug; this "invoice.updated" (nor "invoice.payment_failed") webhook doesn't currently trigger for one-off invoices that are paid by wire when they go past_due
	case "invoice.updated":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.S.Errorf("Error parsing webhook JSON: %v\\n", err)
			return err
		}

		// when invoice goes "past_due", shut off the customer's service (aka switch them to the Free billing plan, which will update their users' quotas)
		if (*invoice.CollectionMethod == stripe.InvoiceCollectionMethodSendInvoice) && (invoice.Status == "past_due") && (invoice.Paid == false) {
			organizationID := uuid.FromStringOrNil(invoice.Customer.Metadata["OrganizationID"])
			freeBillingPlanID := uuid.FromStringOrNil(entity.FreeBillingPlanID)

			driverPayload := make(map[string]interface{})
			driverPayload["customer_id"] = invoice.Customer.ID

			_, err = entity.UpdateBillingInfo(req.Context(), organizationID, freeBillingPlanID, entity.StripeCardDriver, driverPayload)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.S.Errorf("Error updating Billing Info: %v\\n", err)
				return err
			}
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		log.S.Errorf("Unexpected event type: %s", event.Type)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (c *StripeCard) handleGetPaymentDetails(w http.ResponseWriter, req *http.Request) error {
	secret := middleware.GetSecret(req.Context())

	user := entity.FindUser(req.Context(), secret.GetOwnerID())
	if user == nil {
		return httputil.NewError(400, "user not found")
	}
	organizationID := user.OrganizationID

	perms := secret.OrganizationPermissions(req.Context(), organizationID)
	if !perms.Admin {
		return httputil.NewError(403, fmt.Sprintf("not allowed to perform admin functions in organization %s", organizationID.String()))
	}

	billingInfo := entity.FindBillingInfo(req.Context(), organizationID)
	if billingInfo == nil {
		return httputil.NewError(500, fmt.Sprintf("billing info not found for organization %s", organizationID.String()))
	}

	if billingInfo.PaymentsDriver != "stripecard" {
		return httputil.NewError(400, fmt.Sprintf("the organization is not on the 'stripecard' payments driver"))
	}

	customer := stripeutil.RetrieveCustomer(billingInfo.DriverPayload["customer_id"].(string))
	pm := stripeutil.RetrievePaymentMethod(customer.InvoiceSettings.DefaultPaymentMethod.ID)

	// create json response
	json, err := jsonutil.Marshal(map[string]map[string]interface{}{
		"data": {
			"organizationID": organizationID,
			"card": &card{
				Brand:    string(pm.Card.Brand),
				Last4:    pm.Card.Last4,
				ExpMonth: int(pm.Card.ExpMonth),
				ExpYear:  int(pm.Card.ExpYear),
			},
			"billingDetails": &billingDetails{
				Name: pm.BillingDetails.Name,
				Address: &address{
					Line1:      pm.BillingDetails.Address.Line1,
					Line2:      pm.BillingDetails.Address.Line2,
					City:       pm.BillingDetails.Address.City,
					State:      pm.BillingDetails.Address.State,
					PostalCode: pm.BillingDetails.Address.PostalCode,
					Country:    pm.BillingDetails.Address.Country,
				},
				Email: pm.BillingDetails.Email,
			},
		}})

	if err != nil {
		return httputil.NewError(500, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
	w.WriteHeader(http.StatusOK)

	return nil
}

// IssueInvoiceForResources implements Payments interface
func (c *StripeCard) IssueInvoiceForResources(billingInfo driver.BillingInfo, billedResources []driver.BilledResource) error {
	inv := stripeutil.CreateStripeInvoice(billingInfo, billedResources)
	stripeutil.PayInvoice(inv.ID)

	return nil
}
