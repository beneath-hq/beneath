package dependencies

import (
	"github.com/spf13/viper"

	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	"gitlab.com/beneath-hq/beneath/ee/services/billing"
	"gitlab.com/beneath-hq/beneath/ee/services/payments"
	"gitlab.com/beneath-hq/beneath/ee/services/payments/driver"

	// registers all payments drivers
	_ "gitlab.com/beneath-hq/beneath/ee/services/payments/driver/anarchism"
	_ "gitlab.com/beneath-hq/beneath/ee/services/payments/driver/stripe"
)

// See non-EE file for details

// AllServices is a convenience wrapper that initializes all *enterprise* services
type AllServices struct {
	Billing  *billing.Service
	Payments *payments.Service
}

// NewAllServices creates a new AllServices
func NewAllServices(billing *billing.Service, payments *payments.Service) *AllServices {
	return &AllServices{
		Billing:  billing,
		Payments: payments,
	}
}

func init() {
	cli.AddDependency(NewAllServices)
	cli.AddDependency(billing.New)
	cli.AddDependency(payments.New)

	// Payments service takes options
	cli.AddDependency(func(v *viper.Viper) (*payments.Options, error) {
		var opts payments.Options
		err := v.UnmarshalKey("payments", &opts)
		if err != nil {
			return nil, err
		}
		return &opts, nil
	})
	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "payments.drivers",
		Default:     []*payments.DriverOption{&payments.DriverOption{DriverName: driver.Anarchism}},
		Description: "drivers to enable for processing payments",
	})
}
