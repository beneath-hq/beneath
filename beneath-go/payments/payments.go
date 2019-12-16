package payments

import (
	"github.com/beneath-core/beneath-go/payments/driver/anarchism"
	"github.com/beneath-core/beneath-go/payments/driver/stripecard"
	"github.com/beneath-core/beneath-go/payments/driver/stripewire"
)

var (
	global map[string]PaymentsDriver
)

// InitDrivers initializes all of the payments drivers
func InitDrivers(drivers []string) {
	global = make(map[string]PaymentsDriver)
	for _, driver := range drivers {
		switch driver {
		case "stripecard":
			sc := stripecard.New()
			global[driver] = &sc
		case "stripewire":
			sw := stripewire.New()
			global[driver] = &sw
		case "anarchism":
			a := anarchism.New()
			global[driver] = &a
		default:
			panic("unrecognized payments driver")
		}
	}
}

// GetDriver retrieves a payment driver
func GetDriver(name string) (PaymentsDriver, error) {
	d := global[string(name)]
	if d == nil {
		panic("attempted to get a payments driver that was not enabled")
	}
	return d, nil
}
