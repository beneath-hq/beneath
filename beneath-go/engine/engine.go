package engine

import (
	"fmt"
	"log"

	"github.com/beneath-core/beneath-go/engine/driver/bigquery"
	"github.com/beneath-core/beneath-go/engine/driver/bigtable"
	"github.com/beneath-core/beneath-go/engine/driver/pubsub"
)

// Engine interfaces with the data layer
type Engine struct {
	Streams   StreamsDriver
	Tables    TablesDriver
	Warehouse WarehouseDriver
}

// NewEngine creates a new Engine instance
func NewEngine(streamsDriver string, tablesDriver string, warehouseDriver string) *Engine {
	engine := &Engine{}

	// init Streams
	switch streamsDriver {
	case "pubsub":
		engine.Streams = pubsub.New()
	default:
		log.Fatalf("invalid streams platform %s", streamsDriver)
	}

	// init Tables
	switch tablesDriver {
	case "bigtable":
		engine.Tables = bigtable.New()
	default:
		log.Fatalf("invalid tables platform %s", tablesDriver)
	}

	// init Warehouse
	switch warehouseDriver {
	case "bigquery":
		engine.Warehouse = bigquery.New()
	default:
		log.Fatalf("invalid warehouse platform %s", warehouseDriver)
	}

	// done
	return engine
}

// CheckSize validates that the size of a record (its key and its combined encoded avro)
// fits within the constraints of the underlying infrastructure
func (e *Engine) CheckSize(keyBytesLen int, avroBytesLen int) error {
	// check key size
	if keyBytesLen > e.Tables.GetMaxKeySize() {
		return fmt.Errorf("invalid key size <%d bytes> (max key size is <%d bytes>)", keyBytesLen, e.Tables.GetMaxKeySize())
	}

	// find max data size (= min(tablesMaxSize, warehouseMaxSize))
	maxDataSize := e.Tables.GetMaxDataSize()
	if maxDataSize > e.Warehouse.GetMaxDataSize() {
		maxDataSize = e.Warehouse.GetMaxDataSize()
	}

	// check data size
	if avroBytesLen > maxDataSize {
		return fmt.Errorf("invalid data size <%d bytes> (max data size is <%d bytes>)", avroBytesLen, maxDataSize)
	}

	// passed
	return nil
}

// CheckSequenceNumber validates a sequence number as it must tolerate some truncation
// (specifically multiplication by 1000 without overflow)
func (e *Engine) CheckSequenceNumber(sequenceNumber int64) error {
	if sequenceNumber == 0 {
		return nil
	}
	c := sequenceNumber * 1000
	if (c < 0) == (sequenceNumber < 0) {
		if c/1000 == sequenceNumber {
			return nil
		}
	}
	return fmt.Errorf("abs(sequence_number) exceeds limit of '9223372036854775'")
}
