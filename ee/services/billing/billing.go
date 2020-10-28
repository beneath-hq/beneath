package billing

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/bus"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
	"gitlab.com/beneath-hq/beneath/pkg/refreshingval"
	"gitlab.com/beneath-hq/beneath/services/metrics"
	"gitlab.com/beneath-hq/beneath/services/organization"
)

// Service contains functionality for setting billing info and sending bills
type Service struct {
	Bus           *bus.Bus
	DB            db.DB
	Metrics       *metrics.Broker
	Organizations *organization.Service

	defaultBillingPlan *refreshingval.RefreshingValue
}

// New creates a new Service
func New(bus *bus.Bus, db db.DB, metrics *metrics.Broker, organizations *organization.Service) *Service {
	s := &Service{
		Bus:           bus,
		DB:            db,
		Metrics:       metrics,
		Organizations: organizations,
	}

	s.defaultBillingPlan = refreshingval.New(time.Hour, func(ctx context.Context) interface{} {
		return s.FindDefaultBillingPlan(context.Background())
	})

	return s
}
