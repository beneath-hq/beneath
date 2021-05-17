package transpilers

import (
	"fmt"
	"strconv"

	"github.com/beneath-hq/beneath/pkg/schemalang"
	"github.com/beneath-hq/beneath/pkg/schemalang/avro"
)

// ToAvro converts a Schema to the Avro (JSON) representation.
// If doc is false, the result may be considered a canonical schema.
func ToAvro(s schemalang.Schema, doc bool) string {
	t := &toAvro{
		Doc:       doc,
		NamesSeen: make(map[string]bool),
	}
	return t.Transpile(s)
}

type toAvro struct {
	Doc       bool
	NamesSeen map[string]bool
}

func (t *toAvro) Transpile(s schemalang.Schema) string {
	// Field order is name, doc, type, logicalType, ...remaining in alphabetized order
	switch sT := s.(type) {
	case *schemalang.Primitive:
		return t.fromPrimitive(sT)
	case *schemalang.Array:
		return t.fromArray(sT)
	case *schemalang.Nullable:
		return t.fromUnion(sT)
	case *schemalang.Fixed:
		return t.fromFixed(sT)
	case *schemalang.Enum:
		return t.fromEnum(sT)
	case *schemalang.Record:
		return t.fromRecord(sT)
	case *schemalang.RecordField:
		return t.fromRecordField(sT)
	case *schemalang.Ref:
		return t.fromRef(sT)
	default:
		panic(fmt.Errorf("unexpected Avro schema type %T", s))
	}
}

func (t *toAvro) fromPrimitive(s *schemalang.Primitive) string {
	var typeName avro.Type
	switch s.Type {
	case schemalang.StringType:
		typeName = avro.StringType
	case schemalang.BytesType:
		typeName = avro.BytesType
	case schemalang.IntType:
		typeName = avro.IntType
	case schemalang.LongType:
		typeName = avro.LongType
	case schemalang.FloatType:
		typeName = avro.FloatType
	case schemalang.DoubleType:
		typeName = avro.DoubleType
	case schemalang.BooleanType:
		typeName = avro.BooleanType
	default:
		panic(fmt.Errorf("unexpected type '%T'", s.Type))
	}

	if s.LogicalType == "" {
		return `"` + string(typeName) + `"`
	}

	var logicalTypeName avro.LogicalType

	switch s.LogicalType {
	case schemalang.NumericLogicalType:
		logicalTypeName = avro.DecimalLogicalType
	case schemalang.UUIDLogicalType:
		logicalTypeName = avro.UUIDLogicalType
	case schemalang.TimestampMillisLogicalType:
		logicalTypeName = avro.TimestampMillisLogicalType
	}

	inject := `,"logicalType":"` + string(logicalTypeName) + `"`
	if s.LogicalType == schemalang.NumericLogicalType {
		inject += `,"precision":128,"scale":0`
	}
	return `{"type":"` + string(typeName) + `"` + inject + `}`
}

func (t *toAvro) fromArray(s *schemalang.Array) string {
	return `{"type":"array","items":` + t.Transpile(s.ItemType) + `}`
}

func (t *toAvro) fromUnion(s *schemalang.Nullable) string {
	return `["null",` + t.Transpile(s.NonNullType) + `]`
}

func (t *toAvro) fromFixed(s *schemalang.Fixed) string {
	name := fmt.Sprintf("bytes%d", s.Size)
	if t.NamesSeen[name] {
		return `"` + name + `"`
	}
	t.NamesSeen[name] = true
	size := strconv.Itoa(s.Size)
	return `{"name":"` + name + `","type":"fixed","size":` + size + `}`
}

func (t *toAvro) fromEnum(s *schemalang.Enum) string {
	if t.NamesSeen[s.Name] {
		return `"` + s.Name + `"`
	}
	t.NamesSeen[s.Name] = true
	symbols := ""
	for _, sym := range s.Symbols {
		symbols += `"` + sym + `",`
	}
	if len(symbols) > 0 {
		symbols = symbols[:len(symbols)-1] // remove last comma
	}
	start := `{"name":"` + s.Name
	if t.Doc && s.Doc != "" {
		start += `","doc":"` + s.Doc
	}
	return start + `","type":"enum","symbols":[` + symbols + `]}`
}

func (t *toAvro) fromRecord(s *schemalang.Record) string {
	if t.NamesSeen[s.Name] {
		return `"` + s.Name + `"`
	}
	t.NamesSeen[s.Name] = true
	fields := ""
	for _, f := range s.Fields {
		fields += t.Transpile(f) + ","
	}
	if len(fields) > 0 {
		fields = fields[:len(fields)-1] // remove last comma
	}
	start := `{"name":"` + s.Name
	if t.Doc && s.Doc != "" {
		start += `","doc":"` + s.Doc
	}
	return start + `","type":"record","fields":[` + fields + `]}`
}

func (t *toAvro) fromRecordField(s *schemalang.RecordField) string {
	start := `{"name":"` + s.Name
	if t.Doc && s.Doc != "" {
		start += `","doc":"` + s.Doc
	}
	return start + `","type":` + t.Transpile(s.Type) + `}`
}

func (t *toAvro) fromRef(s *schemalang.Ref) string {
	return `"` + s.Name + `"`
}
