package pipeline

import (
	"fmt"
	"log"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/beneath-go/control/model"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/db"
	pb "github.com/beneath-core/beneath-go/proto"
)

type configSpecification struct {
	StreamsDriver   string `envconfig:"ENGINE_STREAMS_DRIVER" required:"true"`
	TablesDriver    string `envconfig:"ENGINE_TABLES_DRIVER" required:"true"`
	WarehouseDriver string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL        string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresURL     string `envconfig:"CONTROL_POSTGRES_URL" required:"true"`
}

var (
	// Config for gateway
	Config configSpecification
)

func init() {
	core.LoadConfig("beneath", &Config)

	db.InitPostgres(Config.PostgresURL)
	db.InitRedis(Config.RedisURL)
	db.InitEngine(Config.StreamsDriver, Config.TablesDriver, Config.WarehouseDriver)
}

// Run runs the pipeline: subscribes from pubsub and sends data to BigTable and BigQuery
func Run() error {
	// log that we're running
	log.Printf("Pipeline processing write requests\n")

	// begin processing write requests -- will run infinitely
	err := db.Engine.Streams.ReadWriteRequests(processWriteRequest)

	// processing incoming write requests crashed for some reason
	if err != nil {
		return err
	}

	return nil
}

// processWriteRequest is called (approximately once) for each new write request
// TODO: add metrics tracking -- group by instanceID and hour: 1) writes and 2) bytes
func processWriteRequest(req *pb.WriteRecordsRequest) error {
	// metrics to track
	startTime := time.Now()
	var bytesWritten int64
	bytesWritten = 0

	// lookup stream for write request
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		return fmt.Errorf("cached stream is null for instanceid %s", instanceID.String())
	}

	// keep track of keys to publish
	numRecords := len(req.Records)
	keys := make([][]byte, numRecords)

	// save each record to bigtable and bigquery
	avroData := make([][]byte, numRecords)
	data := make([]map[string]interface{}, numRecords)
	sequenceNumbers := make([]int64, numRecords)

	// loop through each record in the Write Request
	for i, record := range req.Records {
		avroData[i] = record.AvroData
		// decode the avro data
		dataT, err := stream.AvroCodec.Unmarshal(record.AvroData, false)
		if err != nil {
			return fmt.Errorf("unable to decode avro data")
		}

		// assert that the decoded data is a map
		var ok bool
		data[i], ok = dataT.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected decoded data to be a map, got %T", dataT)
		}

		// get the encoded key
		keys[i], err = stream.KeyCodec.Marshal(data[i])
		if err != nil {
			return fmt.Errorf("unable to encode key")
		}

		// save sequence number
		sequenceNumbers[i] = record.SequenceNumber

		// increment metrics
		bytesWritten += int64(len(record.AvroData))
	}

	// writing encoded data to Table
	err := db.Engine.Tables.WriteRecords(instanceID, keys, avroData, sequenceNumbers)
	if err != nil {
		return err
	}

	// writing decoded data to Warehouse
	err = db.Engine.Warehouse.WriteRecords(stream.ProjectName, stream.StreamName, instanceID, keys, data, sequenceNumbers)
	if err != nil {
		return err
	}

	// publish metrics packet; the keys in the packet will be used to stream data via the gateway's websocket
	err = db.Engine.Streams.QueueWriteReport(&pb.WriteRecordsReport{
		InstanceId:   instanceID.Bytes(),
		Keys:         keys,
		BytesWritten: bytesWritten,
	})
	if err != nil {
		return err
	}

	// finalise metrics
	elapsed := time.Since(startTime)

	// log metrics
	log.Printf("%s/%s (%s): Wrote %d record(s) (%dB) in %s", stream.ProjectName, stream.StreamName, instanceID.String(), len(req.Records), bytesWritten, elapsed)

	// done
	return nil
}
