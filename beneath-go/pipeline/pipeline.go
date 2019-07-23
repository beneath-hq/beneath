package pipeline

import (
	//"fmt"
	//"context"
	"fmt"
	//"log"
	//"time"

	"github.com/beneath-core/beneath-go/control/db"
	"github.com/beneath-core/beneath-go/control/model"

	//"cloud.google.com/go/bigquery"
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

/* for bigquery:
type Item struct {
	Number         int       `bigquery:"number"`
	Hash           []byte    `bigquery:"hash"`
	ParentHash     []byte    `bigquery:"parentHash"`
	Timestamp      time.Time `bigquery:"timestamp"`
	Key            []byte    `bigquery:"_key"`
	SequenceNumber int       `bigquery:"_sequence_number"`
	InsertTime     time.Time `bigquery:"_insert_time"`
}
*/

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
			decodedAvro, err := stream.AvroCodec.Unmarshal(record.AvroData)
			if err != nil {
				return fmt.Errorf("unable to decode avro data")
			}

			decodedAvroMap, ok := decodedAvro.(map[string]interface{})
			if !ok {
				return fmt.Errorf("expected decoded avro data to be a map, got %T", decodedAvro)
			}

			key, err := stream.KeyCodec.Marshal(decodedAvroMap)
			if err != nil {
				return fmt.Errorf("unable to encode key")
			}

			err = Engine.Tables.WriteRecordToTable(instanceID, key, record.AvroData, record.SequenceNumber)
			// writeToBigQuery()
		}

		// read bigtable to validate pipeline
		Engine.Tables.ReadRecordsFromTable(instanceID, []byte{1, 2, 3})

		// test packet for bigquery
		/*
			testPacket := &pb.WriteRecordsRequest{
				InstanceId: uuidTestBytes,
				Records: []*pb.Record{
					&pb.Record{
						AvroData:       avroData1,
						SequenceNumber: 100, // then try out-of-order SequenceNumbers
					},
					&pb.Record{
						AvroData:       avroData2,
						SequenceNumber: 200,
					},
					&pb.Record{
						AvroData:       avroData2,
						SequenceNumber: 300,
					},
				},
			}
		*/

		/* for bigquery:
		// start bigquery client and open table
		// Creates a client.
		ctx := context.Background()
		client, err := bigquery.NewClient(ctx, "beneathcrypto")
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		// create an uploader in order to upload data to the table
		myDataset := "ethereum_2"
		tableName := "block_numbers_4e05d0ff"
		u := client.Dataset(myDataset).Table(tableName).Uploader()

		items := []*Item{
			{Number: 10,
				Hash:           []byte("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				ParentHash:     []byte("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				Timestamp:      time.Now(),
				Key:            []byte("0xzaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				SequenceNumber: 1,
				InsertTime:     time.Now()},
		}
		log.Print("got here")
		if err := u.Put(ctx, items); err != nil {
			if multiError, ok := err.(bigquery.PutMultiError); ok {
				for _, err1 := range multiError {
					for _, err2 := range err1.Errors {
						fmt.Println(err2)
					}
				}
			} else {
				fmt.Println(err)
			}
			return err
		}
		*/
		return nil
	})
}

// create data for testing
// create a uuid
// uuidTest, _ := uuid.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
// uuidTestBytes := uuidTest.Bytes()

// // pubsub will produce a protocol buffer packet (as described in engine.proto)
// // I need to unpack it via: proto.unmarshal()
// // a testRecord (as seen below) will be produced
// // for each Record in the Packet, write to BigTable
// // each packet has only one instance ID
// avroData1 := make([]byte, 10)
// avroData2 := make([]byte, 10)
// rand.Read(avroData1)
// rand.Read(avroData2)

// testPacket := &pb.WriteRecordsRequest{
// 	InstanceId: uuidTestBytes,
// 	Records: []*pb.Record{
// 		&pb.Record{
// 			AvroData:       avroData1,
// 			SequenceNumber: 100, // then try out-of-order SequenceNumbers
// 		},
// 		&pb.Record{
// 			AvroData:       avroData2,
// 			SequenceNumber: 200,
// 		},
// 		&pb.Record{
// 			AvroData:       avroData2,
// 			SequenceNumber: 300,
// 		},
// 	},
// }

// 	// use this random recordkey for testing
// 	RecordKey := make([]byte, 10)
// 	rand.Read(RecordKey)

/*  Notes, questions, and extra code

write both successful rows and failed rows to BigQuery
use kubernetes to auto-scale machines in pipeline if needed (this is a beam replacement)
get schema from benjamin's go helper functions

func writeToBigQuery() {

	// start a bigquery client
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, PROJECTID)
	if err != nil {
		// handle error
	}

	u := table.Uploader()
	// Item implements the ValueSaver interface.
	items := []*Item{
			{Name: "n1", Size: 32.6, Count: 7},
			{Name: "n2", Size: 4, Count: 2},
			{Name: "n3", Size: 101.5, Count: 1},
	}
	if err := u.Put(ctx, items); err != nil {
    // TODO: Handle error.
	}

	// // infer schema from struct
	// schema, err := bigquery.InferSchema(STRUCT)
	// if err != nil {
	// 	// handle error
	// }
}

	// writing records to bigquery:
	// To run this sample, you will need to create (or reuse) a context and
	// an instance of the bigquery client.  For example:
	// import "cloud.google.com/go/bigquery"
	// ctx := context.Background()
	// client, err := bigquery.NewClient(ctx, "beneathcrypto")
	// u := client.Dataset("platform_test").Table(INSTANCEID).Uploader()
	// items := []*Item{
	// 				// Item implements the ValueSaver interface.
	// 				{Name: "Phred Phlyntstone", Age: 32},
	// 				{Name: "Wylma Phlyntstone", Age: 29},
	// }
	// if err := u.Put(ctx, items); err != nil {
	// 				return err
	// }
*/
