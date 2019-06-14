package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/beneath-core/beneath-gateway/beneath"
	"github.com/beneath-core/beneath-gateway/beneath/proto"
	"github.com/beneath-core/beneath-gateway/gateway"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {
	// get ports
	httpPort := beneath.Config.HTTPPort
	grpcPort := beneath.Config.GRPCPort

	// coordinates multiple servers
	group := new(errgroup.Group)

	// http server
	group.Go(func() error {
		fmt.Printf("HTTP server running on port %d\n", httpPort)

		return http.ListenAndServe(fmt.Sprintf(":%d", httpPort), gateway.HTTPServer())
	})

	// gRPC server
	group.Go(func() error {
		fmt.Printf("gRPC server running on port %d\n", grpcPort)

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			return err
		}

		server := grpc.NewServer()
		proto.RegisterGatewayServer(server, &gateway.GRPCServer{})
		return server.Serve(lis)
	})

	// run simultaneously
	log.Fatal(group.Wait())
}
