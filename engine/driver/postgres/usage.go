package postgres

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"

	pb "gitlab.com/beneath-hq/beneath/engine/proto"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// CommitUsage implements engine.LookupService
func (b Postgres) CommitUsage(ctx context.Context, id uuid.UUID, period timeutil.Period, ts time.Time, usage pb.QuotaUsage) error {
	panic("not implemented")
}

// ClearUsage implements engine.LookupService
func (b Postgres) ClearUsage(ctx context.Context, id uuid.UUID) error {
	panic("not implemented")
}

// ReadSingleUsage implements engine.LookupService
func (b Postgres) ReadSingleUsage(ctx context.Context, id uuid.UUID, period timeutil.Period, ts time.Time) (pb.QuotaUsage, error) {
	panic("not implemented")
}

// ReadUsage implements engine.LookupService
func (b Postgres) ReadUsage(ctx context.Context, id uuid.UUID, period timeutil.Period, from time.Time, until time.Time, fn func(ts time.Time, usage pb.QuotaUsage) error) error {
	panic("not implemented")
}
