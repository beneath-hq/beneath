package main

import (
	"github.com/beneath-core/beneath-go/control"
	"github.com/beneath-core/beneath-go/core/log"
)

func main() {
	log.S.Fatal(control.ListenAndServeHTTP(control.Config.ControlPort))
}
