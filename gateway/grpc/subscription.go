package grpc

import (
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-org/beneath/control/entity"
	"gitlab.com/beneath-org/beneath/internal/middleware"
	"gitlab.com/beneath-org/beneath/gateway"
	"gitlab.com/beneath-org/beneath/gateway/subscriptions"
	pb "gitlab.com/beneath-org/beneath/gateway/grpc/proto"
)

func (s *gRPCServer) Subscribe(req *pb.SubscribeRequest, ss pb.Gateway_SubscribeServer) error {
	// get auth
	secret := middleware.GetSecret(ss.Context())
	if secret == nil {
		return grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// read instanceID
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	if instanceID == uuid.Nil {
		return status.Error(codes.InvalidArgument, "instance_id not valid UUID")
	}

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ss.Context(), instanceID)
	if stream == nil {
		return status.Error(codes.NotFound, "stream not found")
	}

	// if batch, check committed
	if stream.Batch && !stream.Committed {
		return status.Error(codes.FailedPrecondition, "batch has not yet been committed, and so can't be read")
	}

	// check permissions
	perms := secret.StreamPermissions(ss.Context(), stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Read {
		return grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
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
