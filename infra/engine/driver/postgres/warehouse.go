package postgres

import (
	"context"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	uuid "github.com/satori/go.uuid"
)

// WriteToWarehouse implements beneath.WarehouseService
func (b Postgres) WriteToWarehouse(ctx context.Context, p driver.Project, s driver.Table, i driver.TableInstance, rs []driver.Record) error {
	panic("todo")
}

// GetWarehouseTableName implements beneath.WarehouseService
func (b Postgres) GetWarehouseTableName(p driver.Project, s driver.Table, i driver.TableInstance) string {
	panic("todo")
}

// AnalyzeWarehouseQuery implements beneath.WarehouseService
func (b Postgres) AnalyzeWarehouseQuery(ctx context.Context, query string) (driver.WarehouseJob, error) {
	panic("todo")
}

// RunWarehouseQuery implements beneath.WarehouseService
func (b Postgres) RunWarehouseQuery(ctx context.Context, jobID uuid.UUID, query string, partitions int, timeoutMs int, maxBytesScanned int) (driver.WarehouseJob, error) {
	panic("todo")
}

// PollWarehouseJob implements beneath.WarehouseService
func (b Postgres) PollWarehouseJob(ctx context.Context, jobID uuid.UUID) (driver.WarehouseJob, error) {
	panic("todo")
}

// ReadWarehouseCursor implements beneath.WarehouseService
func (b Postgres) ReadWarehouseCursor(ctx context.Context, cursor []byte, limit int) (driver.RecordsIterator, error) {
	panic("todo")
}
