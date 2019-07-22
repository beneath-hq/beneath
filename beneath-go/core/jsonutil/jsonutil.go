package jsonutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
)

// Marshal is equivalent to json.Marshal
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal is similar to json.Unmarshal, but decodes numbers as json.Number instead of float64
func Unmarshal(r io.Reader, v interface{}) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec.Decode(v)
}

// UnmarshalBytes is similar to Unmarshal, but takes a bytes slice instead of an io.Reader
func UnmarshalBytes(data []byte, v interface{}) error {
	return Unmarshal(bytes.NewReader(data), v)
}

// ParseUint64 is a helper for safely parsing values found in JSON to uint64
func ParseUint64(val interface{}) (uint64, error) {
	if val == nil {
		return 0, fmt.Errorf("num is nil")
	}

	switch num := val.(type) {
	case string:
		return strconv.ParseUint(num, 10, 64)
	case json.Number:
		return strconv.ParseUint(num.String(), 10, 64)
	}

	log.Panicf("unrecognized json numeric type: %T", val)
	return 0, nil
}
