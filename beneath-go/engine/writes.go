package engine

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/beneath-core/beneath-go/proto"
)

// QueueWriteRequest queues a write request -- concretely, it results in
// the write request being written to Pubsub, then from there read by
// the data processing pipeline and written to BigTable and BigQuery
func (e *Engine) QueueWriteRequest(ctx context.Context, req *pb.WriteRecordsRequest) error {
	panic("todo")
}

// ReadWriteRequests triggers fn for every WriteRecordsRequest that's written with QueueWriteRequest
func (e *Engine) ReadWriteRequests(fn func(context.Context, *pb.WriteRecordsRequest) error) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// handles SIGINT and SIGTERM gracefully
	go func() {
		defer cancel()
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-c:
		case <-ctx.Done():
		}
	}()

	panic("todo")
}

// QueueWriteReport publishes a batch of keys + metrics to the streams driver
func (e *Engine) QueueWriteReport(ctx context.Context, rep *pb.WriteRecordsReport) error {
	panic("todo")
}

// ReadWriteReports reads messages from the Metrics topic
func (e *Engine) ReadWriteReports(fn func(context.Context, *pb.WriteRecordsReport) error) error {
	panic("todo")
}
