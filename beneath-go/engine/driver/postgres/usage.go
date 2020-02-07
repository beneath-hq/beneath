package postgres

import (
	"context"

	pb "github.com/beneath-core/beneath-go/proto"
)

// CommitUsage implements engine.LookupService
func (b Postgres) CommitUsage(ctx context.Context, key []byte, usage pb.QuotaUsage) error {
	panic("not implemented")
}

// ReadSingleUsage implements engine.LookupService
func (b Postgres) ReadSingleUsage(ctx context.Context, key []byte) (pb.QuotaUsage, error) {
	panic("not implemented")
}

// ReadUsage implements engine.LookupService
func (b Postgres) ReadUsage(ctx context.Context, fromKey []byte, toKey []byte, fn func(key []byte, usage pb.QuotaUsage) error) error {
	panic("not implemented")
}
