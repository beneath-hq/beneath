package bigtable

import (
	"context"

	"github.com/beneath-core/beneath-go/engine/driver"
)

// WriteRecords implements beneath.LookupService
func (b BigTable) WriteRecords(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, r driver.RecordsIterator) error {
	panic("todo")
}
