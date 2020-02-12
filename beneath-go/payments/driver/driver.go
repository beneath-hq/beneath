package driver

import (
	"time"

	"github.com/beneath-core/beneath-go/core/httputil"
	uuid "github.com/satori/go.uuid"
)

// PaymentsDriver defines the functions necessary to pay a bill
type PaymentsDriver interface {
	GetHTTPHandlers() map[string]httputil.AppHandler
	IssueInvoiceForResources(billingInfo BillingInfo, billedResources []BilledResource) error
}

// BillingInfo encapsulates metadata about a Beneath billing info
type BillingInfo interface {
	GetOrganizationID() uuid.UUID
	GetBillingPlanCurrency() string
	GetDriverPayload() map[string]interface{}
	GetPaymentsDriver() string
}

// BilledResource encapsulates metadata about a Beneath billed resource
type BilledResource interface {
	GetTotalPriceCents() int32
	GetStartTime() time.Time
	GetEndTime() time.Time
	GetBillingTime() time.Time
	GetProduct() string
	GetEntityName() string
}
