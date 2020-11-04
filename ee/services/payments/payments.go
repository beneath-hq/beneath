package payments

import (
	"context"
	"fmt"

	"gitlab.com/beneath-hq/beneath/bus"

	"github.com/go-chi/chi"

	"gitlab.com/beneath-hq/beneath/ee/models"
	"gitlab.com/beneath-hq/beneath/ee/services/billing"
	"gitlab.com/beneath-hq/beneath/ee/services/payments/driver"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/services/organization"
	"gitlab.com/beneath-hq/beneath/services/permissions"
)

// Service handles payment, i.e., invoicing and charging
type Service struct {
	Bus     *bus.Bus
	Billing *billing.Service
	drivers map[string]driver.Driver
}

// Options configures the payments service
type Options struct {
	Drivers []*struct {
		DriverName string                 `mapstructure:"driver"`
		DriverOpts map[string]interface{} `mapstructure:",remain"`
	} `mapstructure:"drivers"`
}

// New creates a new Payments
func New(opts *Options, bus *bus.Bus, billing *billing.Service, organizations *organization.Service, permissions *permissions.Service) (*Service, error) {
	drivers := make(map[string]driver.Driver)
	for _, driverConf := range opts.Drivers {
		name := driverConf.DriverName
		constructor, ok := driver.Drivers[name]
		if !ok {
			return nil, fmt.Errorf("no driver found with name '%s'", name)
		}

		drv, err := constructor(billing, organizations, permissions, driverConf.DriverOpts)
		if err != nil {
			return nil, err
		}

		drivers[name] = drv
	}

	s := &Service{
		Bus:     bus,
		Billing: billing,
		drivers: drivers,
	}

	s.Bus.AddAsyncListener(s.HandleShouldInvoiceBilledResourcesEvent)

	return s, nil
}

// RegisterHandlers registers
func (s *Service) RegisterHandlers(router *chi.Mux) {
	for name, driver := range s.drivers {
		for subpath, handler := range driver.GetHTTPHandlers() {
			path := fmt.Sprintf("/billing/%s/%s", name, subpath)
			router.Handle(path, httputil.AppHandler(handler))
		}
	}
}

// IssueInvoiceForResources sends an invoice and attempts to charge for the given resources
func (s *Service) IssueInvoiceForResources(ctx context.Context, bi *models.BillingInfo, resources []*models.BilledResource) error {
	if len(resources) == 0 {
		return nil
	}

	driverName := bi.BillingMethod.PaymentsDriver
	drv, ok := s.drivers[driverName]
	if !ok {
		return fmt.Errorf("uninitialized payments driver '%s'", driverName)
	}

	return drv.IssueInvoiceForResources(bi, resources)
}
