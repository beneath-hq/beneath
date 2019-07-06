package main

import (
	"log"

	"github.com/beneath-core/beneath-go/control"
)

func main() {
	log.Fatal(control.ListenAndServeHTTP(4000))
}
