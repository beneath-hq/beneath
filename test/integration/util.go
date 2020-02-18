package integration

import (
	"fmt"
	"io/ioutil"
)

func readTestdata(filename string) string {
	schema, err := ioutil.ReadFile(fmt.Sprintf("../data/%s", filename))
	panicIf(err)
	return string(schema)
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}
