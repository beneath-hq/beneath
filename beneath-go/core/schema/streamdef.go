package schema

// StreamDef has data about a stream defined in a schema
type StreamDef struct {
	Name        string
	Description string
	TypeName    string
	KeyFields   []string
	Compiler    *Compiler
}
