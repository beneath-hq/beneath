package driver

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
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
	GetCountry() string
	GetRegion() string
	IsCompany() bool
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
