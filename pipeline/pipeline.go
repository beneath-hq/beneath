package pipeline

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/sync/errgroup"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/engine/driver"
	pb "github.com/beneath-core/engine/proto"
	pbgw "github.com/beneath-core/gateway/grpc/proto"
	"github.com/beneath-core/pkg/log"
	"github.com/beneath-core/pkg/timeutil"
)

const (
	expirationBuffer = 5 * time.Minute
)

// Run subscribes to new write requests and stores data in derived systems.
// It runs forever unless an error occcurs.
func Run() error {
	return hub.Engine.ReadWriteRequests(processWriteRequest)
}

// ProcessWriteRequest persists a write request
func processWriteRequest(ctx context.Context, req *pb.WriteRequest) error {
	// metrics to track
	start := time.Now()
	var bytesTotal int
	var minTimestamp int64

	// lookup stream
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return fmt.Errorf("cached stream is null for instance id %s", instanceID.String())
	}

	// compute sensible timestamp at which to not even attempt the write
	expirationTimestamp := int64(0)
	if stream.RetentionSeconds != 0 {
		ts := time.Now().Add(-1 * time.Duration(stream.RetentionSeconds) * time.Second).Add(expirationBuffer)
		expirationTimestamp = timeutil.UnixMilli(ts)
	}

	// make records array
	expired := 0
	records := make([]driver.Record, len(req.Records))
	for idx, proto := range req.Records {
		if expirationTimestamp != 0 && proto.Timestamp < expirationTimestamp {
			expired++
			continue
		}

		r := newRecord(stream, proto)
		bytesTotal += len(r.GetAvro())
		records[idx] = r

		if minTimestamp == 0 || proto.Timestamp < minTimestamp {
			minTimestamp = proto.Timestamp
		}
	}

	// remove expired slots from records
	if expired != 0 {
		records = records[:len(records)-expired]

		// there's a chance all records were expired, in which case we'll just stop here
		if len(records) == 0 {
			return nil
		}
	}

	// NOTE: Crashing after writing to the log (but before returning) will cause the write to be retried,
	// hence ensuring eventual consistency. But the records will appear multiple times in the log. That
	// is acceptable within our at-least-once semantics, but we want to avoid it as much as possible.

	// overriding ctx to the background context in an attempt to push through with all the writes
	// (a cancel is most likely due to receiving a SIGINT/SIGTERM, so we'll have a little leeway before being force killed)
	ctx = context.Background()
	group, ctx := errgroup.WithContext(ctx)

	// write to lookup and warehouse
	group.Go(func() error {
		return hub.Engine.Lookup.WriteRecords(ctx, stream, stream, stream, records)
	})

	group.Go(func() error {
		return hub.Engine.Warehouse.WriteToWarehouse(ctx, stream, stream, stream, records)
	})

	err := group.Wait()
	if err != nil {
		return err
	}

	// publish write report (used for streaming updates)
	err = hub.Engine.QueueWriteReport(ctx, &pb.WriteReport{
		WriteId:      req.WriteId,
		InstanceId:   instanceID.Bytes(),
		RecordsCount: int32(len(req.Records)),
		BytesTotal:   int32(bytesTotal),
	})
	if err != nil {
		return err
	}

	// finalise metrics
	elapsed := time.Since(start)
	log.S.Infow(
		"pipeline write",
		"project", stream.ProjectName,
		"stream", stream.StreamName,
		"instance", instanceID.String(),
		"records", len(req.Records),
		"bytes", bytesTotal,
		"elapsed", elapsed,
	)

	// done
	return nil
}

// record implements driver.Record
type record struct {
	Proto      *pbgw.Record
	Structured map[string]interface{}
}

func newRecord(stream driver.Stream, proto *pbgw.Record) record {
	structured, err := stream.GetCodec().UnmarshalAvro(proto.AvroData)
	if err != nil {
		panic(err)
	}

	structured, err = stream.GetCodec().ConvertFromAvroNative(structured, false)
	if err != nil {
		panic(err)
	}

	return record{
		Proto:      proto,
		Structured: structured,
	}
}

func (r record) GetTimestamp() time.Time {
	return timeutil.FromUnixMilli(r.Proto.Timestamp)
}

func (r record) GetAvro() []byte {
	return r.Proto.AvroData
}

func (r record) GetStructured() map[string]interface{} {
	return r.Structured
}
