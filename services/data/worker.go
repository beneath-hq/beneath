package data

import (
	"context"
	"fmt"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/sync/errgroup"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	pb "github.com/beneath-hq/beneath/infra/engine/proto"
	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/pkg/timeutil"
	pbgw "github.com/beneath-hq/beneath/server/data/grpc/proto"
)

// RunWorker subscribes to new write requests and stores data in derived systems.
// It runs forever unless an error occcurs.
func (s *Service) RunWorker(ctx context.Context) error {
	return s.ReadWriteRequests(ctx, s.processWriteRequest)
}

// ProcessWriteRequest persists a write request
func (s *Service) processWriteRequest(ctx context.Context, req *pb.WriteRequest) error {
	// metrics to track
	start := time.Now()
	bytesTotal := 0
	recordsCount := 0
	mu := sync.Mutex{}

	// NOTE: Crashing after writing to the log (but before returning) will cause the write to be retried,
	// hence ensuring eventual consistency. But the records will appear multiple times in the log. That
	// is acceptable within our at-least-once semantics, but we want to avoid it as much as possible.

	// overriding ctx to the background context in an attempt to push through with all the writes
	// (a cancel is most likely due to receiving a SIGINT/SIGTERM, so we'll have a little leeway before being force killed)
	ctx = context.Background()

	// concurrently process each InstanceRecords
	group, cctx := errgroup.WithContext(ctx)
	for idx := range req.InstanceRecords {
		ir := req.InstanceRecords[idx] // see https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		group.Go(func() error {
			// process
			bytesWritten, err := s.processInstanceRecords(cctx, req.WriteId, ir)
			if err != nil {
				return err
			}

			// track metrics
			mu.Lock()
			bytesTotal += bytesWritten
			recordsCount += len(ir.Records)
			mu.Unlock()

			// done
			return nil
		})
	}

	// wait for group
	err := group.Wait()
	if err != nil {
		return err
	}

	// finalise metrics
	elapsed := time.Since(start)
	s.Logger.Infow(
		"records write",
		"write_id", req.WriteId,
		"records", recordsCount,
		"bytes", bytesTotal,
		"elapsed", elapsed,
	)

	// done
	return nil
}

func (s *Service) processInstanceRecords(ctx context.Context, writeID []byte, ir *pbgw.InstanceRecords) (int, error) {
	// lookup table
	instanceID := uuid.FromBytesOrNil(ir.InstanceId)
	table := s.Tables.FindCachedInstance(ctx, instanceID)
	if table == nil {
		// TODO: use dead letter queue that retries
		s.Logger.Errorw("instance not found", "instance", instanceID.String(), "records", ir.Records)
		return 0, nil
	}

	// make records array
	var bytesWritten int
	records := make([]driver.Record, len(ir.Records))
	for idx, proto := range ir.Records {
		r := newRecord(table, proto)
		bytesWritten += len(r.GetAvro())
		records[idx] = r
	}

	// use errgroup for concurrently writing to lookup and warehouse
	group, cctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return s.Engine.Lookup.WriteRecords(cctx, table, table, models.EfficientTableInstance(instanceID), records)
	})

	if table.UseWarehouse {
		group.Go(func() error {
			return s.Engine.Warehouse.WriteToWarehouse(cctx, table, table, models.EfficientTableInstance(instanceID), records)
		})
	}

	err := group.Wait()
	if err != nil {
		return 0, err
	}

	// publish write report (used for streaming updates)
	err = s.QueueWriteReport(ctx, &pb.WriteReport{
		WriteId:      writeID,
		InstanceId:   ir.InstanceId,
		RecordsCount: int32(len(ir.Records)),
		BytesTotal:   int32(bytesWritten),
	})
	if err != nil {
		return 0, err
	}

	return bytesWritten, nil
}

// record implements driver.Record
type record struct {
	Proto      *pbgw.Record
	Structured map[string]interface{}
}

func newRecord(table driver.Table, proto *pbgw.Record) record {
	structured, err := table.GetCodec().UnmarshalAvro(proto.AvroData)
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

func (r record) GetJSON() map[string]interface{} {
	panic(fmt.Errorf("not implemented"))
}

func (r record) GetPrimaryKey() []byte {
	panic(fmt.Errorf("not implemented"))
}
