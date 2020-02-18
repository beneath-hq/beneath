package anarchism

import (
	"fmt"
	"net/http"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/control/payments/driver"
	"github.com/beneath-core/internal/middleware"
	"github.com/beneath-core/pkg/envutil"
	"github.com/beneath-core/pkg/httputil"
	"github.com/beneath-core/pkg/log"
	uuid "github.com/satori/go.uuid"
)

// Anarchism implements beneath.PaymentsDriver
type Anarchism struct {
	config configSpecification
}

type configSpecification struct {
	PaymentsAdminSecret string `envconfig:"CONTROL_PAYMENTS_ADMIN_SECRET" required:"true"`
}

// New initializes a Anarchism object
func New() Anarchism {
	var config configSpecification
	envutil.LoadConfig("beneath", &config)
	return Anarchism{
		config: config,
	}
}

// GetHTTPHandlers returns the necessary handlers to implement Stripe card payments
func (a *Anarchism) GetHTTPHandlers() map[string]httputil.AppHandler {
	return map[string]httputil.AppHandler{
		"initialize_customer": a.handleInitializeCustomer,
	}
}

// update a customer's billing info
// this does NOT get called when a customer initially signs up, since we place them on the Anarchy plan in the ...control/entity/user/CreateOrUpdateUser fxn
// this gets called when users downgrade to the Free plan
func (a *Anarchism) handleInitializeCustomer(w http.ResponseWriter, req *http.Request) error {
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

	billingInfo := entity.FindBillingInfo(req.Context(), organizationID)
	if billingInfo == nil {
		return httputil.NewError(400, "billing info not found")
	}

	secret := middleware.GetSecret(req.Context())

	// enterprise plans require a Beneath Payments Admin to cancel the plan
	// non-enterprise plans require organization admin permissions to cancel the plan
	if !billingInfo.BillingPlan.Personal { // checks for enterprise plans
		if secret.GetSecretID().String() != a.config.PaymentsAdminSecret {
			return httputil.NewError(403, fmt.Sprintf("Enterprise plans require a Beneath Payments Admin to cancel"))
		}
	} else {
		perms := secret.OrganizationPermissions(req.Context(), organizationID)
		if !perms.Admin {
			return httputil.NewError(403, fmt.Sprintf("You are not allowed to perform admin functions in organization %s", organizationID.String()))
		}
	}

	driverPayload := billingInfo.DriverPayload // this ensures we retain the customer's Stripe customerID in the event a customer goes from paying->free->paying

	_, err = entity.UpdateBillingInfo(req.Context(), organization.OrganizationID, billingPlan.BillingPlanID, entity.AnarchismDriver, driverPayload)
	if err != nil {
		log.S.Errorf("Error updating billing info: %v\\n", err)
		return httputil.NewError(500, "error updating billing info: %v\\n", err)
	}

	return nil
}

// IssueInvoiceForResources implements Payments interface
func (a *Anarchism) IssueInvoiceForResources(billingInfo driver.BillingInfo, billedResources []driver.BilledResource) error {

	// Amazing... no payment required!
	log.S.Infof("anarchism! organization %s does not pay for its usage so no invoice was sent", billingInfo.GetOrganizationID())

	return nil
}
