package billing

import (
	"context"
	"time"

	"github.com/beneath-hq/beneath/bus"
	"github.com/beneath-hq/beneath/infra/db"
	"github.com/beneath-hq/beneath/pkg/refreshingval"
	"github.com/beneath-hq/beneath/services/organization"
	"github.com/beneath-hq/beneath/services/usage"
	"go.uber.org/zap"
)

// Service contains functionality for setting billing info and sending bills
type Service struct {
	Logger        *zap.SugaredLogger
	Bus           *bus.Bus
	DB            db.DB
	Usage         *usage.Service
	Organizations *organization.Service

	defaultBillingPlan *refreshingval.RefreshingValue
}

// New creates a new Service
func New(logger *zap.Logger, bus *bus.Bus, db db.DB, usage *usage.Service, organizations *organization.Service) *Service {
	s := &Service{
		Logger:        logger.Named("billing").Sugar(),
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
