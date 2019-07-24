package pipeline

import (
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/beneath-go/control/db"
	"github.com/beneath-core/beneath-go/control/model"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/engine"
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

	// Engine is the data plane
	Engine *engine.Engine
)

func init() {
	core.LoadConfig("beneath", &Config)

	Engine = engine.NewEngine(Config.StreamsDriver, Config.TablesDriver, Config.WarehouseDriver)

	db.InitPostgres(Config.PostgresURL)
	db.InitRedis(Config.RedisURL)
}

// Run runs the pipeline: subscribes from pubsub and sends data to BigTable and BigQuery
func Run() error {
	// log that we're running
	log.Printf("Pipeline processing write requests\n")

	// begin processing write requests -- will run infinitely
	err := Engine.Streams.ReadWriteRequests(processWriteRequest)

	// processing incoming write requests crashed for some reason
	if err != nil {
		return err
	}

	return nil
}

// processWriteRequest is called (approximately once) for each new write request
// TODO: add metrics tracking -- group by instanceID and hour: 1) writes and 2) bytes
func processWriteRequest(req *pb.WriteRecordsRequest) error {
	// lookup stream for write request
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		return fmt.Errorf("cached stream is null for instanceid %s", instanceID.String())
	}

	// save each record to bigtable and bigquery
	// TODO: Refactor so that we write batch records when >1 record in a write request
	for _, record := range req.Records {
		// decode the avro data
		dataT, err := stream.AvroCodec.Unmarshal(record.AvroData)
		if err != nil {
			return fmt.Errorf("unable to decode avro data")
		}

		// assert that the decoded data is a map
		data, ok := dataT.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected decoded data to be a map, got %T", dataT)
		}

		// get the encoded key
		key, err := stream.KeyCodec.Marshal(data)
		if err != nil {
			return fmt.Errorf("unable to encode key")
		}

		// writing encoded data to Table
		err = Engine.Tables.WriteRecord(instanceID, key, record.AvroData, record.SequenceNumber)
		if err != nil {
			return err
		}

		// writing decoded data to Warehouse
		err = Engine.Warehouse.WriteRecord(stream.ProjectName, stream.StreamName, instanceID, key, data, record.SequenceNumber)
		if err != nil {
			return err
		}
	}

	// done
	return nil
}
