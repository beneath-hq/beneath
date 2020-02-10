package grpc

import (
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-core/beneath-go/gateway"
	"github.com/beneath-core/beneath-go/gateway/subscriptions"
	pb "github.com/beneath-core/beneath-go/proto"
)

func (s *gRPCServer) Subscribe(req *pb.SubscribeRequest, ss pb.Gateway_SubscribeServer) error {
	// read instanceID
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	if instanceID == uuid.Nil {
		return status.Error(codes.InvalidArgument, "instance_id not valid UUID")
	}

	// get subscription channel
	ch := make(chan subscriptions.Message)
	cancel, err := gateway.Subscriptions.Subscribe(instanceID, req.Cursor, func(msg subscriptions.Message) {
		ch <- msg
	})
	if err != nil {
		return err
	}
	defer cancel()

	// read messages until disconnect
	for {
		select {
		case <-ss.Context().Done():
			return nil
		case _ = <-ch:
			err := ss.Send(&pb.SubscribeResponse{})
			if err != nil {
				return err
			}
		}
	}
}
