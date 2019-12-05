package bigtable

import (
	"context"

	"github.com/beneath-core/beneath-go/engine/driver"
)

// MaxKeySize implements beneath.Service
func (b BigTable) MaxKeySize() int {
	panic("todo")
}

// MaxRecordSize implements beneath.Service
func (b BigTable) MaxRecordSize() int {
	panic("todo")
}

// MaxRecordsInBatch implements beneath.Service
func (b BigTable) MaxRecordsInBatch() int {
	panic("todo")
}

// RegisterProject implements beneath.Service
func (b BigTable) RegisterProject(ctx context.Context, p driver.Project) error {
	panic("todo")
}

// RemoveProject implements beneath.Service
func (b BigTable) RemoveProject(ctx context.Context, p driver.Project) error {
	panic("todo")
}

// RegisterInstance implements beneath.Service
func (b BigTable) RegisterInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")
}

// PromoteInstance implements beneath.Service
func (b BigTable) PromoteInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")
}

// RemoveInstance implements beneath.Service
func (b BigTable) RemoveInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")
}
