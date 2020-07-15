package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/gateway"
	pb "gitlab.com/beneath-hq/beneath/gateway/grpc/proto"
	"gitlab.com/beneath-hq/beneath/gateway/subscriptions"
	"gitlab.com/beneath-hq/beneath/gateway/util"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
)

func (s *gRPCServer) Subscribe(req *pb.SubscribeRequest, ss pb.Gateway_SubscribeServer) error {
	// get auth
	secret := middleware.GetSecret(ss.Context())
	if secret == nil {
		return grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// parse cursor
	cursor, err := util.CursorFromBytes(req.Cursor)
	if err != nil {
		return grpc.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	// ensure log cursor
	if cursor.GetType() != util.LogCursorType {
		return grpc.Errorf(codes.InvalidArgument, "cannot subscribe to non-log cursor")
	}

	// read instanceID
	instanceID := cursor.GetID()

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ss.Context(), instanceID)
	if stream == nil {
		return status.Error(codes.NotFound, "stream not found")
	}

	// check permissions
	perms := secret.StreamPermissions(ss.Context(), stream.StreamID, stream.ProjectID, stream.Public)
	if !perms.Read {
		return grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// get subscription channel
	ch := make(chan subscriptions.Message)
	cancel, err := gateway.Subscriptions.Subscribe(instanceID, cursor.GetPayload(), func(msg subscriptions.Message) {
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
