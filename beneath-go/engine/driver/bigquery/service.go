package bigquery

import (
	"context"

	"github.com/beneath-core/beneath-go/engine/driver"
)

// MaxKeySize implements beneath.Service
func (b BigQuery) MaxKeySize() int {
	panic("todo")
}

// MaxRecordSize implements beneath.Service
func (b BigQuery) MaxRecordSize() int {
	panic("todo")
}

// MaxRecordsInBatch implements beneath.Service
func (b BigQuery) MaxRecordsInBatch() int {
	panic("todo")
}

// RegisterProject implements beneath.Service
func (b BigQuery) RegisterProject(ctx context.Context, p driver.Project) error {
	panic("todo")
}

// RemoveProject implements beneath.Service
func (b BigQuery) RemoveProject(ctx context.Context, p driver.Project) error {
	panic("todo")
}

// RegisterInstance implements beneath.Service
func (b BigQuery) RegisterInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")
}

// PromoteInstance implements beneath.Service
func (b BigQuery) PromoteInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")
}

// RemoveInstance implements beneath.Service
func (b BigQuery) RemoveInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")
}
