package postgres

import (
	"context"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	"github.com/beneath-hq/beneath/pkg/queryparse"
)

// ParseQuery implements driver.LookupService
func (b Postgres) ParseQuery(ctx context.Context, p driver.Project, s driver.Table, i driver.TableInstance, where queryparse.Query, compacted bool, partitions int) ([][]byte, [][]byte, error) {
	panic("todo")
}

// Peek implements driver.LookupService
func (b Postgres) Peek(ctx context.Context, p driver.Project, s driver.Table, i driver.TableInstance) ([]byte, []byte, error) {
	panic("todo")
}

// ReadCursor implements driver.LookupService
func (b Postgres) ReadCursor(ctx context.Context, p driver.Project, s driver.Table, i driver.TableInstance, cursor []byte, limit int) (driver.RecordsIterator, error) {
	panic("todo")
}

// WriteRecords implements driver.LookupService
func (b Postgres) WriteRecords(ctx context.Context, p driver.Project, s driver.Table, i driver.TableInstance, rs []driver.Record) error {
	// postgres: s.compacted
	// - indexes for secondary, primary key for key, index for offset
	// - upsert on primary key with new offset (autoincrement) conditional on timestamp greater
	// - secondary indexes fix themselves

	// postgres: !s.compacted
	// - primary key on offset
	// - current = true|false
	// - index on (key/secondary, current=true)
	// - set current to false on old when insert new (if timesamp greater)

	panic("todo")
}
