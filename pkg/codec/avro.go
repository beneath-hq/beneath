package codec

import (
	"fmt"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/pkg/schemalang"
	"github.com/beneath-hq/beneath/pkg/schemalang/avro"
)

// "Avro native" is our term for the map-based representation defined in "JSON Encoding" in the Avro spec.
// It only differs from the intuitive map-based representation for unions, which require an extra level of
// nesting denoting the specific union type.
//
// So, the rules are:
//   The Nullable case is the only one that has a different representation in Avro native.
//   The Array, Record and Ref cases need to recurse on their children values in case there's a Nullable on that path.
//   The remaining cases return the value outright.
//
// NOTE: These functions do NOT check if val conforms to the schema, except where necessary to continue processing.
//       If val does not conform to the schema, this will surface when (un)marshalling to(/from) Avro.

func (c *Codec) convertFromAvroNative(val interface{}) (interface{}, error) {
	return avroNativeConverter{Refs: c.getRefs()}.convert(c.Schema, val, true)
}

func (c *Codec) convertToAvroNative(val interface{}) (interface{}, error) {
	return avroNativeConverter{Refs: c.getRefs()}.convert(c.Schema, val, false)
}

type avroNativeConverter struct {
	Refs map[string]schemalang.Schema
}

func (c avroNativeConverter) convert(schema schemalang.Schema, val interface{}, fromAvroNative bool) (interface{}, error) {
	if val == nil {
		return val, nil
	}
	switch s := schema.(type) {
	case *schemalang.Nullable:
		// this is the case we need to convert
		if fromAvroNative {
			return c.convertNullableFromAvroNative(s, val)
		}
		return c.convertNullableToAvroNative(s, val)
	case *schemalang.Array:
		// this case we just need to recurse on
		arr, ok := val.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array value for schema <array>, got %v", val)
		}
		res := make([]interface{}, len(arr))
		for i, v := range arr {
			item, err := c.convert(s.ItemType, v, fromAvroNative)
			if err != nil {
				return nil, err
			}
			res[i] = item
		}
		return res, nil
	case *schemalang.Record:
		// this case we just need to recurse on
		record, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected map value for schema <record>, got %v", val)
		}
		res := make(map[string]interface{}, len(record))
		for _, field := range s.Fields {
			fieldVal, err := c.convert(field.Type, record[field.Name], fromAvroNative)
			if err != nil {
				return nil, err
			}
			res[field.Name] = fieldVal
		}
		return res, nil
	case *schemalang.Ref:
		// this case we just need to recurse on
		return c.convert(c.Refs[s.Name], val, fromAvroNative)
	case *schemalang.Primitive:
		if fromAvroNative {
			return c.convertPrimitiveFromAvroNative(s, val)
		}
		return c.convertPrimitiveToAvroNative(s, val)
	case *schemalang.Fixed:
		return val, nil
	case *schemalang.Enum:
		return val, nil
	default:
		panic(fmt.Errorf("unexpected Avro schema type %T", s))
	}
}

func (c avroNativeConverter) convertNullableFromAvroNative(schema *schemalang.Nullable, val interface{}) (interface{}, error) {
	// convert(...) ensures val is not nil
	valMap, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected map type for nullable schema, got %T", schema)
	}
	name := c.unionNameForNullable(schema)
	if valMap[name] == nil {
		return nil, fmt.Errorf("expected key '%s' in native Avro union value '%v'", name, valMap)
	}
	return c.convert(schema.NonNullType, valMap[name], true)
}

func (c avroNativeConverter) convertNullableToAvroNative(schema *schemalang.Nullable, val interface{}) (interface{}, error) {
	// convert(...) ensures val is not nil
	childVal, err := c.convert(schema.NonNullType, val, false)
	if err != nil {
		return nil, err
	}
	name := c.unionNameForNullable(schema)
	return map[string]interface{}{name: childVal}, nil
}

func (c avroNativeConverter) unionNameForNullable(schema *schemalang.Nullable) string {
	var name string
	switch nonNullSchema := schema.NonNullType.(type) {
	case *schemalang.Primitive:
		// name should be "primitive[.logicalType]"
		switch nonNullSchema.Type {
		case schemalang.StringType:
			name = string(avro.StringType)
		case schemalang.BytesType:
			name = string(avro.BytesType)
		case schemalang.IntType:
			name = string(avro.IntType)
		case schemalang.LongType:
			name = string(avro.LongType)
		case schemalang.FloatType:
			name = string(avro.FloatType)
		case schemalang.DoubleType:
			name = string(avro.DoubleType)
		case schemalang.BooleanType:
			name = string(avro.BooleanType)
		default:
			panic(fmt.Errorf("unexpected type '%s'", nonNullSchema.Type))
		}
		if nonNullSchema.LogicalType != "" {
			name += "."
			switch nonNullSchema.LogicalType {
			case schemalang.NumericLogicalType:
				name += string(avro.DecimalLogicalType)
			case schemalang.UUIDLogicalType:
				name += string(avro.UUIDLogicalType)
			case schemalang.TimestampMillisLogicalType:
				name += string(avro.TimestampMillisLogicalType)
			default:
				panic(fmt.Errorf("unexpected logical type '%s'", nonNullSchema.LogicalType))
			}
		}
	case *schemalang.Array:
		name = "array"
	case *schemalang.Fixed:
		name = fmt.Sprintf("bytes%d", nonNullSchema.Size)
	case *schemalang.Enum:
		name = nonNullSchema.Name
	case *schemalang.Record:
		name = nonNullSchema.Name
	case *schemalang.Ref:
		name = nonNullSchema.Name
	default:
		// not expecting another Nullable or RecordField
		panic(fmt.Errorf("unexpected underlying type for nullable: %T", nonNullSchema))
	}
	return name
}

func (c avroNativeConverter) convertPrimitiveFromAvroNative(schema *schemalang.Primitive, val interface{}) (interface{}, error) {
	// the avro library uses strings for UUIDs; we convert to type github.com/satori/go.uuid
	if schema.LogicalType == schemalang.UUIDLogicalType {
		if val, ok := val.(string); ok {
			return uuid.FromString(val)
		}
	}

	// fallback for all other types
	return val, nil
}

func (c avroNativeConverter) convertPrimitiveToAvroNative(schema *schemalang.Primitive, val interface{}) (interface{}, error) {
	// the avro library expects strings for UUID; we convert from type github.com/satori/go.uuid
	if schema.LogicalType == schemalang.UUIDLogicalType {
		if val, ok := val.(uuid.UUID); ok {
			return val.String(), nil
		}
	}

	// fallback for all other types
	return val, nil
}
