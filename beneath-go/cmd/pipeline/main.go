// the executable that calls files in beneath-go/pipeline
// as little as possible code
// see other main.go files for examples

package main

import (
	"log"

	"github.com/beneath-core/beneath-go/pipeline"
)

func main() {
	log.Fatal(pipeline.Run())
}
