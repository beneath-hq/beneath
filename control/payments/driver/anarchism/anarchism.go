package anarchism

import (
	"gitlab.com/beneath-hq/beneath/control/payments/driver"
	"gitlab.com/beneath-hq/beneath/pkg/envutil"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
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
