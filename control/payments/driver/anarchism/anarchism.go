package anarchism

import (
	"fmt"
	"net/http"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/payments/driver"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/envutil"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
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
// TODO: this function should be deleted and moved to a resolver for UpdateBillingInfo()
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

	bm := entity.FindBillingMethodByOrganizationAndDriver(req.Context(), organizationID, entity.AnarchismDriver)
	if bm == nil {
		panic("error!")
	}

	_, err = billingInfo.Update(req.Context(), bm.BillingMethodID, billingPlan.BillingPlanID)
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
