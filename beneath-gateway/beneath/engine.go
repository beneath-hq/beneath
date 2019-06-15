package beneath

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
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
func (e *Engine) QueueWrite(req *pb.WriteEncodedRecordsRequest) error {
	// validate instanceID
	if uuid.FromBytesOrNil(req.InstanceId) == uuid.Nil {
		return errors.New("invalid instanceId")
	}

	// validate records
	for idx, record := range req.Records {
		// check sequence number
		if len(record.SequenceNumber) > sequenceNumberSize {
			return fmt.Errorf(
				"record at index %d has invalid sequence number <%x>",
				record.SequenceNumber, idx,
			)
		}

		// check encoded key length
		if len(record.Key) == 0 || len(record.Key) > e.Tables.GetMaxKeySize() {
			return fmt.Errorf(
				"record at index %d has invalid key size <%d bytes> (max key size is <%d bytes>)",
				idx, len(record.Key), e.Tables.GetMaxKeySize(),
			)
		}

		// check encoded data length
		if len(record.Data) > e.Tables.GetMaxDataSize() {
			return fmt.Errorf(
				"record at index %d has invalid size <%d bytes> (max key size is <%d bytes>)",
				idx, len(record.Data), e.Tables.GetMaxDataSize(),
			)
		}
	}

	// encode message
	encoded, err := proto.Marshal(req)
	if err != nil {
		log.Panicf("error marshalling WriteEncodedRecordsRequest: %v", err)
	}

	// check encoded message size
	if len(encoded) > e.Streams.GetMaxMessageSize() {
		return fmt.Errorf(
			"write request has invalid size <%d> (max message size is <%d bytes>)",
			len(encoded), e.Streams.GetMaxMessageSize(),
		)
	}

	// push to queue
	err = e.Streams.PushWriteRequest(encoded)
	return err
}
