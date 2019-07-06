package main

import (
	"log"

	"github.com/beneath-core/beneath-gateway/control"
)

func main() {
	log.Fatal(control.ListenAndServeHTTP(4000))
}
