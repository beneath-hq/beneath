package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "gitlab.com/beneath-hq/beneath/server/data/grpc/proto"
	"gitlab.com/beneath-hq/beneath/services/data"
	"gitlab.com/beneath-hq/beneath/services/middleware"
)

func (s *gRPCServer) Subscribe(req *pb.SubscribeRequest, ss pb.Gateway_SubscribeServer) error {
	// get auth
	secret := middleware.GetSecret(ss.Context())
	if secret.IsAnonymous() {
		return grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// get subscription channel
	ch := make(chan data.SubscriptionMessage)
	res, err := s.Service.HandleSubscribe(ss.Context(), &data.SubscribeRequest{
		Secret: secret,
		Cursor: req.Cursor,
		Cb: func(msg data.SubscriptionMessage) {
			ch <- msg
		},
	})
	if err != nil {
		return err
	}
	defer res.Cancel()

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
