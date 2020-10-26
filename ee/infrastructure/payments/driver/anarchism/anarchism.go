package anarchism

import (
	"gitlab.com/beneath-hq/beneath/ee/infrastructure/payments/driver"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
)

// Anarchism implements beneath.PaymentsDriver
type Anarchism struct{}

// New initializes a Anarchism object
func New() Anarchism {
	return Anarchism{}
}

// GetHTTPHandlers returns the necessary handlers to implement Anarchism
func (a *Anarchism) GetHTTPHandlers() map[string]httputil.AppHandler {
	return map[string]httputil.AppHandler{
		// "task" : a.handleTask
	}
}

// IssueInvoiceForResources implements Payments interface
func (a *Anarchism) IssueInvoiceForResources(billingInfo driver.BillingInfo, billedResources []driver.BilledResource) error {

	// Amazing... no payment required!
	log.S.Infof("anarchism! organization %s does not pay for its usage so no invoice was sent", billingInfo.GetOrganizationID())

	return nil
}
