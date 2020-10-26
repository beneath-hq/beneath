package mock

import (
	"context"
	"encoding/hex"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/infrastructure/engine/driver"
)

// WriteToWarehouse implements beneath.WarehouseService
func (m Mock) WriteToWarehouse(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, rs []driver.Record) error {
	return nil
}

// GetWarehouseTableName implements beneath.WarehouseService
func (m Mock) GetWarehouseTableName(p driver.Project, s driver.Stream, i driver.StreamInstance) string {
	instanceID := i.GetStreamInstanceID()
	shortID := hex.EncodeToString(instanceID[0:4])
	return fmt.Sprintf("%s.%s.%s_%s", p.GetOrganizationName(), p.GetProjectName(), s.GetStreamName(), shortID)
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
