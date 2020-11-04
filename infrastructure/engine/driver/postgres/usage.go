package postgres

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/infrastructure/engine/driver"
	pb "gitlab.com/beneath-hq/beneath/infrastructure/engine/proto"
)

// WriteUsage implements engine.LookupService
func (b Postgres) WriteUsage(ctx context.Context, id uuid.UUID, label driver.UsageLabel, ts time.Time, usage pb.QuotaUsage) error {
	panic("not implemented")
}

// ClearUsage implements engine.LookupService
func (b Postgres) ClearUsage(ctx context.Context, id uuid.UUID) error {
	panic("not implemented")
}

// ReadUsageSingle implements engine.LookupService
func (b Postgres) ReadUsageSingle(ctx context.Context, id uuid.UUID, label driver.UsageLabel, ts time.Time) (pb.QuotaUsage, error) {
	panic("not implemented")
}

// ReadUsageRange engine.LookupService
func (b Postgres) ReadUsageRange(ctx context.Context, id uuid.UUID, label driver.UsageLabel, from time.Time, to time.Time, limit int, fn func(ts time.Time, usage pb.QuotaUsage) error) error {
	panic("not implemented")
}
