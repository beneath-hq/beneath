package transpilers

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/beneath-hq/beneath/pkg/schemalang"
)

// ToBigQuery transpiles an Avro schema to a BigQuery schema
func ToBigQuery(s schemalang.Schema, doc bool) bigquery.Schema {
	t := &toBigQuery{
		Doc:  doc,
		Refs: make(map[string]*bigquery.FieldSchema),
	}
	root := t.Transpile(s)
	return root.Schema
}

type toBigQuery struct {
	Doc  bool
	Refs map[string]*bigquery.FieldSchema
}

func (t *toBigQuery) Transpile(s schemalang.Schema) *bigquery.FieldSchema {
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
	case *schemalang.Ref:
		return t.fromRef(sT)
	default:
		panic(fmt.Errorf("unexpected Avro schema type %T", s))
	}
}

func (t *toBigQuery) fromPrimitive(s *schemalang.Primitive) *bigquery.FieldSchema {
	field := &bigquery.FieldSchema{
		Required: true,
	}

	if s.LogicalType == "" {
		switch s.Type {
		case schemalang.BooleanType:
			field.Type = bigquery.BooleanFieldType
		case schemalang.IntType:
			field.Type = bigquery.IntegerFieldType
		case schemalang.LongType:
			field.Type = bigquery.IntegerFieldType
		case schemalang.FloatType:
			field.Type = bigquery.FloatFieldType
		case schemalang.DoubleType:
			field.Type = bigquery.FloatFieldType
		case schemalang.BytesType:
			field.Type = bigquery.BytesFieldType
		case schemalang.StringType:
			field.Type = bigquery.StringFieldType
		default:
			panic(fmt.Errorf("unexpected primitive type '%s'", s.LogicalType))
		}
	} else {
		switch s.LogicalType {
		case schemalang.NumericLogicalType:
			field.Type = bigquery.NumericFieldType
		case schemalang.UUIDLogicalType:
			field.Type = bigquery.StringFieldType
		case schemalang.TimestampMillisLogicalType:
			field.Type = bigquery.TimestampFieldType
		default:
			panic(fmt.Errorf("unexpected logical type '%s'", s.LogicalType))
		}
	}

	return field
}

func (t *toBigQuery) fromArray(s *schemalang.Array) *bigquery.FieldSchema {
	field := t.Transpile(s.ItemType)
	field.Repeated = true
	return field
}

func (t *toBigQuery) fromUnion(s *schemalang.Nullable) *bigquery.FieldSchema {
	field := t.Transpile(s.NonNullType)
	field.Required = false
	return field
}

func (t *toBigQuery) fromFixed(s *schemalang.Fixed) *bigquery.FieldSchema {
	return &bigquery.FieldSchema{
		Type:     bigquery.BytesFieldType,
		Required: true,
	}
}

func (t *toBigQuery) fromEnum(s *schemalang.Enum) *bigquery.FieldSchema {
	enum := &bigquery.FieldSchema{
		Type:     bigquery.StringFieldType,
		Required: true,
	}
	t.Refs[s.Name] = enum
	return enum
}

func (t *toBigQuery) fromRecord(s *schemalang.Record) *bigquery.FieldSchema {
	fields := make([]*bigquery.FieldSchema, len(s.Fields))
	for idx, avroField := range s.Fields {
		field := t.Transpile(avroField.Type)
		field.Name = avroField.Name
		if t.Doc {
			field.Description = avroField.Doc
		}
		fields[idx] = field
	}

	record := &bigquery.FieldSchema{
		Name:     s.Name,
		Required: true,
		Type:     bigquery.RecordFieldType,
		Schema:   fields,
	}
	if t.Doc {
		record.Description = s.Doc
	}

	t.Refs[s.Name] = record

	return record
}

func (t *toBigQuery) fromRef(s *schemalang.Ref) *bigquery.FieldSchema {
	ref := t.Refs[s.Name]
	return &bigquery.FieldSchema{
		Required: ref.Required,
		Schema:   ref.Schema,
		Type:     ref.Type,
	}
}
