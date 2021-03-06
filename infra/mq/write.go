package mq

// import (
// 	"context"
// 	"fmt"

// 	"github.com/golang/protobuf/proto"

// 	pb "github.com/beneath-hq/beneath/infra/engine/proto"
// )

// const (
// 	writeRequestsTopic        = "write-requests"
// 	writeRequestsSubscription = "write-requests-worker"
// 	writeReportsTopic         = "write-reports"
// 	writeReportsSubscription  = "write-reports-reader"
// )

// // QueueWriteRequest queues a write request -- concretely, it results in
// // the write request being written to Pubsub, then from there read by
// // the data processing pipeline and written to BigTable and BigQuery
// func (e *MQ) QueueWriteRequest(ctx context.Context, req *pb.WriteRequest) error {
// 	msg, err := proto.Marshal(req)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if len(msg) > e.MQ.MaxMessageSize() {
// 		return fmt.Errorf("total write size <%d bytes> exceeds maximum <%d bytes>", len(msg), e.MQ.MaxMessageSize())
// 	}
// 	return e.MQ.Publish(ctx, writeRequestsTopic, msg)
// }

// // ReadWriteRequests triggers fn for every WriteRecordsRequest that's written with QueueWriteRequest
// func (e *MQ) ReadWriteRequests(ctx context.Context, fn func(context.Context, *pb.WriteRequest) error) error {
// 	return e.MQ.Subscribe(ctx, writeRequestsTopic, writeRequestsSubscription, true, func(ctx context.Context, msg []byte) error {
// 		req := &pb.WriteRequest{}
// 		err := proto.Unmarshal(msg, req)
// 		if err != nil {
// 			return err
// 		}
// 		return fn(ctx, req)
// 	})
// }

// // QueueWriteReport publishes a WriteReport (used to notify of completed processing of a WriteRequest)
// func (e *Engine) QueueWriteReport(ctx context.Context, rep *pb.WriteReport) error {
// 	msg, err := proto.Marshal(rep)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if len(msg) > e.MQ.MaxMessageSize() {
// 		return fmt.Errorf("total write report size <%d bytes> exceeds maximum <%d bytes>", len(msg), e.MQ.MaxMessageSize())
// 	}
// 	return e.MQ.Publish(ctx, writeReportsTopic, msg)
// }

// // ReadWriteReports reads messages published with QueueWriteReport
// func (e *Engine) ReadWriteReports(ctx context.Context, fn func(context.Context, *pb.WriteReport) error) error {
// 	return e.MQ.Subscribe(ctx, writeReportsTopic, writeReportsSubscription, false, func(ctx context.Context, msg []byte) error {
// 		rep := &pb.WriteReport{}
// 		err := proto.Unmarshal(msg, rep)
// 		if err != nil {
// 			return err
// 		}
// 		return fn(ctx, rep)
// 	})
// }
