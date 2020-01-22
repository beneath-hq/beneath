package anarchism

import (
	"fmt"
	"net/http"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core/httputil"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/core/middleware"
	uuid "github.com/satori/go.uuid"
)

// Anarchism implements beneath.PaymentsDriver
type Anarchism struct{}

// New initializes a Anarchism object
func New() Anarchism {
	return Anarchism{}
}

// GetHTTPHandlers returns the necessary handlers to implement Stripe card payments
func (a *Anarchism) GetHTTPHandlers() map[string]httputil.AppHandler {
	return map[string]httputil.AppHandler{
		"initialize_customer": handleInitializeCustomer,
	}
}

// update a customer's billing info
// this does NOT get called when a customer initially signs up, since we place them on the Anarchy plan in the ...control/entity/user/CreateOrUpdateUser fxn
// we will call this manually when users downgrade to the Free plan
func handleInitializeCustomer(w http.ResponseWriter, req *http.Request) error {
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

	billingInfo := entity.FindBillingInfo(req.Context(), organizationID)
	if billingInfo == nil {
		return httputil.NewError(400, "billingInfo not found")
	}

	driverPayload := billingInfo.DriverPayload // this ensures we retain the customer's Stripe customerID in the event a customer goes from paying->free->paying

	_, err = entity.UpdateBillingInfo(req.Context(), organization.OrganizationID, billingPlan.BillingPlanID, entity.AnarchismDriver, driverPayload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.S.Errorf("Error updating Billing Info: %v\\n", err)
		return err
	}

	if billingPlan.Personal {
		err = organization.UpdatePersonalStatus(req.Context(), true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.S.Errorf("Error updating organization: %v\\n", err)
			return err
		}
	}

	return nil
}

// IssueInvoiceForResources implements Payments interface
func (a *Anarchism) IssueInvoiceForResources(billingInfo *entity.BillingInfo, billedResources []*entity.BilledResource) error {

	// Amazing... no payment required!
	log.S.Infof("anarchism! organization %s does not pay for its usage so no invoice was sent", billingInfo.OrganizationID)

	return nil
}
