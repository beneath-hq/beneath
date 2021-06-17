package postgres

import (
	"context"

	"github.com/beneath-hq/beneath/infra/engine/driver"
)

// MaxKeySize implements beneath.Service
func (b Postgres) MaxKeySize() int {
	panic("todo")
}

// MaxRecordSize implements beneath.Service
func (b Postgres) MaxRecordSize() int {
	panic("todo")
}

// MaxRecordsInBatch implements beneath.Service
func (b Postgres) MaxRecordsInBatch() int {
	panic("todo")
}

// RegisterInstance implements beneath.Service
func (b Postgres) RegisterInstance(ctx context.Context, s driver.Table, i driver.TableInstance) error {
	panic("todo")
}

// RemoveInstance implements beneath.Service
func (b Postgres) RemoveInstance(ctx context.Context, s driver.Table, i driver.TableInstance) error {
	panic("todo")
}

// Reset implements beneath.Service
func (b Postgres) Reset(ctx context.Context) error {
	panic("todo")
}
