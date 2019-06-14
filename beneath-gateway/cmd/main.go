package main

import (
	"log"

	"github.com/beneath-core/beneath-gateway/beneath"
	"github.com/beneath-core/beneath-gateway/gateway"
	"golang.org/x/sync/errgroup"
)

func main() {
	// get ports
	httpPort := beneath.Config.HTTPPort
	grpcPort := beneath.Config.GRPCPort

	// coordinates multiple servers
	group := new(errgroup.Group)

	// http server
	group.Go(func() error {
		log.Printf("HTTP server running on port %d\n", httpPort)
		return gateway.ListenAndServeHTTP(httpPort)
	})

	// gRPC server
	group.Go(func() error {
		log.Printf("gRPC server running on port %d\n", grpcPort)
		return gateway.ListenAndServeGRPC(grpcPort)
	})

	// run simultaneously
	log.Fatal(group.Wait())
}
