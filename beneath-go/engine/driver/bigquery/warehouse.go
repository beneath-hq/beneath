package bigquery

import (
	"context"

	"github.com/beneath-core/beneath-go/engine/driver"
)

// WriteRecords implements beneath.WarehouseService
func (b BigQuery) WriteRecords(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, r driver.RecordsReader) error {
	panic("todo")
}
