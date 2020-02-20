package integration

import (
	"fmt"
	"io/ioutil"
)

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func readTestdata(filename string) string {
	schema, err := ioutil.ReadFile(fmt.Sprintf("test/testdata/%s", filename))
	panicIf(err)
	return string(schema)
}
