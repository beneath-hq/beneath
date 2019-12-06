package engine

import (
	"context"

	pb "github.com/beneath-core/beneath-go/proto"
)

// CommitUsage writes a batch of usage metrics
func (e *Engine) CommitUsage(ctx context.Context, key []byte, usage pb.QuotaUsage) error {
	panic("todo")
}

// ReadSingleUsage reads usage metrics for one key
func (e *Engine) ReadSingleUsage(ctx context.Context, key []byte) (pb.QuotaUsage, error) {
	panic("todo")
}

// ReadUsage reads usage metrics for multiple periods and calls fn one by one
func (e *Engine) ReadUsage(ctx context.Context, fromKey []byte, toKey []byte, fn func(key []byte, usage pb.QuotaUsage) error) error {
	panic("todo")
}
