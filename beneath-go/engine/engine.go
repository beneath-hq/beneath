package engine

import (
	"fmt"

	"github.com/beneath-core/beneath-go/core/codec"
	"github.com/beneath-core/beneath-go/core/mathutil"
	"github.com/beneath-core/beneath-go/engine/driver"
	"github.com/beneath-core/beneath-go/engine/driver/bigquery"
	"github.com/beneath-core/beneath-go/engine/driver/bigtable"
	"github.com/beneath-core/beneath-go/engine/driver/postgres"
	"github.com/beneath-core/beneath-go/engine/driver/pubsub"
)

// Engine interfaces with the data layer
type Engine struct {
	MQ        driver.MessageQueue
	Lookup    driver.LookupService
	Warehouse driver.WarehouseService

	maxBatchLength int
	maxRecordSize  int
	maxKeySize     int
}

// NewEngine creates a new Engine instance
func NewEngine(mqDriver, lookupDriver, warehouseDriver string) *Engine {
	e := &Engine{}
	e.MQ = makeMQ(mqDriver)
	e.Lookup = makeLookup(lookupDriver)
	e.Warehouse = makeWarehouse(warehouseDriver)

	e.maxBatchLength = mathutil.MinInts(e.Lookup.MaxRecordsInBatch(), e.Warehouse.MaxRecordsInBatch())
	e.maxRecordSize = mathutil.MinInts(e.Lookup.MaxRecordSize(), e.Warehouse.MaxRecordSize())
	e.maxKeySize = mathutil.MinInts(e.Lookup.MaxKeySize(), e.Warehouse.MaxKeySize())

	return e
}

func makeMQ(driver string) driver.MessageQueue {
	switch driver {
	case "pubsub":
		return pubsub.GetMessageQueue()
	default:
		panic(fmt.Errorf("invalid mq driver '%s'", driver))
	}
}

func makeLookup(driver string) driver.LookupService {
	switch driver {
	case "bigtable":
		return bigtable.GetLookupService()
	case "postgres":
		return postgres.GetLookupService()
	default:
		panic(fmt.Errorf("invalid lookup driver '%s'", driver))
	}
}

func makeWarehouse(driver string) driver.WarehouseService {
	switch driver {
	case "bigquery":
		return bigquery.GetWarehouseService()
	case "postgres":
		return postgres.GetWarehouseService()
	default:
		panic(fmt.Errorf("invalid warehouse driver '%s'", driver))
	}
}

// Healthy returns true if connected to all services
func (e *Engine) Healthy() bool {
	return true
}

// CheckRecordSize validates that the record fits within the constraints of the underlying infrastructure
func (e *Engine) CheckRecordSize(s driver.Stream, structured map[string]interface{}, avroBytesLen int) error {
	if avroBytesLen > e.maxRecordSize {
		return fmt.Errorf("encoded record size exceeds maximum of %d bytes", e.maxRecordSize)
	}

	codec := s.GetCodec()

	err := e.checkKeySize(codec, codec.PrimaryIndex, structured)
	if err != nil {
		return err
	}

	for _, index := range codec.SecondaryIndexes {
		err := e.checkKeySize(codec, index, structured)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) checkKeySize(codec *codec.Codec, index codec.Index, structured map[string]interface{}) error {
	key, err := codec.MarshalKey(index, structured)
	if err != nil {
		return err
	}

	if len(key) > e.maxKeySize {
		return fmt.Errorf("encoded key size for index on fields %v exceeds maximum length of %d bytes", index.GetFields(), e.maxKeySize)
	}

	return nil
}

// CheckBatchLength validates that the number of records in a batch fits within the constraints of the underlying infrastructure
func (e *Engine) CheckBatchLength(length int) error {
	if length > e.maxBatchLength {
		return fmt.Errorf("batch length exceeds maximum of %d", e.maxBatchLength)
	}
	return nil
}
