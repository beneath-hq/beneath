package bigtable

import (
	"context"

	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/engine/driver"
)

// ParseQuery implements driver.LookupService
func (b BigTable) ParseQuery(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, where queryparse.Query, compacted bool, partitions int) ([][]byte, [][]byte, error) {
	panic("todo")
}

// ReadCursor implements driver.LookupService
func (b BigTable) ReadCursor(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, cursor []byte, limit int) (driver.RecordsIterator, error) {
	panic("todo")
}

// WriteRecords implements driver.LookupService
func (b BigTable) WriteRecords(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, rs []driver.Record) error {
	// build lookup for all rows (using key)
	// execute lookup (get current value and timestamp)

	// for each record to write
	//   if no old value, skip
	//   if current value has a higher timestamp, skip
	//   for each secondary index
	//     if old and new value is different, create delete mutation

	// execute delete mutations

	// write to log

	// write to primary and secondary indexes simultaneously

	panic("todo")
}
