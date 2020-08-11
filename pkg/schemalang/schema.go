package schemalang

// Schema is our internal representation of a schema.
// It should be one of Primitive, Array, Nullable, Fixed, Enum, Record, RecordField.
// It is largely inspired by Avro (see https://avro.apache.org/docs/current/spec.html).
// Notable differences from Avro: a) unions replaced with Nullable, b) the fixed type is
// unnamed, c) records may specify a key and indexed fields.
type Schema interface {
	GetType() Type
}

// Type is an Avro schema type
type Type string

// Constants for Type
const (
	StringType  Type = "<string>"
	BytesType   Type = "<bytes>"
	IntType     Type = "<int>"
	LongType    Type = "<long>"
	FloatType   Type = "<float>"
	DoubleType  Type = "<double>"
	BooleanType Type = "<boolean>"

	NullableType Type = "<nullable>"
	FixedType    Type = "<fixed>"
	ArrayType    Type = "<array>"
	RecordType   Type = "<record>"
	EnumType     Type = "<enum>"
	RefType      Type = "<ref>"
)

// LogicalType is a supported logical schema type
type LogicalType string

// Constants for LogicalType
const (
	NumericLogicalType         LogicalType = "<numeric>"
	TimestampMillisLogicalType LogicalType = "<timestamp-millis>"
	UUIDLogicalType            LogicalType = "<uuid>"
)

// Primitive represents a primitive or logical Avro type
type Primitive struct {
	Type        Type
	LogicalType LogicalType
}

// GetType implements Schema
func (p *Primitive) GetType() Type {
	return p.Type
}

// Nullable represents an Avro union type of "null" and a sub type
type Nullable struct {
	NonNullType Schema
}

// GetType implements Schema
func (u *Nullable) GetType() Type {
	return NullableType
}

// Fixed represents an Avro fixed type
type Fixed struct {
	Size int
}

// GetType implements Schema
func (f *Fixed) GetType() Type {
	return FixedType
}

// Array represents an Avro array type
type Array struct {
	ItemType Schema
}

// GetType implements Schema
func (a *Array) GetType() Type {
	return ArrayType
}

// Record represents an Avro record type
type Record struct {
	Name   string
	Doc    string
	Fields []*RecordField
}

// GetType implements Schema
func (r *Record) GetType() Type {
	return RecordType
}

// RecordField represents a field in an Avro record type
type RecordField struct {
	Name string
	Doc  string
	Type Schema
}

// GetType implements Schema
func (f *RecordField) GetType() Type {
	return f.Type.GetType()
}

// Enum represents an Avro enum type
type Enum struct {
	Name    string
	Doc     string
	Symbols []string
}

// GetType implements Schema
func (e *Enum) GetType() Type {
	return EnumType
}

// Ref represents a reference to a previously defined named type (record, fixed or enum)
type Ref struct {
	Name string
}

// GetType implements Schema
func (r *Ref) GetType() Type {
	return RefType
}
