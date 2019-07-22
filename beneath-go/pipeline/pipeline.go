package pipeline

import (
	"context"
	"crypto/rand"
	"log"
	"strings"

	"cloud.google.com/go/bigtable"
	"cloud.google.com/go/pubsub"

	pb "github.com/beneath-core/beneath-go/proto"
	uuid "github.com/satori/go.uuid"
)

const (
	project = "beneathcrypto"
)

// initialize things
func init() {
	// db.InitPostgres(Config.PostgresURL) // initialize the Config variable; use line of code in gateway.go
	// db.InitRedis(Config.RedisURL) // these two lines of code should enable us to retrieve schemas
}

// RunIt runs the pipeline: subscribes from pubsub and sends data to BigTable and BigQuery
func RunIt() { // call this function func init() or func main()?

	// create data for testing
	// create a uuid
	uuidTest, _ := uuid.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	uuidTestBytes := uuidTest.Bytes()

	// pubsub will produce a protocol buffer packet (as described in engine.proto)
	// I need to unpack it via: proto.unmarshal()
	// a testRecord (as seen below) will be produced
	// for each Record in the Packet, write to BigTable
	// each packet has only one instance ID
	avroData1 := make([]byte, 10)
	avroData2 := make([]byte, 10)
	rand.Read(avroData1)
	rand.Read(avroData2)

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

	// consume data from PubSub; unpack the protocol buffer packet
	// consumeFromPubSub()

	// create a BigTable table called "main" (if it doesn't already exist)
	createBigTableMain(project, uuidTestBytes)

	// loop through records in packet and write to BigTable and BigQuery
	// do I want to open/close the client every time within writeToBigTable, or open it once outside the loop
	for _, record := range testPacket.Records {
		writeToBigTable(project, testPacket.InstanceId, record)
		// writeToBigQuery()
	}

	// inspect and validate BigTable records
	readFromBigTable(project, testPacket.InstanceId)
}

func createBigTableMain(project string, instanceID []byte) {
	// for testing:
	// "project" and "instance" parameters in NewClient() are ignored when using the emulator
	// signal that you are using the BigTable emulator by setting the BIGTABLE_EMULATOR_HOST environment variable
	// once the emulator is running, run "$(gcloud beta emulators bigtable env-init)" in a terminal to set the environment variable
	ctx := context.Background()
	adminClient, err := bigtable.NewAdminClient(ctx, project, uuid.FromBytesOrNil(instanceID).String())
	if err != nil {
		log.Fatalf("Could not create admin client: %v", err)
	}

	err = adminClient.CreateTable(ctx, "main")
	if err != nil {
		// ignore error if table already exists
		if !strings.Contains(err.Error(), "AlreadyExists") {
			log.Fatalf("Could not create table: %v", err)
		}
	}

	err = adminClient.CreateColumnFamily(ctx, "main", "columnfamily0")
	if err != nil {
		// ignore error if column family already exists
		if !strings.Contains(err.Error(), "AlreadyExists") {
			log.Fatalf("Could not create column family: %v", err)
		}
	}
}

func writeToBigTable(project string, instanceID []byte, record *pb.Record) {
	// instead of sequence no., possibly use 3rd dimension of cell timestamp as described here: https://syslog.ravelin.com/the-joy-and-pain-of-using-google-bigtable-4210604c75be
	// I should be able to do so if it is a 64-bit unsigned int
	// replace bigtable.Now()

	ctx := context.Background()

	client, err := bigtable.NewClient(ctx, project, uuid.FromBytesOrNil(instanceID).String())
	if err != nil {
		log.Fatalf("Could not create data operations client: %v", err)
	}

	tbl := client.Open("main")

	// use this random recordkey for testing
	RecordKey := make([]byte, 10)
	rand.Read(RecordKey)

	rowKeys := append(instanceID, RecordKey...)
	columnFamilyName := "columnfamily0"
	column0Name := "data"
	// column1Name := "sequenceNumber"

	log.Printf("writing to table.")
	muts := bigtable.NewMutation()
	muts.Set(columnFamilyName, column0Name, bigtable.Timestamp(record.SequenceNumber), record.AvroData)
	//muts.Set(columnFamilyName, column1Name, bigtable.Now(), []byte(record.SequenceNumber))

	err = tbl.Apply(ctx, string(rowKeys), muts)
	if err != nil {
		log.Fatalf("Could not apply row mutation: %v", err)
	}

	if err = client.Close(); err != nil {
		log.Fatalf("Could not close data operations client: %v", err)
	}
}

func readFromBigTable(project string, instanceID []byte) {

	ctx := context.Background()

	client, err := bigtable.NewClient(ctx, project, uuid.FromBytesOrNil(instanceID).String())
	if err != nil {
		log.Fatalf("Could not create data operations client: %v", err)
	}

	tbl := client.Open("main")

	log.Printf("Reading all rows:")
	rr := bigtable.PrefixRange("")
	err = tbl.ReadRows(ctx, rr, func(row bigtable.Row) bool {
		log.Print(row["columnfamily0"][0].Row)
		log.Print(row["columnfamily0"][0].Column)
		log.Print(row["columnfamily0"][0].Timestamp)
		log.Print(row["columnfamily0"][0].Value)
		return true
	})

	if err = client.Close(); err != nil {
		log.Fatalf("Could not close data operations client: %v", err)
	}
}

var topics []*pubsub.Topic
var subs []*pubsub.Subscription

/*  Notes, questions, and extra code

write both successful rows and failed rows to BigQuery
will the bigquery table already exist? or will I sometimes have to create it?
use kubernetes to auto-scale machines in pipeline if needed (this is a beam replacement)
get schema from benjamin's go helper functions

func consumeFromPubSub() {

	// set environment variables
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	os.Setenv("PUBSUB_PROJECT_ID", "beneathcrypto")

	// prepare pubsub client
	client, err := pubsub.NewClient(context.Background(), "beneathcrypto")
	if err != nil {
		log.Fatalf("could not create pubsub client: %v", err)
	}

	// create subscriber to topic
	sub, err := client.CreateSubscription(context.Background(), "mySubscriptionName3", pubsub.SubscriptionConfig{
		Topic:       client.Topic("test-topic"),
		AckDeadline: 20 * time.Second,
	})
	if err != nil {
		log.Printf(err.Error())
		log.Printf("could not subscribe to topic")
	} else {
		log.Printf("Created subscription: %v\n", sub)
	}

	// consume messages
	ctx := context.Background()
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		fmt.Printf("Got message: %q\n", string(msg.Data))
	})
	if err != nil {
		log.Fatalf("Error!")
	}
}

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
