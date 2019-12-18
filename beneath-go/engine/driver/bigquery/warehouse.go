package bigquery

import (
	"context"

	"github.com/beneath-core/beneath-go/engine/driver"
)

// WriteToWarehouse implements beneath.WarehouseService
func (b BigQuery) WriteToWarehouse(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, r driver.RecordsIterator) error {
	panic("todo")
}
