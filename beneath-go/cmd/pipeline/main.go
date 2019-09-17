package main

import (
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/pipeline"
)

func main() {
	log.S.Fatal(pipeline.Run())
}
