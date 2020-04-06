package payments

import (
	"gitlab.com/beneath-hq/beneath/control/payments/driver"
	"gitlab.com/beneath-hq/beneath/control/payments/driver/anarchism"
	"gitlab.com/beneath-hq/beneath/control/payments/driver/stripecard"
	"gitlab.com/beneath-hq/beneath/control/payments/driver/stripewire"
)

// InitDrivers initializes all of the payments drivers
func InitDrivers(drivers []string) map[string]driver.PaymentsDriver {
	payments := make(map[string]driver.PaymentsDriver)
	for _, driver := range drivers {
		switch driver {
		case "stripecard":
			sc := stripecard.New()
			payments[driver] = &sc
		case "stripewire":
			sw := stripewire.New()
			payments[driver] = &sw
		case "anarchism":
			a := anarchism.New()
			payments[driver] = &a
		default:
			panic("unrecognized payments driver")
		}
	}
	return payments
}
