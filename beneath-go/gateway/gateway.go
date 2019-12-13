package gateway

import (
	"github.com/beneath-core/beneath-go/metrics"
)

var (
	// Metrics collects stats on records read from/written to Beneath
	Metrics *metrics.Broker
)

func init() {
	Metrics = metrics.NewBroker()
}
