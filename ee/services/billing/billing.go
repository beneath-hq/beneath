package billing

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/bus"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
	"gitlab.com/beneath-hq/beneath/pkg/refreshingval"
	"gitlab.com/beneath-hq/beneath/services/organization"
	"gitlab.com/beneath-hq/beneath/services/usage"
)

// Service contains functionality for setting billing info and sending bills
type Service struct {
	Bus           *bus.Bus
	DB            db.DB
	Usage         *usage.Service
	Organizations *organization.Service

	defaultBillingPlan *refreshingval.RefreshingValue
}

// New creates a new Service
func New(bus *bus.Bus, db db.DB, usage *usage.Service, organizations *organization.Service) *Service {
	s := &Service{
		Bus:           bus,
		DB:            db,
		Usage:         usage,
		Organizations: organizations,
	}

	s.defaultBillingPlan = refreshingval.New(time.Hour, func(ctx context.Context) interface{} {
		return s.FindDefaultBillingPlan(context.Background())
	})

	s.Bus.AddSyncListener(s.HandleOrganizationCreatedEvent)
	s.Bus.AddSyncListener(s.HandleOrganizationTransferredUserEvent)

	return s
}
