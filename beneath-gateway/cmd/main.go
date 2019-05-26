package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/beneath-core/beneath-gateway/pkg/beneath"
	"github.com/beneath-core/beneath-gateway/pkg/rest"
)

func main() {
	port := beneath.Config.Port

	fmt.Printf("Running on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), rest.GetHandler()))

	// TODO: Add grpc
}
