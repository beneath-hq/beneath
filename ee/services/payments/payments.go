package payments

import (
	"context"
	"fmt"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/beneath-hq/beneath/bus"
	"github.com/beneath-hq/beneath/ee/models"
	"github.com/beneath-hq/beneath/ee/services/billing"
	"github.com/beneath-hq/beneath/ee/services/payments/driver"
	"github.com/beneath-hq/beneath/pkg/httputil"
	"github.com/beneath-hq/beneath/services/organization"
	"github.com/beneath-hq/beneath/services/permissions"
)

// Service handles payment, i.e., invoicing and charging
type Service struct {
	Logger  *zap.SugaredLogger
	Bus     *bus.Bus
	Billing *billing.Service
	drivers map[string]driver.Driver
}

// Options configures the payments service
type Options struct {
	Drivers []*DriverOption `mapstructure:"drivers"`
}

// DriverOption specifies a payments driver
type DriverOption struct {
	DriverName string                 `mapstructure:"driver"`
	DriverOpts map[string]interface{} `mapstructure:",remain"`
}

// New creates a new Payments
func New(opts *Options, logger *zap.Logger, bus *bus.Bus, billing *billing.Service, organizations *organization.Service, permissions *permissions.Service) (*Service, error) {
	l := logger.Named("payments")
	drivers := make(map[string]driver.Driver)
	for _, driverConf := range opts.Drivers {
		name := driverConf.DriverName
		constructor, ok := driver.Drivers[name]
		if !ok {
			return nil, fmt.Errorf("no driver found with name '%s'", name)
		}

		drv, err := constructor(l, billing, organizations, permissions, driverConf.DriverOpts)
		if err != nil {
			return nil, err
		}

		drivers[name] = drv
	}

	s := &Service{
		Logger:  l.Sugar(),
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
