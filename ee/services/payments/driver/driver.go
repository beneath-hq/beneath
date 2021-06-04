package driver

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/beneath-hq/beneath/ee/models"
	"github.com/beneath-hq/beneath/ee/services/billing"
	"github.com/beneath-hq/beneath/pkg/httputil"
	"github.com/beneath-hq/beneath/services/organization"
	"github.com/beneath-hq/beneath/services/permissions"
)

// Name represents the available drivers
type Name string

// Driver name constants
const (
	StripeCard = "stripecard"
	StripeWire = "stripewire"
	Anarchism  = "anarchism"
)

// Driver handles charging money
type Driver interface {
	GetHTTPHandlers() map[string]httputil.AppHandler
	IssueInvoiceForResources(bi *models.BillingInfo, resources []*models.BilledResource) error
}

// Constructor is a function that creates a payments driver from a config object
type Constructor func(logger *zap.Logger, billing *billing.Service, organizations *organization.Service, permissions *permissions.Service, opts map[string]interface{}) (Driver, error)

// Drivers is a registry of driver constructors
var Drivers = make(map[string]Constructor)

// AddDriver registers a new driver (by passing the driver's constructor)
func AddDriver(name string, constructor Constructor) {
	if Drivers[name] != nil {
		panic(fmt.Errorf("Payments driver already registered with name '%s'", name))
	}
	Drivers[name] = constructor
}
