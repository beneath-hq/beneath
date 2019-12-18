package bigtable

import (
	"context"

	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/engine/driver"
)

// ParseLookupQuery implements driver.LookupService
func (b BigTable) ParseLookupQuery(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, where queryparse.Query) ([]byte, error) {
	panic("todo")
}

// ReadLookup implements driver.LookupService
func (b BigTable) ReadLookup(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, cursor []byte, limit int) (driver.RecordsIterator, error) {
	panic("todo")
}

// WriteToLookup implements driver.LookupService
func (b BigTable) WriteToLookup(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, r driver.RecordsIterator) error {
	panic("todo")
}
