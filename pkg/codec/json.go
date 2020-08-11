package codec

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/pkg/schemalang"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

var one = big.NewInt(1)

type jsonConverter struct {
	Refs map[string]schemalang.Schema
}

func (c jsonConverter) convert(schema schemalang.Schema, val interface{}, fromJSON bool) (interface{}, error) {
	if val == nil {
		return val, nil
	}

	// recurses on all cases, except Primitive and Fixed, for which we need to convert val
	switch s := schema.(type) {
	case *schemalang.Primitive:
		return c.convertPrimitive(s, val, fromJSON)
	case *schemalang.Fixed:
		return c.convertBytes(val, fromJSON)
	case *schemalang.Nullable:
		return c.convert(s.NonNullType, val, fromJSON)
	case *schemalang.Array:
		arr, ok := val.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array value for schema <array>, got %v", val)
		}
		res := make([]interface{}, len(arr))
		for i, v := range arr {
			item, err := c.convert(s.ItemType, v, fromJSON)
			if err != nil {
				return nil, err
			}
			res[i] = item
		}
		return res, nil
	case *schemalang.Record:
		record, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected map value for schema <record>, got %v", val)
		}
		res := make(map[string]interface{}, len(record))
		for _, field := range s.Fields {
			fieldVal, err := c.convert(field.Type, record[field.Name], fromJSON)
			if err != nil {
				return nil, err
			}
			res[field.Name] = fieldVal
		}
		return res, nil
	case *schemalang.Ref:
		return c.convert(c.Refs[s.Name], val, fromJSON)
	case *schemalang.Enum:
		return val, nil
	default:
		panic(fmt.Errorf("unexpected Avro schema type %T", s))
	}
}

func (c jsonConverter) convertPrimitive(schema *schemalang.Primitive, val interface{}, fromJSON bool) (interface{}, error) {
	switch schema.Type {
	case schemalang.StringType:
		if schema.LogicalType == schemalang.UUIDLogicalType {
			return c.convertUUID(val, fromJSON)
		}
		return val, nil
	case schemalang.BytesType:
		if schema.LogicalType == schemalang.NumericLogicalType {
			return c.convertNumeric(val, fromJSON)
		}
		return c.convertBytes(val, fromJSON)
	case schemalang.IntType:
		return c.convertIntOrLong(val, fromJSON)
	case schemalang.LongType:
		if schema.LogicalType == schemalang.TimestampMillisLogicalType {
			return c.convertTimestamp(val, fromJSON)
		}
		return c.convertIntOrLong(val, fromJSON)
	case schemalang.FloatType:
		return c.convertFloatOrDouble(val, fromJSON)
	case schemalang.DoubleType:
		return c.convertFloatOrDouble(val, fromJSON)
	case schemalang.BooleanType:
		return val, nil
	default:
		panic(fmt.Errorf("unexpected primitive type '%s'", schema.Type))
	}
}

func (c jsonConverter) convertBytes(val interface{}, fromJSON bool) (interface{}, error) {
	if fromJSON {
		if str, ok := val.(string); ok {
			bytes, err := base64.StdEncoding.DecodeString(str)
			if err == nil {
				return bytes, nil
			}
		}
		return nil, fmt.Errorf("expected base64-encoded string for <bytes> or <fixed> schema, got %v", val)
	}

	if bytes, ok := val.([]byte); ok {
		return base64.StdEncoding.EncodeToString(bytes), nil
	}
	return nil, fmt.Errorf("expected bytes value for <bytes> or <fixed> schema, got %v", val)
}

func (c jsonConverter) convertIntOrLong(val interface{}, fromJSON bool) (interface{}, error) {
	if fromJSON {
		switch t := val.(type) {
		case string:
			res, err := strconv.ParseInt(t, 0, 64)
			if err == nil {
				return res, nil
			}
		case float64:
			return int64(t), nil
		case json.Number:
			res, err := t.Int64()
			if err == nil {
				return res, nil
			}
		}
		return nil, fmt.Errorf("couldn't parse value <%v> as <int> or <long>", val)
	}

	switch num := val.(type) {
	case int:
		return float64(num), nil
	case int32:
		return float64(num), nil
	case int64:
		if num > 9007199254740992 || num < -9007199254740992 {
			return strconv.FormatInt(num, 10), nil
		}
		return float64(num), nil
	}

	return nil, fmt.Errorf("can't convert value for <int> or <long>, got %v", val)
}

func (c jsonConverter) convertFloatOrDouble(val interface{}, fromJSON bool) (interface{}, error) {
	if fromJSON {
		switch t := val.(type) {
		case string:
			if t == "NaN" {
				return math.NaN(), nil
			} else if t == "Infinity" {
				return math.Inf(1), nil
			} else if t == "-Infinity" {
				return math.Inf(-1), nil
			}
		case float64:
			return t, nil
		case json.Number:
			res, err := t.Float64()
			if err == nil {
				return res, nil
			}
		}
		return nil, fmt.Errorf("couldn't parse value <%v> as <float> or <double>", val)
	}

	switch num := val.(type) {
	case float32:
		return num, nil
	case float64:
		if math.IsNaN(num) {
			return "NaN", nil
		} else if math.IsInf(num, 1) {
			return "Infinity", nil
		} else if math.IsInf(num, -1) {
			return "-Infinity", nil
		}
		return num, nil
	}

	return nil, fmt.Errorf("expected float-point value for <float> or <double>, got %v", val)
}

func (c jsonConverter) convertNumeric(val interface{}, fromJSON bool) (interface{}, error) {
	if fromJSON {
		switch t := val.(type) {
		case int:
			return big.NewInt(int64(t)), nil
		case string:
			n := new(big.Int)
			n, ok := n.SetString(t, 0)
			if ok {
				r := new(big.Rat)
				r.SetInt(n)
				return r, nil
			}
		}
		return nil, fmt.Errorf("expected valid string value for schema <bytes.decimal>, got %v", val)
	}

	switch t := val.(type) {
	case *big.Rat:
		return t.FloatString(0), nil
	case []byte:
		// TODO: Remove soon if nobody objects to this: https://github.com/linkedin/goavro/pull/178
		n := new(big.Int)
		n.SetBytes(t)
		if len(t) > 0 && t[0]&0x80 > 0 {
			n.Sub(n, new(big.Int).Lsh(one, uint(len(t))*8))
		}
		return n.String(), nil
	}
	return nil, fmt.Errorf("expected big.Rat or bytes value for schema <bytes.decimal>, got %v", val)
}

func (c jsonConverter) convertTimestamp(val interface{}, fromJSON bool) (interface{}, error) {
	if fromJSON {
		return timeutil.Parse(val, false)
	}

	if t, ok := val.(time.Time); ok {
		l := t.UnixNano() / int64(time.Millisecond)
		return c.convertIntOrLong(l, fromJSON)
	}
	return nil, fmt.Errorf("expected time value for schema <long.timestamp-millis>, got %v", val)
}

func (c jsonConverter) convertUUID(val interface{}, fromJSON bool) (interface{}, error) {
	if fromJSON {
		if str, ok := val.(string); ok {
			uid, err := uuid.FromString(str)
			if err == nil {
				return uid, nil
			}
		}
		return nil, fmt.Errorf("expected string-encoded UUID for schema <string.uuid>, got %v", val)
	}

	if uid, ok := val.(uuid.UUID); ok {
		return uid.String(), nil
	}
	return nil, fmt.Errorf("expected UUID value for schema <string.uuid>, got %v", val)
}
