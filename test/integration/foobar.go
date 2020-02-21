package integration

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

var nextIdx = 0

const alphabet = "abcdefghijklmnopqrstuuvwxyz"

type foobar struct {
	A string
	B int
	C []byte
	D time.Time
	E *big.Rat
}

func (f foobar) Native() map[string]interface{} {
	return map[string]interface{}{
		"a": f.A,
		"b": f.B,
		"c": f.C,
		"d": f.D,
		"e": map[string]interface{}{"bytes.decimal": f.E},
		"f": map[string]interface{}{"null": nil},
	}
}

func (f foobar) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		A string      `json:"a"`
		B int         `json:"b"`
		C string      `json:"c"`
		D time.Time   `json:"d"`
		E *big.Rat    `json:"e"`
		F interface{} `json:"f"`
	}{
		A: f.A,
		B: f.B,
		C: fmt.Sprintf("0x%s", hex.EncodeToString(f.C)),
		D: f.D,
		E: f.E,
		F: nil,
	})
}

func nextInt() int {
	nextIdx++
	return nextIdx
}

func nextChar() byte {
	return alphabet[nextInt()%len(alphabet)]
}

func nextString(n int) string {
	res := make([]byte, n)
	for i := 0; i < n; i++ {
		res[i] = nextChar()
	}
	return string(res)
}

func nextFoobar() foobar {
	return foobar{
		A: nextString(4),
		B: nextInt(),
		C: []byte{nextChar(), nextChar(), nextChar(), nextChar()},
		D: time.Now(),
		E: new(big.Rat).SetInt64(int64(nextInt())),
	}
}

func nextFoobars(n int) []foobar {
	res := make([]foobar, n)
	for i := 0; i < n; i++ {
		res[i] = nextFoobar()
	}
	return res
}
