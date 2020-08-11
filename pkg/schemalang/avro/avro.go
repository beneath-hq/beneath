package avro

// Type is an Avro type
type Type string

// LogicalType is an Avro logical type
type LogicalType string

// Constants for Type
const (
	StringType  Type = "string"
	BytesType   Type = "bytes"
	IntType     Type = "int"
	LongType    Type = "long"
	FloatType   Type = "float"
	DoubleType  Type = "double"
	BooleanType Type = "boolean"
	NullType    Type = "null"

	FixedType  Type = "fixed"
	ArrayType  Type = "array"
	RecordType Type = "record"
	EnumType   Type = "enum"

	DecimalLogicalType         LogicalType = "decimal"
	TimestampMillisLogicalType LogicalType = "timestamp-millis"
	UUIDLogicalType            LogicalType = "uuid"
)
