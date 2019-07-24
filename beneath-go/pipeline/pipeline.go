package pipeline

import (
	"fmt"

	"github.com/beneath-core/beneath-go/control/db"
	"github.com/beneath-core/beneath-go/control/model"

	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/engine"
	pb "github.com/beneath-core/beneath-go/proto"
	uuid "github.com/satori/go.uuid"
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

const (
	project = "beneathcrypto"
)

// Run runs the pipeline: subscribes from pubsub and sends data to BigTable and BigQuery
func Run() error {

	return Engine.Streams.ReadWriteRequests(func(req *pb.WriteRecordsRequest) error {
		instanceID := uuid.FromBytesOrNil(req.InstanceId)
		stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
		if stream == nil {
			return fmt.Errorf("cached stream is null for instanceid %s", instanceID.String())
		}

		// save to bigtable, save to bigquery
		for _, record := range req.Records { // in the future, refactor so that we write batch records when >1 record in a packet
			// decode the avro data
			decodedData, err := stream.AvroCodec.Unmarshal(record.AvroData)
			if err != nil {
				return fmt.Errorf("unable to decode avro data")
			}

			// assert that the decoded data is a map
			decodedMap, ok := decodedData.(map[string]interface{})
			if !ok {
				return fmt.Errorf("expected decoded data to be a map, got %T", decodedData)
			}

			// get the encoded key
			encodedKey, err := stream.KeyCodec.Marshal(decodedMap)
			if err != nil {
				return fmt.Errorf("unable to get encoded key")
			}

			// writing encoded data to Table
			err = Engine.Tables.WriteRecordToTable(instanceID, encodedKey, record.AvroData, record.SequenceNumber)

			// writing decoded data to Warehouse
			err = Engine.Warehouse.WriteRecordToWarehouse(stream.ProjectName, stream.StreamName, instanceID, encodedKey, decodedMap, record.SequenceNumber)

			// calculate metrics: 1) writes and 2) bytes; group by instanceID and hour
			// metrics[instanceID][writes][TIME] += 1
			// metrics[instanceID][bytes][TIME] += len(decodedData)  // or something like this
		}

		// read bigtable to validate pipeline
		Engine.Tables.ReadRecordsFromTable(instanceID, []byte{1, 2, 3})

		return nil
	})
}
