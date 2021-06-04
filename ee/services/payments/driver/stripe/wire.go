package stripe

import (
	"fmt"
	"net/http"

	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"

	"github.com/beneath-hq/beneath/ee/models"
	"github.com/beneath-hq/beneath/ee/services/billing"
	"github.com/beneath-hq/beneath/ee/services/payments/driver"
	"github.com/beneath-hq/beneath/ee/services/payments/driver/stripe/stripeutil"
	"github.com/beneath-hq/beneath/pkg/httputil"
	"github.com/beneath-hq/beneath/services/middleware"
	"github.com/beneath-hq/beneath/services/organization"
	"github.com/beneath-hq/beneath/services/permissions"
)

// WireDriver implements payments.Driver
type WireDriver struct {
	Logger        *zap.SugaredLogger
	Billing       *billing.Service
	Organizations *organization.Service
	Permissions   *permissions.Service
}

// WireOptions for WireDriver
type WireOptions struct {
	StripeSecret string `mapstructure:"stripe_secret"`
}

func init() {
	driver.AddDriver(driver.StripeWire, newStripeWire)
}

func newStripeWire(logger *zap.Logger, billing *billing.Service, organizations *organization.Service, permissions *permissions.Service, optsMap map[string]interface{}) (driver.Driver, error) {
	// load options
	var opts WireOptions
	err := mapstructure.Decode(optsMap, &opts)
	if err != nil {
		return nil, fmt.Errorf("error decoding bigquery options: %s", err.Error())
	}

	// init stripe
	stripeutil.InitStripe(opts.StripeSecret)

	return &WireDriver{
		Logger:        logger.Named("stripe.wire").Sugar(),
		Billing:       billing,
		Organizations: organizations,
		Permissions:   permissions,
	}, nil
}

// GetHTTPHandlers returns the necessary handlers to implement Stripe card payments
func (d *WireDriver) GetHTTPHandlers() map[string]httputil.AppHandler {
	return map[string]httputil.AppHandler{
		"initialize_customer": d.handleInitializeCustomer,
	}
}

// IssueInvoiceForResources implements Payments interface
func (d *WireDriver) IssueInvoiceForResources(bi *models.BillingInfo, resources []*models.BilledResource) error {
	invoice, err := stripeutil.CreateStripeInvoice(bi, resources)
	if err != nil {
		return err
	}

	err = stripeutil.SendInvoice(invoice.ID)
	if err != nil {
		return err
	}

	return nil
}

// handleInitializeCustomer can be called by Beneath masters to create/update a customer's wire billing info
// TODO: make this idempotent -- there shouldn't be multiple wire billing methods
func (d *WireDriver) handleInitializeCustomer(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()

	organizationID, err := uuid.FromString(req.URL.Query().Get("organization_id"))
	if err != nil {
		return httputil.NewError(400, "couldn't get organization_id from the request")
	}

	organization := d.Organizations.FindOrganization(ctx, organizationID)
	if organization == nil {
		return httputil.NewError(400, "organization not found")
	}

	emailAddress := req.URL.Query().Get("email_address")
	if emailAddress == "" {
		return httputil.NewError(400, "couldn't get email_address from the request")
	}

	// Beneath will call the function from an admin panel (after a customer discussion)
	secret := middleware.GetSecret(ctx)
	if !secret.IsMaster() {
		return httputil.NewError(403, fmt.Sprintf("caller must be a Beneath master to enable payment by wire"))
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

	// Our requests to Stripe differ whether or not the customer is already registered in Stripe
	if customerID != "" {
		_, err := stripeutil.UpdateWireCustomer(customerID, emailAddress)
		if err != nil {
			return err
		}
	} else {
		customer, err := stripeutil.CreateWireCustomer(organization.OrganizationID, organization.Name, emailAddress)
		if err != nil {
			return err
		}
		customerID = customer.ID
	}

	driverPayload := make(map[string]interface{})
	driverPayload["customer_id"] = customerID
	driverPayload["email_address"] = emailAddress

	bm, err := d.Billing.CreateBillingMethod(ctx, organization.OrganizationID, driver.StripeWire, driverPayload)
	if err != nil {
		return httputil.NewError(400, "error creating billing method: %s", err.Error())
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
