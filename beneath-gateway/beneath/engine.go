package beneath

import (
	"errors"
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"

	pb "github.com/beneath-core/beneath-gateway/beneath/beneath_proto"
	"github.com/beneath-core/beneath-gateway/beneath/driver/bigtable"
	"github.com/beneath-core/beneath-gateway/beneath/driver/pubsub"
)

const (
	sequenceNumberSize = 16 // 128 bit
)

// Engine interfaces with the data layer
type Engine struct {
	Streams StreamsDriver
	Tables  TablesDriver
}

// NewEngine creates a new Engine instance
func NewEngine() *Engine {
	engine := &Engine{}

	// init Streams
	switch Config.StreamsPlatform {
	case "pubsub":
		engine.Streams = pubsub.New()
	default:
		log.Fatalf("invalid streams platform %s", Config.StreamsPlatform)
	}

	// init Tables
	switch Config.TablesPlatform {
	case "bigtable":
		engine.Tables = bigtable.New()
	default:
		log.Fatalf("invalid tables platform %s", Config.TablesPlatform)
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
