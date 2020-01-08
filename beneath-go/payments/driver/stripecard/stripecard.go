package stripecard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/beneath-core/beneath-go/payments/driver/stripeutil"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/httputil"
	"github.com/beneath-core/beneath-go/core/jsonutil"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/core/middleware"
	uuid "github.com/satori/go.uuid"
	stripe "github.com/stripe/stripe-go"
)

// StripeCard implements beneath.PaymentsDriver
type StripeCard struct {
}

type card struct {
	Brand string
	Last4 string
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
	Phone   string
	Address *address
}

const (
	maxCardRetries = 4
)

type configStripe struct {
	StripeSecret string `envconfig:"CONTROL_STRIPE_SECRET" required:"true"`
}

// New initializes a StripeCard object
func New() StripeCard {
	var config configStripe
	core.LoadConfig("beneath", &config)
	stripeutil.InitStripe(config.StripeSecret)

	return StripeCard{}
}

// GetHTTPHandlers returns the necessary handlers to implement Stripe card payments
func (c *StripeCard) GetHTTPHandlers() map[string]httputil.AppHandler {
	return map[string]httputil.AppHandler{
		"generate_setup_intent": handleGenerateSetupIntent, // this generates a secret that is used in the front-end; then using that secret, front-end sends card info straight to stripe
		"webhook":               handleStripeWebhook,       // stripe sends the card setup outcome to this webhook
		"get_payment_details":   handleGetPaymentDetails,
	}
}

func handleGenerateSetupIntent(w http.ResponseWriter, req *http.Request) error {
	organizationID, err := uuid.FromString(req.URL.Query().Get("organizationID"))
	if err != nil {
		return httputil.NewError(400, "couldn't get organizationID from the request")
	}

	organization := entity.FindOrganization(req.Context(), organizationID)
	if organization == nil {
		return httputil.NewError(400, "organization not found")
	}

	billingPlanID, err := uuid.FromString(req.URL.Query().Get("billingPlanID"))
	if err != nil {
		return httputil.NewError(400, "couldn't get billingPlanID from the request")
	}

	billingPlan := entity.FindBillingPlan(req.Context(), billingPlanID)
	if billingPlan == nil {
		return httputil.NewError(400, "billing plan not found")
	}

	secret := middleware.GetSecret(req.Context())
	perms := secret.OrganizationPermissions(req.Context(), organizationID)
	if !perms.Admin {
		return httputil.NewError(403, fmt.Sprintf("not allowed to perform admin functions in organization %s", organizationID.String()))
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
func handleStripeWebhook(w http.ResponseWriter, req *http.Request) error {
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

		paymentMethod := stripeutil.RetrievePaymentMethod(setupIntent.PaymentMethod.ID)

		// Our requests to Stripe differ whether or not the customer is already registered in Stripe
		var customer *stripe.Customer
		billingInfo := entity.FindBillingInfo(req.Context(), organization.OrganizationID)
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
			w.WriteHeader(http.StatusInternalServerError)
			log.S.Errorf("Error updating Billing Info: %v\\n", err)
			return err
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

	// case "invoice.payment_succeeded":
	// Q: do we want this^ for anything?

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

func handleGetPaymentDetails(w http.ResponseWriter, req *http.Request) error {
	secret := middleware.GetSecret(req.Context())

	user := entity.FindUser(req.Context(), secret.GetOwnerID())
	if user == nil {
		return httputil.NewError(400, "user not found")
	}
	organizationID := *user.OrganizationID

	perms := secret.OrganizationPermissions(req.Context(), organizationID)
	if !perms.Admin {
		return httputil.NewError(403, fmt.Sprintf("not allowed to perform admin functions in organization %s", organizationID.String()))
	}

	billingInfo := entity.FindBillingInfo(req.Context(), organizationID)
	if billingInfo == nil {
		return httputil.NewError(400, fmt.Sprintf("billing info not found for organization %s", organizationID.String()))
	}

	customer := stripeutil.RetrieveCustomer(billingInfo.DriverPayload["customer_id"].(string))
	pm := stripeutil.RetrievePaymentMethod(customer.InvoiceSettings.DefaultPaymentMethod.ID)

	// create json response
	json, err := jsonutil.Marshal(map[string]map[string]interface{}{
		"data": {
			"organization_id": organizationID,
			"card": &card{
				Brand: string(pm.Card.Brand),
				Last4: pm.Card.Last4,
			},
			"billing_details": &billingDetails{
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
				Phone: pm.BillingDetails.Phone,
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
func (c *StripeCard) IssueInvoiceForResources(billingInfo *entity.BillingInfo, billedResources []*entity.BilledResource) error {
	if billingInfo.DriverPayload["customer_id"] == nil {
		panic("stripe customer id is not set")
	}

	for _, item := range billedResources {
		// only itemize the products that cost money (i.e. don't itemize the included Reads and Writes)
		if item.TotalPriceCents > 0 {
			stripeutil.NewInvoiceItem(billingInfo.DriverPayload["customer_id"].(string), int64(item.TotalPriceCents), string(billingInfo.BillingPlan.Currency), stripeutil.PrettyDescription(item.Product))
		}
	}

	inv := stripeutil.CreateInvoice(billingInfo.DriverPayload["customer_id"].(string), billingInfo.PaymentsDriver)

	stripeutil.PayInvoice(inv.ID)

	return nil
}
