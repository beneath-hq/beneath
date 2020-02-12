package pipeline

// import (
// 	"context"
// 	"fmt"
// 	"log"

// 	"github.com/apache/beam/sdks/go/pkg/beam"
// 	"github.com/apache/beam/sdks/go/pkg/beam/io/pubsubio"
// 	"github.com/apache/beam/sdks/go/pkg/beam/runners/direct"
// )

// var (
// 	Config configSpecification
// )

// type configSpecification struct {
// 	StreamsDriver   string `envconfig:"ENGINE_STREAMS_DRIVER" required:"true"`
// 	TablesDriver    string `envconfig:"ENGINE_TABLES_DRIVER" required:"true"`
// 	WarehouseDriver string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
// 	RedisURL        string `envconfig:"CONTROL_REDIS_URL" required:"true"`
// 	PostgresURL     string `envconfig:"CONTROL_POSTGRES_URL" required:"true"`
// }

// type InputOutputSchema struct {
// 	// get this from the user's schema definition
// 	// because there are no transformations, input and output schema are the same
// 	// schema might have to be split into separate types for separate sinks
// }

// // RunIt runs the pipeline: subscribes from pubsub and sends data to BigTable and BigQuery
// func RunItBeam() { // call this function func init() or func main()? look at Go SDK tutorials for recommendation

// 	// initialize things
// 	// hub.InitPostgres(Config.PostgresURL) // initialize the Config variable; use line of code in gateway.go
// 	// hub.InitRedis(Config.RedisURL) // these two lines of code should enable us to retrieve schemas

// 	// beam.Init() is an initialization hook that must called on startup. On
// 	// distributed runners, it is used to intercept control.
// 	beam.Init()

// 	// create a Pipeline object (p) and a Scope object (s)
// 	p := beam.NewPipeline()
// 	s := p.Root()

// 	buildPipeline(s)

// 	// in production, run the pipeline using the Google Cloud Dataflow runner
// 	// need to set the GCP project, else dataflow runner will fail. the GCP project will be information embedded in the context object
// 	// log.Print(gcpopts.GetProject(ctx))
// 	/*
// 		ctx := context.Background()
// 		if err := dataflow.Execute(ctx, p); err != nil {
// 			fmt.Printf("Pipeline failed: %v", err)
// 		}
// 	*/

// 	// in development, run the pipeline using the local Direct runner
// 	ctx := context.Background()
// 	if err := direct.Execute(ctx, p); err != nil {
// 		fmt.Printf("Pipeline failed: %v", err)
// 	}
// }

// func buildPipeline(s beam.Scope) {

// 	// set options for reading from pubsub
// 	opts := &pubsubio.ReadOptions{
// 		Subscription:       "mySubscription3",
// 		IDAttribute:        "",
// 		TimestampAttribute: "",
// 		WithAttributes:     false}

// 	// read data from PubSub
// 	// to test: put "hello world" on the pubsub emulator, via cmd/pubsub_test/publisher.go file
// 	data := pubsubio.Read(s, "beneathcrypto", "test-topic", opts)
// 	log.Print(data.IsValid())

// 	// write to BigQuery
// 	// bigqueryio.Write(s, "beneathcrypto", table, data) // need to specify schema somehow?

// 	// write to BigTable
// 	// textio.Write(s, "protocol://output/path", allData) // change this to be BigTable compatible (textio is wrong)
// }

// /* Notes and extra code:

// *** this is great: https://blog.gopheracademy.com/advent-2018/apache-beam/
// ** this is good too: https://github.com/herohde/beam/blob/go-sdk/sdks/go/examples/cookbook/tornadoes/tornadoes.go
// * this might be a little helpful (though it's in Java): https://www.varunpant.com/posts/integration-testing-with-apache-beam-using-pubsub-and-bigtable-emulators-and-direct-runner

// decode protocol buffers message to get 1. instanceID, 2. avro data, 3. sequence number (not needed for hello world example)

// get schema for avro data
// stream := model.FindCachedStreamByCurrentInstanceID(instanceID)

// TODO: calculate metrics for billing purposes
// calculate 1) bytes and 2) writes; group by instance_id and hour
// see Example(MetricsReusable) in https://godoc.org/github.com/apache/beam/sdks/go/pkg/beam

// The external Go SDK will be using Splittable DoFn, when supported by the FnAPI, and has no special support for sources or sinks. Sources currently use Impulse and ParDos. Sinks are just ParDos.

// "Coder" is a contract for transforming data to/from a byte sequence. The assumption is that all data types have a coder. The user can assign their coder on a per-transform-output basis. If the user doesn’t assign a coder, the SDK assigns a default coding behavior.
// If the data type cannot be encoded by the default coder, the pipeline construction fails.

// "Create()" inserts a fixed set of values into the pipeline. Outputs a PCollection type.
// func Create(s Scope, values ...interface{}) PCollection

// what is a Scope type in the SDK? the scope is obtained from the pipeline object

// When you run locally, your Apache Beam pipeline always runs as the GCP account that you configured with the gcloud command-line tool.

// what's the difference between beamx.Run and [runner].Execute? beamx uses a flag to determine which runner to use; beamx defaults to the direct runner

// For Beam to successfully unmarshal your data, the types must also be registered. This is typically done within the init() function, by called beam.RegisterType

// "PCollection" is "pipeline" data. It can be unbounded or bounded data.
// By default, each element in a PCollection is assigned to a single GlobalWindow. Smaller windows can be specified.
// */
