package jsonutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// Marshal is equivalent to json.Marshal
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalWriter is similar to Marshal, but encodes directly to a writer
func MarshalWriter(v interface{}, w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(v)
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

// ParseInt64 is a helper for safely parsing values found in JSON to int64
func ParseInt64(val interface{}) (int64, error) {
	if val == nil {
		return 0, fmt.Errorf("num is nil")
	}

	switch num := val.(type) {
	case string:
		return strconv.ParseInt(num, 10, 64)
	case json.Number:
		return num.Int64()
	}

	panic(fmt.Errorf("unrecognized json numeric type: %T", val))
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

	panic(fmt.Errorf("unrecognized json numeric type: %T", val))
}
