package main

import (
	"log"

	"github.com/beneath-core/beneath-go/gateway"
	"golang.org/x/sync/errgroup"
)

func main() {
	// coordinates multiple servers
	group := new(errgroup.Group)

	// http server
	group.Go(func() error {
		return gateway.ListenAndServeHTTP(gateway.Config.HTTPPort)
	})

	// gRPC server
	group.Go(func() error {
		return gateway.ListenAndServeGRPC(gateway.Config.GRPCPort)
	})

	// websocket server
	group.Go(func() error {
		return gateway.ListenAndServeWS(gateway.Config.WSPort)
	})

	// run simultaneously
	log.Fatal(group.Wait())
}
