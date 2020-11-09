package anarchism

import (
	"go.uber.org/zap"

	"gitlab.com/beneath-hq/beneath/ee/models"
	"gitlab.com/beneath-hq/beneath/ee/services/billing"
	"gitlab.com/beneath-hq/beneath/ee/services/payments/driver"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/services/organization"
	"gitlab.com/beneath-hq/beneath/services/permissions"
)

// Anarchism implements driver.PaymentsDriver
type Anarchism struct {
	Logger *zap.SugaredLogger
}

func init() {
	driver.AddDriver("anarchism", newAnarchism)
}

func newAnarchism(logger *zap.Logger, billing *billing.Service, organizations *organization.Service, permissions *permissions.Service, opts map[string]interface{}) (driver.Driver, error) {
	return &Anarchism{
		Logger: logger.Named("anarchism").Sugar(),
	}, nil
}

// GetHTTPHandlers returns the necessary handlers to implement Anarchism
func (a *Anarchism) GetHTTPHandlers() map[string]httputil.AppHandler {
	return map[string]httputil.AppHandler{}
}

// IssueInvoiceForResources implements Payments interface
func (a *Anarchism) IssueInvoiceForResources(bi *models.BillingInfo, resources []*models.BilledResource) error {
	// Amazing... no payment required!
	a.Logger.Infof("organization %s does not pay for its usage so no invoice was sent", bi.OrganizationID)
	return nil
}
