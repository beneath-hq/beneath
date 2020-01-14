package postgres

import (
	"context"

	"github.com/beneath-core/beneath-go/engine/driver"
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

// RegisterProject implements beneath.Service
func (b Postgres) RegisterProject(ctx context.Context, p driver.Project) error {
	panic("todo")
}

// RemoveProject implements beneath.Service
func (b Postgres) RemoveProject(ctx context.Context, p driver.Project) error {
	panic("todo")
}

// RegisterInstance implements beneath.Service
func (b Postgres) RegisterInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")
}

// PromoteInstance implements beneath.Service
func (b Postgres) PromoteInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")
}

// RemoveInstance implements beneath.Service
func (b Postgres) RemoveInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")
}
