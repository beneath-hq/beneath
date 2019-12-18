package bigtable

import (
	"context"

	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/engine/driver"
)

// ParseLogQuery implements beneath.Log
func (b BigTable) ParseLogQuery(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, where queryparse.Query) ([]byte, error) {
	panic("todo")
}

// ReadLog implements beneath.Log
func (b BigTable) ReadLog(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, cursor []byte, limit int) (driver.RecordsIterator, error) {
	panic("todo")
}

// WriteToLog implements beneath.Log
func (b BigTable) WriteToLog(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, r driver.RecordsIterator) error {
	panic("todo")
}
