package engine

import (
	"errors"
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/beneath-go/engine/driver/bigquery"
	"github.com/beneath-core/beneath-go/engine/driver/bigtable"
	"github.com/beneath-core/beneath-go/engine/driver/pubsub"
	pb "github.com/beneath-core/beneath-go/proto"
)

const (
	sequenceNumberSize = 16 // 128 bit
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
	case "warehouse":
		engine.Warehouse = bigquery.New()
	default:
		log.Fatalf("invalid warehouse platform %s", warehouseDriver)
	}

	// done
	return engine
}

// QueueWrite todo
func (e *Engine) QueueWrite(req *pb.WriteInternalRecordsRequest) error {
	// validate instanceID
	if uuid.FromBytesOrNil(req.InstanceId) == uuid.Nil {
		return errors.New("invalid instanceId")
	}

	// validate records
	for idx, record := range req.Records {
		// check encoded key length
		if len(record.EncodedKey) == 0 || len(record.EncodedKey) > e.Tables.GetMaxKeySize() {
			return fmt.Errorf(
				"record at index %d has invalid key size <%d bytes> (max key size is <%d bytes>)",
				idx, len(record.EncodedKey), e.Tables.GetMaxKeySize(),
			)
		}

		// check encoded data length
		if len(record.EncodedData) > e.Tables.GetMaxDataSize() {
			return fmt.Errorf(
				"record at index %d has invalid size <%d bytes> (max key size is <%d bytes>)",
				idx, len(record.EncodedData), e.Tables.GetMaxDataSize(),
			)
		}
	}

	// push to queue
	err := e.Streams.PushWriteRequest(req)
	return err
}
