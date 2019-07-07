package main

import (
	"log"

	"github.com/beneath-core/beneath-go/control"
)

func main() {
	log.Fatal(control.ListenAndServeHTTP(control.Config.HTTPPort))
}
