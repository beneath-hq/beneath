package payments

import (
	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core/httputil"
)

// PaymentsDriver defines the functions necessary to pay a bill
type PaymentsDriver interface {
	// GetName() string
	GetHTTPHandlers() map[string]httputil.AppHandler
	IssueInvoiceForResources(billingInfo *entity.BillingInfo, billedResources []*entity.BilledResource) error
}
