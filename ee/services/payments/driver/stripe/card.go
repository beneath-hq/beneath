package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
	"github.com/stripe/stripe-go"
	"go.uber.org/zap"

	"gitlab.com/beneath-hq/beneath/ee/models"
	"gitlab.com/beneath-hq/beneath/ee/services/billing"
	"gitlab.com/beneath-hq/beneath/ee/services/payments/driver"
	"gitlab.com/beneath-hq/beneath/ee/services/payments/driver/stripe/stripeutil"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/jsonutil"
	"gitlab.com/beneath-hq/beneath/services/middleware"
	"gitlab.com/beneath-hq/beneath/services/organization"
	"gitlab.com/beneath-hq/beneath/services/permissions"
)

const (
	maxCardRetries = 4
)

// CardDriver implements driver.Driver
type CardDriver struct {
	Logger        *zap.SugaredLogger
	Billing       *billing.Service
	Organizations *organization.Service
	Permissions   *permissions.Service
}

// CardOptions for CardDriver
type CardOptions struct {
	StripeSecret string `mapstructure:"stripe_secret"`
}

func init() {
	driver.AddDriver(driver.StripeCard, newCardDriver)
}

func newCardDriver(logger *zap.Logger, billing *billing.Service, organizations *organization.Service, permissions *permissions.Service, optsMap map[string]interface{}) (driver.Driver, error) {
	// load options
	var opts CardOptions
	err := mapstructure.Decode(optsMap, &opts)
	if err != nil {
		return nil, fmt.Errorf("error decoding bigquery options: %s", err.Error())
	}

	// init stripe (todo: get rid of globals)
	stripeutil.InitStripe(opts.StripeSecret)

	return &CardDriver{
		Logger:        logger.Named("stripe.card").Sugar(),
		Billing:       billing,
		Organizations: organizations,
		Permissions:   permissions,
	}, nil
}

// GetHTTPHandlers returns the necessary handlers to implement Stripe card payments
func (d *CardDriver) GetHTTPHandlers() map[string]httputil.AppHandler {
	return map[string]httputil.AppHandler{
		"generate_setup_intent": d.handleGenerateSetupIntent,
		"webhook":               d.handleStripeWebhook,
	}
}

// IssueInvoiceForResources implements Payments interface
func (d *CardDriver) IssueInvoiceForResources(bi *models.BillingInfo, resources []*models.BilledResource) error {
	invoice, err := stripeutil.CreateStripeInvoice(bi, resources)
	if err != nil {
		return err
	}

	err = stripeutil.PayInvoice(d.Logger, invoice.ID)
	if err != nil {
		return err
	}

	return nil
}

// handleGenerateSetupIntent generates a secret that the front-end uses to send card info straight to Stripe
func (d *CardDriver) handleGenerateSetupIntent(w http.ResponseWriter, req *http.Request) error {
	organizationID, err := uuid.FromString(req.URL.Query().Get("organizationID"))
	if err != nil {
		return httputil.NewError(400, "Unable to get organizationID from the request")
	}

	organization := d.Organizations.FindOrganization(req.Context(), organizationID)
	if organization == nil {
		return httputil.NewError(400, "Organization not found")
	}

	secret := middleware.GetSecret(req.Context())
	perms := d.Permissions.OrganizationPermissionsForSecret(req.Context(), secret, organizationID)
	if !perms.Admin {
		return httputil.NewError(403, fmt.Sprintf("You are not allowed to perform admin functions in organization %s", organizationID.String()))
	}

	setupIntent, err := stripeutil.GenerateSetupIntent(organizationID)
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// create json response
	json, err := jsonutil.Marshal(map[string]interface{}{
		"client_secret": setupIntent.ClientSecret,
	})
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)

	return nil
}

func (d *CardDriver) handleStripeWebhook(w http.ResponseWriter, req *http.Request) error {
	// TODO: Add check for Stripe's signature on each webhook event. See: https://stripeutil.com/docs/webhooks/signatures
	// get payload
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		d.Logger.Errorf("webhook: error reading request body: %s", err.Error())
		return err
	}

	// parse event
	event := &stripe.Event{}
	if err := json.Unmarshal(payload, event); err != nil {
		d.Logger.Errorf("webhook: failed to parse body: %s", err.Error())
		return err
	}

	// handle event
	switch event.Type {
	case "setup_intent.succeeded":
		err = d.handleSetupIntentSucceeded(req.Context(), w, event)
	case "invoice.payment_failed":
		err = d.handlePaymentFailed(req.Context(), w, event)
	default:
		w.WriteHeader(http.StatusOK)
		d.Logger.Errorf("unexpected event type: %s", event.Type)
		return nil
	}

	if err != nil {
		d.Logger.Errorf("webhook handler error for event '%s': %s", event.Type, err.Error())
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (d *CardDriver) handleSetupIntentSucceeded(ctx context.Context, w http.ResponseWriter, event *stripe.Event) error {
	setupIntent := &stripe.SetupIntent{}
	err := json.Unmarshal(event.Data.Raw, setupIntent)
	if err != nil {
		return fmt.Errorf("error parsing webhook json: %s", err.Error())
	}

	organizationID := uuid.FromStringOrNil(setupIntent.Metadata["OrganizationID"])
	organization := d.Organizations.FindOrganization(ctx, organizationID)
	if organization == nil {
		return fmt.Errorf("organization '%s' not found", setupIntent.Metadata["OrganizationID"])
	}

	billingMethods := d.Billing.FindBillingMethodsByOrganization(ctx, organization.OrganizationID)

	// if customer has already been registered with stripe, get customerID from driver_payload
	customerID := ""
	for _, bm := range billingMethods {
		if bm.PaymentsDriver == driver.StripeCard || bm.PaymentsDriver == driver.StripeWire {
			customerID = bm.DriverPayload["customer_id"].(string)
			break
		}
	}

	// get full paymentMethod object
	paymentMethod, err := stripeutil.RetrievePaymentMethod(setupIntent.PaymentMethod.ID)
	if err != nil {
		return err
	}

	// Our requests to Stripe differ whether or not the customer is already registered in Stripe
	var customer *stripe.Customer
	if customerID != "" {
		paymentMethod, err := stripeutil.AttachPaymentMethod(customerID, paymentMethod.ID)
		if err != nil {
			return err
		}

		customer, err = stripeutil.UpdateCardCustomer(customerID, paymentMethod.BillingDetails.Email, paymentMethod.ID)
		if err != nil {
			return err
		}
	} else {
		customer, err = stripeutil.CreateCardCustomer(organization.OrganizationID, organization.Name, paymentMethod.BillingDetails.Email, setupIntent.PaymentMethod)
		if err != nil {
			return err
		}
		customerID = customer.ID
	}

	driverPayload := map[string]interface{}{
		"customer_id":       customerID,
		"payment_method_id": paymentMethod.ID,
		"brand":             paymentMethod.Card.Brand,
		"last4":             paymentMethod.Card.Last4,
		"expMonth":          paymentMethod.Card.ExpMonth,
		"expYear":           paymentMethod.Card.ExpYear,
	}

	bm, err := d.Billing.CreateBillingMethod(ctx, organization.OrganizationID, driver.StripeCard, driverPayload)
	if err != nil {
		return err
	}

	bi := d.Billing.FindBillingInfoByOrganization(ctx, organization.OrganizationID)
	if bi == nil {
		return fmt.Errorf("Existing billing info not found for organization %s", organizationID.String())
	}

	err = d.Billing.UpdateBillingMethod(ctx, bi, bm)
	if err != nil {
		return err
	}

	return nil
}

func (d *CardDriver) handlePaymentFailed(ctx context.Context, w http.ResponseWriter, event *stripe.Event) error {
	invoice := &stripe.Invoice{}
	err := json.Unmarshal(event.Data.Raw, invoice)
	if err != nil {
		return fmt.Errorf("error parsing webhook JSON: %s", err.Error())
	}

	customer, err := stripeutil.RetrieveCustomer(invoice.Customer.ID)
	if err != nil {
		return err
	}

	organizationID := uuid.FromStringOrNil(customer.Metadata["OrganizationID"])
	d.Logger.Infof("invoice payment failed on attempt #%d for organization %s", invoice.AttemptCount, organizationID.String())

	// after X failed card retries, if the customer is not an multi-user customer, downgrade the customer to a free plan, which will update the users' quotas
	if (invoice.CollectionMethod != nil && *invoice.CollectionMethod == stripe.InvoiceCollectionMethodChargeAutomatically) && (invoice.AttemptCount >= maxCardRetries) && (invoice.Paid == false) {
		defaultPlan := d.Billing.GetDefaultBillingPlan(ctx)

		bi := d.Billing.FindBillingInfoByOrganization(ctx, organizationID)
		if bi == nil {
			return fmt.Errorf("could not find billing info for organization '%s'", organizationID.String())
		}

		if !bi.BillingPlan.MultipleUsers {
			err := d.Billing.UpdateBillingPlan(ctx, bi, defaultPlan)
			if err != nil {
				return err
			}
			d.Logger.Infof("organization %s downgraded to the default plan due to payment failure", organizationID.String())
		}
	}

	return nil
}
