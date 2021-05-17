package transpilers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beneath-hq/beneath/pkg/schemalang"
	"github.com/beneath-hq/beneath/pkg/schemalang/avro"
)

// FromAvro converts a json-encoded Avro schema to the conventional
// in-memory representation of an Avro schema.
func FromAvro(jsonStr string) (schemalang.Schema, error) {
	var valT interface{}
	dec := json.NewDecoder(strings.NewReader(jsonStr))
	dec.UseNumber()
	err := dec.Decode(&valT)
	if err != nil {
		return nil, err
	}
	t := fromAvro{
		FixedRefs: make(map[string]*schemalang.Fixed),
	}
	return t.Transpile(valT)
}

type fromAvro struct {
	FixedRefs map[string]*schemalang.Fixed
}

func (t *fromAvro) Transpile(valT interface{}) (schemalang.Schema, error) {
	switch val := valT.(type) {
	case string:
		return t.fromString(val)
	case []interface{}:
		return t.fromSlice(val)
	case map[string]interface{}:
		return t.fromMap(val)
	default:
		return nil, fmt.Errorf("unexpected value in schema: '%v'", val)
	}
}

func (t *fromAvro) fromString(val string) (schemalang.Schema, error) {
	switch avro.Type(val) {
	case avro.NullType:
		return nil, fmt.Errorf("found unexpected 'null' primitive (only valid as the first option in a union)")
	case avro.StringType:
		return &schemalang.Primitive{Type: schemalang.StringType}, nil
	case avro.BytesType:
		return &schemalang.Primitive{Type: schemalang.BytesType}, nil
	case avro.IntType:
		return &schemalang.Primitive{Type: schemalang.IntType}, nil
	case avro.LongType:
		return &schemalang.Primitive{Type: schemalang.LongType}, nil
	case avro.FloatType:
		return &schemalang.Primitive{Type: schemalang.FloatType}, nil
	case avro.DoubleType:
		return &schemalang.Primitive{Type: schemalang.DoubleType}, nil
	case avro.BooleanType:
		return &schemalang.Primitive{Type: schemalang.BooleanType}, nil
	default:
		if t.FixedRefs[val] != nil {
			return t.FixedRefs[val], nil
		}
		return &schemalang.Ref{Name: val}, nil
	}
}

func (t *fromAvro) fromSlice(val []interface{}) (schemalang.Schema, error) {
	if len(val) == 2 {
		first, ok := val[0].(string)
		if !ok || avro.Type(first) != avro.NullType {
			return nil, fmt.Errorf("expected first type in union to be '%s', got: '%v'", avro.NullType, val[0])
		}

		second, err := t.Transpile(val[1])
		if err != nil {
			return nil, err
		}
		if second.GetType() == schemalang.NullableType {
			return nil, fmt.Errorf("the second type in a union must not be another union")
		}

		return &schemalang.Nullable{NonNullType: second}, nil
	}
	return nil, fmt.Errorf("unions must have exactly two types and the first type must be 'null'")
}

func (t *fromAvro) fromMap(val map[string]interface{}) (schemalang.Schema, error) {
	typeStr, ok := val["type"].(string)
	if !ok {
		return t.Transpile(val["type"])
	}

	switch avro.Type(typeStr) {
	case avro.ArrayType:
		return t.fromMapArray(val)
	case avro.FixedType:
		return t.fromMapFixed(val)
	case avro.EnumType:
		return t.fromMapEnum(val)
	case avro.RecordType:
		return t.fromMapRecord(val)
	default:
		// it's a primitive (maybe with a logical type) or ref, or ref resolved to a fixed
		transpiledT, err := t.Transpile(typeStr)
		if err != nil {
			return nil, err
		}
		switch transpiled := transpiledT.(type) {
		case *schemalang.Primitive:
			return t.primitiveWithLogicalType(transpiled, val)
		case *schemalang.Ref:
			return transpiled, nil
		case *schemalang.Fixed:
			return transpiled, nil
		default:
			panic(fmt.Errorf("unexpected transpiled type '%T'", transpiled))
		}
	}
}

func (t *fromAvro) fromMapArray(val map[string]interface{}) (schemalang.Schema, error) {
	if val["items"] == nil {
		return nil, fmt.Errorf("expected field 'items' in schema with type 'array'")
	}
	itemType, err := t.Transpile(val["items"])
	if err != nil {
		return nil, err
	}
	return &schemalang.Array{
		ItemType: itemType,
	}, nil
}

func (t *fromAvro) fromMapFixed(val map[string]interface{}) (schemalang.Schema, error) {
	name, ok := val["name"].(string)
	if !ok || len(name) == 0 {
		return nil, fmt.Errorf("expected non-empty string field 'name' for type 'fixed'")
	}
	sizeN, ok := val["size"].(json.Number)
	if !ok {
		return nil, fmt.Errorf("expected non-empty integer field 'size' for type 'fixed'")
	}
	size, err := sizeN.Int64()
	if err != nil {
		return nil, fmt.Errorf("expected integer 'size' for type 'fixed', got float: '%s'", sizeN.String())
	}
	// check name has format "bytesN" (only one we support)
	if len(name) < 6 || name[0:5] != "bytes" || name[5:len(name)] != sizeN.String() {
		return nil, fmt.Errorf("the name of 'fixed' type with size %d must be 'bytes%d', got name '%s' (this convention helps portability with other schema types)", size, size, name)
	}
	if t.FixedRefs[name] != nil {
		return nil, fmt.Errorf("found multiple declarations of 'fixed' type '%s'", name)
	}
	fixed := &schemalang.Fixed{
		Size: int(size),
	}
	t.FixedRefs[name] = fixed
	return fixed, nil
}

func (t *fromAvro) fromMapEnum(val map[string]interface{}) (schemalang.Schema, error) {
	name, ok := val["name"].(string)
	if !ok || len(name) == 0 {
		return nil, fmt.Errorf("expected non-empty string field 'name' for type 'enum'")
	}
	doc, ok := val["doc"].(string)
	if !ok {
		doc = ""
	}
	symbols, ok := val["symbols"].([]interface{})
	if !ok || len(symbols) == 0 {
		return nil, fmt.Errorf("expected non-empty array-of-strings field 'symbols' for type 'enum'")
	}
	symbolsStrings := make([]string, len(symbols))
	for idx, val := range symbols {
		symbol, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("expected string symbol for enum, got: %v", symbol)
		}
		symbolsStrings[idx] = symbol
	}
	enum := &schemalang.Enum{
		Name:    name,
		Doc:     doc,
		Symbols: symbolsStrings,
	}
	return enum, nil
}

func (t *fromAvro) fromMapRecord(val map[string]interface{}) (schemalang.Schema, error) {
	name, ok := val["name"].(string)
	if !ok || len(name) == 0 {
		return nil, fmt.Errorf("expected non-empty string field 'name' for type 'record'")
	}
	doc, ok := val["doc"].(string)
	if !ok {
		doc = ""
	}
	fieldVals, ok := val["fields"].([]interface{})
	if !ok || len(fieldVals) == 0 {
		return nil, fmt.Errorf("expected non-empty array field 'fields' for type 'record'")
	}
	record := &schemalang.Record{
		Name: name,
		Doc:  doc,
	}
	fields := make([]*schemalang.RecordField, len(fieldVals))
	for idx, fieldValT := range fieldVals {
		fieldVal, ok := fieldValT.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected dict for record field, got '%v'", fieldVal)
		}
		fieldName, ok := fieldVal["name"].(string)
		if !ok || len(fieldName) == 0 {
			return nil, fmt.Errorf("expected non-empty string field 'name' for all record fields")
		}
		fieldDoc, ok := fieldVal["doc"].(string)
		if !ok {
			fieldDoc = ""
		}
		fieldType, err := t.Transpile(fieldVal)
		if err != nil {
			return nil, err
		}
		fields[idx] = &schemalang.RecordField{
			Name: fieldName,
			Doc:  fieldDoc,
			Type: fieldType,
		}
	}
	record.Fields = fields
	return record, nil
}

func (t *fromAvro) primitiveWithLogicalType(primitive *schemalang.Primitive, val map[string]interface{}) (*schemalang.Primitive, error) {
	if val["logicalType"] == nil {
		return primitive, nil
	}

	logicalType, ok := val["logicalType"].(string)
	if !ok {
		return nil, fmt.Errorf("unexpected logicalType: '%v'", val["logicalType"])
	}

	switch avro.LogicalType(logicalType) {
	case avro.DecimalLogicalType:
		// maybe also check precision is at most 128?
		if scaleT, ok := val["scale"]; ok {
			scale, ok := scaleT.(json.Number)
			if !ok {
				return nil, fmt.Errorf("unexpected type for 'scale' on 'decimal' logical type: %T", scale)
			}
			if scale.String() != "0" {
				return nil, fmt.Errorf("property 'scale' on 'decimal' logical type must be 0, got: <%s>", scale.String())
			}
		}
		if primitive.GetType() == schemalang.BytesType {
			primitive.LogicalType = schemalang.NumericLogicalType
			return primitive, nil
		}
	case avro.UUIDLogicalType:
		if primitive.GetType() == schemalang.StringType {
			primitive.LogicalType = schemalang.UUIDLogicalType
			return primitive, nil
		}
	case avro.TimestampMillisLogicalType:
		if primitive.GetType() == schemalang.LongType {
			primitive.LogicalType = schemalang.TimestampMillisLogicalType
			return primitive, nil
		}
	}

	return nil, fmt.Errorf("logical type '%s' not supported for primitive type '%s'", logicalType, primitive.GetType())
}
