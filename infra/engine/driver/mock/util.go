package mock

import (
	"gitlab.com/beneath-hq/beneath/infra/engine/driver"
)

type mockRecordsIterator struct{}

// Next implements driver.RecordsIterator
func (m mockRecordsIterator) Next() bool {
	return false
}

// Record implements driver.RecordsIterator
func (m mockRecordsIterator) Record() driver.Record {
	return nil
}

// NextCursor implements driver.RecordsIterator
func (m mockRecordsIterator) NextCursor() []byte {
	return nil
}
