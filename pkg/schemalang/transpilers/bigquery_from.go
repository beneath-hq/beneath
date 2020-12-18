package transpilers

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"gitlab.com/beneath-hq/beneath/pkg/schemalang"
)

// FromBigQuery transpiles a BigQuery schema to an Avro schema
func FromBigQuery(schema bigquery.Schema) (schemalang.Schema, error) {
	if len(schema) == 0 {
		return nil, nil
	}
	return fromBigQuery{}.fromSchema(schema, "schema")
}

type fromBigQuery struct{}

func (t fromBigQuery) fromSchema(bq bigquery.Schema, path string) (schemalang.Schema, error) {
	fields := make([]*schemalang.RecordField, len(bq))
	for idx, field := range bq {
		conv, err := t.fromField(field, path)
		if err != nil {
			return nil, err
		}
		fields[idx] = conv
	}
	return &schemalang.Record{
		Name:   path,
		Fields: fields,
	}, nil
}

func (t fromBigQuery) fromField(field *bigquery.FieldSchema, path string) (*schemalang.RecordField, error) {
	var fieldSchema schemalang.Schema

	switch field.Type {
	case bigquery.StringFieldType:
		fieldSchema = &schemalang.Primitive{Type: schemalang.StringType}
	case bigquery.BytesFieldType:
		fieldSchema = &schemalang.Primitive{Type: schemalang.BytesType}
	case bigquery.IntegerFieldType:
		fieldSchema = &schemalang.Primitive{Type: schemalang.LongType}
	case bigquery.FloatFieldType:
		fieldSchema = &schemalang.Primitive{Type: schemalang.DoubleType}
	case bigquery.BooleanFieldType:
		fieldSchema = &schemalang.Primitive{Type: schemalang.BooleanType}
	case bigquery.TimestampFieldType:
		fieldSchema = &schemalang.Primitive{
			Type:        schemalang.LongType,
			LogicalType: schemalang.TimestampMillisLogicalType,
		}
	case bigquery.RecordFieldType:
		conv, err := t.fromSchema(field.Schema, fmt.Sprintf("%s_%s", path, field.Name))
		if err != nil {
			return nil, err
		}
		fieldSchema = conv
	case bigquery.DateTimeFieldType:
		return nil, fmt.Errorf("Use of data type DATETIME for field '%s' not supported; use TIMESTAMP instead", field.Name)
	case bigquery.DateFieldType:
		return nil, fmt.Errorf("Use of data type DATE for field '%s' not supported; use TIMESTAMP instead", field.Name)
	case bigquery.TimeFieldType:
		return nil, fmt.Errorf("Use of data type TIME for field '%s' not supported; use TIMESTAMP instead", field.Name)
	case bigquery.NumericFieldType:
		fieldSchema = &schemalang.Primitive{
			Type:        schemalang.BytesType,
			LogicalType: schemalang.NumericLogicalType,
		}
	case bigquery.GeographyFieldType:
		fieldSchema = &schemalang.Primitive{Type: schemalang.StringType}
	default:
		panic(fmt.Errorf("unsupported field type: %v", field.Type))
	}

	if field.Repeated || !field.Required {
		fieldSchema = &schemalang.Nullable{NonNullType: fieldSchema}
	}

	if field.Repeated {
		fieldSchema = &schemalang.Array{ItemType: fieldSchema}
	}

	return &schemalang.RecordField{
		Name: field.Name,
		Doc:  field.Description,
		Type: fieldSchema,
	}, nil
}
