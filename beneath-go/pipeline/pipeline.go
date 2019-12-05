package pipeline

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
	pb "github.com/beneath-core/beneath-go/proto"
)

type configSpecification struct {
	MQDriver         string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LogDriver        string `envconfig:"ENGINE_LOG_DRIVER" required:"true"`
	LookupDriver     string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver  string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL         string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresUser     string `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`
}

var (
	// Config for gateway
	Config configSpecification
)

func init() {
	core.LoadConfig("beneath", &Config)

	db.InitPostgres(Config.PostgresHost, Config.PostgresUser, Config.PostgresPassword)
	db.InitRedis(Config.RedisURL)
	db.InitEngine(Config.MQDriver, Config.LogDriver, Config.LookupDriver, Config.WarehouseDriver)
}

// Run runs the pipeline: subscribes from pubsub and sends data to BigTable and BigQuery
func Run() error {
	// log that we're running
	log.S.Info("pipeline started")

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
func processWriteRequest(ctx context.Context, req *pb.WriteRecordsRequest) error {
	// metrics to track
	startTime := time.Now()
	var bytesWritten int64

	// lookup stream for write request
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return fmt.Errorf("cached stream is null for instanceid %s", instanceID.String())
	}

	// keep track of keys to publish
	n := len(req.Records)
	keys := make([][]byte, n)
	avros := make([][]byte, n)
	records := make([]map[string]interface{}, n)
	timestamps := make([]time.Time, n)

	// loop through each record in the Write Request
	for i, record := range req.Records {
		// set avro data
		avros[i] = record.AvroData

		// decode the avro data
		obj, err := stream.Codec.UnmarshalAvro(record.AvroData)
		if err != nil {
			return fmt.Errorf("unable to decode avro data: %v", err.Error())
		}
		records[i], err = stream.Codec.ConvertFromAvroNative(obj, false)
		if err != nil {
			return fmt.Errorf("unable to decode avro data: %v", err.Error())
		}

		// get the encoded key
		keys[i], err = stream.Codec.MarshalKey(records[i])
		if err != nil {
			return fmt.Errorf("unable to encode key")
		}

		// save timestamp
		timestamps[i] = timeutil.FromUnixMilli(record.Timestamp)

		// increment metrics
		bytesWritten += int64(len(record.AvroData))
	}

	// writing encoded data to Table
	err := db.Engine.Tables.WriteRecords(ctx, instanceID, keys, avros, timestamps, !stream.Batch)
	if err != nil {
		return err
	}

	// writing decoded data to Warehouse
	err = db.Engine.Warehouse.WriteRecords(ctx, stream.ProjectName, stream.StreamName, instanceID, keys, avros, records, timestamps)
	if err != nil {
		return err
	}

	// publish metrics packet; the keys in the packet will be used to stream data via the gateway's websocket
	err = db.Engine.Streams.QueueWriteReport(ctx, &pb.WriteRecordsReport{
		InstanceId:   instanceID.Bytes(),
		Keys:         keys,
		BytesWritten: bytesWritten,
	})
	if err != nil {
		return err
	}

	// finalise metrics
	elapsed := time.Since(startTime)
	log.S.Infow(
		"pipeline write",
		"project", stream.ProjectName,
		"stream", stream.StreamName,
		"instance", instanceID.String(),
		"records", len(req.Records),
		"bytes", bytesWritten,
		"elapsed", elapsed,
	)

	// done
	return nil
}
