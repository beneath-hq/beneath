package mock

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	uuid "github.com/satori/go.uuid"
)

// WriteToWarehouse implements beneath.WarehouseService
func (m Mock) WriteToWarehouse(ctx context.Context, p driver.Project, s driver.Table, i driver.TableInstance, rs []driver.Record) error {
	return nil
}

// GetWarehouseTableName implements beneath.WarehouseService
func (m Mock) GetWarehouseTableName(p driver.Project, s driver.Table, i driver.TableInstance) string {
	instanceID := i.GetTableInstanceID()
	shortID := hex.EncodeToString(instanceID[0:4])
	return fmt.Sprintf("%s.%s.%s_%s", p.GetOrganizationName(), p.GetProjectName(), s.GetTableName(), shortID)
}

// AnalyzeWarehouseQuery implements beneath.WarehouseService
func (m Mock) AnalyzeWarehouseQuery(ctx context.Context, query string) (driver.WarehouseJob, error) {
	panic("todo")
}

// RunWarehouseQuery implements beneath.WarehouseService
func (m Mock) RunWarehouseQuery(ctx context.Context, jobID uuid.UUID, query string, partitions int, timeoutMs int, maxBytesScanned int) (driver.WarehouseJob, error) {
	panic("todo")
}

// PollWarehouseJob implements beneath.WarehouseService
func (m Mock) PollWarehouseJob(ctx context.Context, jobID uuid.UUID) (driver.WarehouseJob, error) {
	panic("todo")
}

// ReadWarehouseCursor implements beneath.WarehouseService
func (m Mock) ReadWarehouseCursor(ctx context.Context, cursor []byte, limit int) (driver.RecordsIterator, error) {
	return mockRecordsIterator{}, nil
}
