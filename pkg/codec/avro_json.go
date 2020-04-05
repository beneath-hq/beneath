package codec

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gitlab.com/beneath-org/beneath/pkg/timeutil"
)

var one = big.NewInt(1)

// NOTE about definedTypes
// definedTypes keeps track of custom types encountered during depth-first traversal
// of the schema (as per the avro spec) -- they may later in the schema be referenced
// by name, so we need to be able to look them up.
// The only types that can be named are enum, record and fixed.
// A tricky edge case is that of types being defined in types on a path we wouldn't
// otherwise traverse -- e.g., a defined type in the array definition of an empty array.

// extractDefinedTypes is a helper function used in jsonNativeToAvroNative and avroNativeToJSONNative
// to extract and save named types as per the comment above.
// Recursive extraction is an option since we want to avoid double-traversals on types that
// jsonNativeToAvroNative or avroNativeToJSONNative will traverse anyway
func extractDefinedTypes(schemaT interface{}, definedTypes map[string]interface{}, recursively bool) {
	switch schema := schemaT.(type) {
	case []interface{}:
		if recursively {
			extractDefinedTypes(schema[1], definedTypes, recursively)
		}
	case map[string]interface{}:
		switch subtype := schema["type"].(type) {
		case string:
			switch subtype {
			case "fixed":
				definedTypes[schema["name"].(string)] = schema
			case "array":
				if recursively {
					extractDefinedTypes(schema["items"], definedTypes, recursively)
				}
			case "record":
				definedTypes[schema["name"].(string)] = schema
				if recursively {
					fields := schema["fields"].([]interface{})
					for _, field := range fields {
						extractDefinedTypes(field, definedTypes, recursively)
					}
				}
			}
		default:
			extractDefinedTypes(subtype, definedTypes, recursively)
		}
	}
}

// jsonNativeToAvroNative converts the result of unmarshalling json into
// a structure suitable as input to marshalling avro; namely, the transforms are:
//   - unions: if value is not null, wrap in map with the type's name as the key
//   - bytes: turn from string prefixed with 0x into []byte
//   - fixed: turn from string prefixed with 0x matching size into [size]byte
//   - bytes.decimal: turn from string containing a number into *big.Rat
//   - long: if value is a string, convert to long
func jsonNativeToAvroNative(schemaT interface{}, valT interface{}, definedTypes map[string]interface{}) (interface{}, error) {
	switch schema := schemaT.(type) {
	case string:
		// handle bytes case
		if schema == "bytes" {
			val, ok := valT.(string)
			if !ok {
				return nil, fmt.Errorf("expected string value for schema <bytes>, got %v", valT)
			}

			if strings.HasPrefix(val, "0x") {
				binary, err := hex.DecodeString(val[2:])
				if err == nil {
					return binary, nil
				}
			}

			return nil, fmt.Errorf("expected hex string with '0x' prefix for <bytes>, got <%v>", val)
		}

		// handle case where long is string
		if schema == "long" {
			if val, ok := valT.(string); ok {
				res, err := strconv.ParseInt(val, 0, 64)
				if err != nil {
					return nil, fmt.Errorf("couldn't parse string <%v> as long", val)
				}
				return res, nil
			}
			if val, ok := valT.(float64); ok {
				return int64(val), nil
			}
		}

		// handle case where valT is a json.Number
		if val, ok := valT.(json.Number); ok {
			switch schema {
			case "int":
				// if doesn't fit into int32, should throw error during avro encoding
				return val.Int64()
			case "long":
				return val.Int64()
			case "float":
				// if doesn't fit into float32, should throw error during avro encoding
				return val.Float64()
			case "double":
				return val.Float64()
			}
		}

		// check if it's a defined type
		if definedTypes[schema] != nil {
			return jsonNativeToAvroNative(definedTypes[schema], valT, definedTypes)
		}

		// default
		return valT, nil
	case []interface{}:
		// BuildAvroSchema guarantees that unions have signature ["null", {...}],
		// but just to be sure
		if len(schema) != 2 && schema[0] != "null" {
			return nil, fmt.Errorf("encountered illegal union %v", schema)
		}

		// handle the null case
		if valT == nil {
			// we'll be returning nil, but first we have to traverse the alternate path
			// for definedTypes, which may show up later in the schema (see note at top of file)
			extractDefinedTypes(schema[1], definedTypes, true)
			return nil, nil
		}

		// recurse
		childVal, err := jsonNativeToAvroNative(schema[1], valT, definedTypes)
		if err != nil {
			return nil, err
		}

		// handle various cases
		switch childSchema := schema[1].(type) {
		case string:
			// simple type string case
			return map[string]interface{}{childSchema: childVal}, nil
		case map[string]interface{}:
			// named type case
			if name, ok := childSchema["name"].(string); ok {
				return map[string]interface{}{name: childVal}, nil
			}
			// type as dict case
			if typeName, ok := childSchema["type"].(string); ok {
				if logicalTypeName, ok := childSchema["logicalType"].(string); ok {
					typeName += "." + logicalTypeName
				}
				return map[string]interface{}{typeName: childVal}, nil
			}
		}

		// fallback
		return nil, fmt.Errorf("couldn't parse value <%v> as union <%v>", valT, schema)
	case map[string]interface{}:
		t := schema["type"]

		// handle fixed
		if t == "fixed" {
			val, ok := valT.(string)
			if !ok {
				return nil, fmt.Errorf("expected string value for schema <fixed>, got %v", valT)
			}

			size, ok := schema["size"].(float64)
			if !ok {
				return nil, fmt.Errorf("expected number key 'size' for schema <fixed>, got %v", schema)
			}

			// save as defined type
			extractDefinedTypes(schema, definedTypes, false)

			if strings.HasPrefix(val, "0x") && len(val) == 2*int(size)+2 {
				binary, err := hex.DecodeString(val[2:])
				if err == nil {
					return binary, nil
				}
			}

			return nil, fmt.Errorf("expected hex string with '0x' prefix of size %d for <fixed>, got <%v>", int(size), val)
		}

		// handle logical types
		if lt := schema["logicalType"]; lt != nil {
			// handle bytes.decimal
			if t == "bytes" && lt == "decimal" {
				switch val := valT.(type) {
				case int:
					return big.NewInt(int64(val)), nil
				case string:
					n := new(big.Int)
					n, ok := n.SetString(val, 0)
					if !ok {
						return nil, fmt.Errorf("couldn't convert string to decimal: %v", val)
					}
					r := new(big.Rat)
					r.SetInt(n)
					return r, nil
				default:
					return nil, fmt.Errorf("expected string value for schema <bytes.decimal>, got %v", valT)
				}
			}

			// handle long.timestamp-millis
			if t == "long" && lt == "timestamp-millis" {
				return timeutil.Parse(valT, false)
			}
		}

		// handle array case
		if t == "array" {
			val, ok := valT.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array value for schema <array>, got %v", valT)
			}

			res := make([]interface{}, len(val))
			for i, v := range val {
				r, err := jsonNativeToAvroNative(schema["items"], v, definedTypes)
				if err != nil {
					return nil, err
				}
				res[i] = r
			}

			if len(val) == 0 {
				// when len == 0, no calls happen to jsonNativeToAvroNative
				// so we have to manually make sure to extract types for schema["items"]
				extractDefinedTypes(schema["items"], definedTypes, true)
			}

			return res, nil
		}

		if t == "record" {
			val, ok := valT.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected dict value for schema <record>, got %v", valT)
			}

			fields, ok := schema["fields"].([]interface{})
			if !ok {
				return nil, fmt.Errorf("record type doesn't have fields: %v", schema)
			}

			// save as defined type
			extractDefinedTypes(schema, definedTypes, false)

			for _, fieldT := range fields {
				field, ok := fieldT.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("expected map for record field, got %v", fieldT)
				}

				fieldName, ok := field["name"].(string)
				if !ok {
					return nil, fmt.Errorf("expected field name to be string, not %v", fieldName)
				}

				if val[fieldName] != nil {
					r, err := jsonNativeToAvroNative(field, val[fieldName], definedTypes)
					if err != nil {
						return nil, err
					}
					val[fieldName] = r
				} else {
					// make sure to traverse the field for defined types, even though
					// the value is empty (see note at top of file)
					extractDefinedTypes(field, definedTypes, true)
				}
			}

			return val, nil
		}
		return jsonNativeToAvroNative(t, valT, definedTypes)
	default:
		return nil, fmt.Errorf("unknown schema type: %T", schemaT)
	}
}

// avroNativeToJSONNative converts the result of unmarshalling avro into
// a structure suitable as input to marshalling JSON; namely, the transforms are:
//   - unions: if value is not null, remove the map that wraps around the value (where the type name is the key)
//   - bytes: turn from []byte into string prefixed with 0x
//   - fixed: same as for bytes
//   - bytes.decimal: turn into string representing the number in decimal
func avroNativeToJSONNative(schemaT interface{}, valT interface{}, definedTypes map[string]interface{}, convertTypes bool) (interface{}, error) {
	switch schema := schemaT.(type) {
	case string:
		// handle bytes case
		if schema == "bytes" {
			val, ok := valT.([]byte)
			if !ok {
				return nil, fmt.Errorf("expected []byte value for schema <bytes>, got %v", valT)
			}

			if convertTypes {
				return "0x" + hex.EncodeToString(val), nil
			}
			return val, nil
		}

		// handle case where long overflows float64
		if convertTypes {
			if schema == "long" {
				if val, ok := valT.(int64); ok {
					if val > 9007199254740992 || val < -9007199254740992 {
						return strconv.FormatInt(val, 10), nil
					}
					return float64(val), nil
				}
			}
		}

		// check if it's a defined type
		if definedTypes[schema] != nil {
			return avroNativeToJSONNative(definedTypes[schema], valT, definedTypes, convertTypes)
		}

		// default
		return valT, nil
	case []interface{}:
		// BuildAvroSchema guarantees that unions have signature ["null", {...}],
		// but just to be sure
		if len(schema) != 2 && schema[0] != "null" {
			return nil, fmt.Errorf("encountered illegal union %v", schema)
		}

		// handle the null case
		if valT == nil {
			// we're returning nil, but must make sure we pick up on all
			// defined types on the alternative path (see note at top of file)
			extractDefinedTypes(schema[1], definedTypes, true)
			return nil, nil
		}

		// unwrap the type and check length
		val, ok := valT.(map[string]interface{})
		if !ok || len(val) != 1 {
			return nil, fmt.Errorf("expected native avro for union to be a map with one key")
		}

		// return wrapped value
		for _, v := range val {
			return avroNativeToJSONNative(schema[1], v, definedTypes, convertTypes)
		}

		// not actually reachable
		return nil, nil
	case map[string]interface{}:
		t := schema["type"]

		// handle fixed
		if t == "fixed" {
			val, ok := valT.([]byte)
			if !ok {
				return nil, fmt.Errorf("expected []byte value for schema <fixed>, got %v", valT)
			}

			// save as defined type
			extractDefinedTypes(schema, definedTypes, false)

			if convertTypes {
				return "0x" + hex.EncodeToString(val), nil
			}
			return val, nil
		}

		// handle logical type
		if lt := schema["logicalType"]; lt != nil {
			// handle bytes.decimal
			if t == "bytes" && lt == "decimal" {
				switch val := valT.(type) {
				case *big.Rat:
					if convertTypes {
						return val.FloatString(0), nil
					}
					return val, nil
				case []byte:
					// TODO: Remove soon if nobody objects to this:https://github.com/linkedin/goavro/pull/178
					n := new(big.Int)
					n.SetBytes(val)
					if len(val) > 0 && val[0]&0x80 > 0 {
						n.Sub(n, new(big.Int).Lsh(one, uint(len(val))*8))
					}
					if convertTypes {
						return n.String(), nil
					}
					return n, nil
				default:
					return nil, fmt.Errorf("expected big.Rat or bytes value for schema <bytes.decimal>, got %v", valT)
				}
			}

			// handle long.timestamp-millis
			if t == "long" && lt == "timestamp-millis" {
				val, ok := valT.(time.Time)
				if !ok {
					return nil, fmt.Errorf("expected time.time value for schema <long.timestamp-millis>, got %v", valT)
				}
				if convertTypes {
					l := val.UnixNano() / int64(time.Millisecond)
					return avroNativeToJSONNative(t, l, definedTypes, convertTypes)
				}
				return val, nil
			}
		}

		// handle array case
		if t == "array" {
			val, ok := valT.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array value for schema <array>, got %v", valT)
			}

			res := make([]interface{}, len(val))
			for i, v := range val {
				r, err := avroNativeToJSONNative(schema["items"], v, definedTypes, convertTypes)
				if err != nil {
					return nil, err
				}
				res[i] = r
			}

			if len(val) == 0 {
				// when len == 0, no calls happen to avroNativeToJSONNative
				// so we have to manually make sure to extract types for schema["items"]
				extractDefinedTypes(schema["items"], definedTypes, true)
			}

			return res, nil
		}

		if t == "record" {
			val, ok := valT.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected dict value for schema <record>, got %v", valT)
			}

			fields, ok := schema["fields"].([]interface{})
			if !ok {
				return nil, fmt.Errorf("record type doesn't have fields: %v", schema)
			}

			// save as defined type
			extractDefinedTypes(schema, definedTypes, false)

			for _, fieldT := range fields {
				field, ok := fieldT.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("expected map for record field, got %v", fieldT)
				}

				fieldName, ok := field["name"].(string)
				if !ok {
					return nil, fmt.Errorf("expected field name to be string, not %v", fieldName)
				}

				if val[fieldName] != nil {
					r, err := avroNativeToJSONNative(field, val[fieldName], definedTypes, convertTypes)
					if err != nil {
						return nil, err
					}
					val[fieldName] = r
				} else {
					// make sure to traverse the field for defined types, even though
					// the value is empty (see note at top of file)
					extractDefinedTypes(field, definedTypes, true)
					val[fieldName] = nil
				}
			}

			return val, nil
		}

		return avroNativeToJSONNative(t, valT, definedTypes, convertTypes)
	default:
		return nil, fmt.Errorf("unknown schema type: %T", schemaT)
	}
}
