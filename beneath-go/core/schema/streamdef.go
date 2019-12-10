package schema

// StreamDef has data about a stream defined in a schema
type StreamDef struct {
	Name             string
	Description      string
	TypeName         string
	KeyIndex         Index
	SecondaryIndexes []Index
	Compiler         *Compiler
}

// Index represents an index on one or more columns
type Index struct {
	Fields      []string
	Denormalize bool
}
