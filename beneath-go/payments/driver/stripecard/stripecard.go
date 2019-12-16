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
			customer = stripeutil.CreateCardCustomer(organization.Name, paymentMethod.BillingDetails.Email, setupIntent.PaymentMethod)
		}

		driverPayload := make(map[string]interface{})
		driverPayload["customer_id"] = customer.ID

		_, err = entity.UpdateBillingInfo(req.Context(), organization.OrganizationID, billingPlan.BillingPlanID, entity.StripeCardDriver, driverPayload)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.S.Errorf("Error updating Billing Info: %v\\n", err)
			return err
		}
	// case "setup_intent.created":
	// 	// do nothing
	// case "payment_method.attached":
	// 	// do nothing
	// case "customer.created":
	// 	// do nothing
	// case "customer.updated":
	// 	// do nothing
	// case "setup_intent.setup_failed":
	// TODO: handle this case in the front-end and leave this webhook event alone
	case "invoice.payment_failed": // use this for stripe card failures
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.S.Errorf("Error parsing webhook JSON: %v\\n", err)
			return err
		}

		if (*invoice.CollectionMethod == stripe.InvoiceCollectionMethodChargeAutomatically) && (invoice.Paid == false) && (invoice.AttemptCount == maxCardRetries) {
			// TODO: shut off the card customer's service
		}
	// case "invoice.payment_succeeded":
	//	// Q: do we want this for anything?
	case "invoice.updated": // use this for stripe wire failures
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.S.Errorf("Error parsing webhook JSON: %v\\n", err)
			return err
		}

		// shut off a wire customer's service when it is late on payment
		if (*invoice.CollectionMethod == stripe.InvoiceCollectionMethodSendInvoice) && (invoice.Status == "past_due") && (invoice.Paid == false) {
			// TODO: shut off the wire customer's service; waiting on Stripe to fix their bug (this currently isn't implemented for our setup)
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
		},
	})

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
