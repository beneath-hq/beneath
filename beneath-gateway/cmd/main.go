package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/beneath-core/beneath-gateway/beneath"
	"github.com/beneath-core/beneath-gateway/gateway"
)

func main() {
	port := beneath.Config.Port

	fmt.Printf("Running on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), gateway.GetHandler()))

	// TODO: Add grpc
}
