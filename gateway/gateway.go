package gateway

import (
	"github.com/beneath-core/engine"
	"github.com/beneath-core/gateway/subscriptions"
	"github.com/beneath-core/metrics"
)

var (
	// Metrics collects stats on records read from/written to Beneath
	Metrics *metrics.Broker

	// Subscriptions handles real-time data subscriptions
	Subscriptions *subscriptions.Broker
)

// InitMetrics initializes the Metrics global
func InitMetrics() {
	Metrics = metrics.NewBroker()
}

// InitSubscriptions initializes the Subscriptionsglobal
func InitSubscriptions(eng *engine.Engine) {
	Subscriptions = subscriptions.NewBroker(eng)
}
