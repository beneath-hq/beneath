package bigquery

import (
	"context"

	"github.com/beneath-core/beneath-go/engine/driver"
)

// WriteToWarehouse implements beneath.WarehouseService
func (b BigQuery) WriteToWarehouse(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, rs []driver.Record) error {
	panic("todo")
}
