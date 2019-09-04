// the executable that calls files in beneath-go/pipeline
// as little as possible code
// see other main.go files for examples

package main

import (
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/pipeline"
)

func main() {
	log.S.Fatal(pipeline.Run())
}
