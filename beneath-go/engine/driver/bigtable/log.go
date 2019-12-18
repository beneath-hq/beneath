package bigtable

import (
	"context"

	"github.com/beneath-core/beneath-go/engine/driver"
)

// ReadRecords implements beneath.Log
func (b BigTable) ReadRecords(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, offset int, limit int) (driver.RecordsIterator, error) {
	panic("todo")
}

// AppendRecords implements beneath.Log
func (b BigTable) AppendRecords(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, r driver.RecordsIterator) (offset int, err error) {
	panic("todo")
}
