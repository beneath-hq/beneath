package dependencies

import (
	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	"gitlab.com/beneath-hq/beneath/ee/services/billing"
)

// See non-EE file for details

// AllServices is a convenience wrapper that initializes all *enterprise* services
type AllServices struct {
	Billing *billing.Service
}

// NewAllServices creates a new AllServices
func NewAllServices(billing *billing.Service) *AllServices {
	return &AllServices{
		Billing: billing,
	}
}

func init() {
	cli.AddDependency(NewAllServices)
	cli.AddDependency(billing.New)
}
