package main

import (
	"github.com/beneath-core/control"
	"github.com/beneath-core/core/log"
)

func main() {
	log.S.Fatal(control.ListenAndServeHTTP(control.Config.ControlPort))
}
