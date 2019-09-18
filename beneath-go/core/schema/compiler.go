package schema

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/golang-collections/collections/set"
)

// Compiler represents a single compilation input
type Compiler struct {
	Input        string
	AST          *File
	Declarations map[string]*Declaration
	Streams      map[string]*StreamDef
}

//
const (
	maxFieldNameLen = 64
)

// MustCompileToAvro is a helper function very useful in tests.
// It compiles the schema to avro and panics on any errors.
func MustCompileToAvro(schema string) interface{} {
	avroSchemaString := MustCompileToAvroString(schema)

	var avroSchema interface{}
	err := json.Unmarshal([]byte(avroSchemaString), &avroSchema)
	if err != nil {
		panic(err)
	}

	return avroSchema
}

// MustCompileToAvroString is a variant of MustCompileToAvro
func MustCompileToAvroString(schema string) string {
	c := NewCompiler(schema)
	err := c.Compile()
	if err != nil {
		panic(err)
	}

	avroSchemaString, err := c.GetStream().BuildCanonicalAvroSchema()
	if err != nil {
		panic(err)
	}

	return avroSchemaString
}

// NewCompiler creates a new Compiler for schema -- don't forget to call Compile()
func NewCompiler(schema string) *Compiler {
	return &Compiler{
		Input:        schema,
		AST:          &File{},
		Declarations: make(map[string]*Declaration),
		Streams:      make(map[string]*StreamDef),
	}
}

// GetStream returns the stream found during compilation (must call Compile() first)
// (will refactor when adding support for multiple stream definitions in one file)
func (c *Compiler) GetStream() *StreamDef {
	for _, s := range c.Streams {
		return s
	}
	return nil
}

// Compile parses the schema (given in NewCompiler)
func (c *Compiler) Compile() error {
	// parse ast
	err := GetParser().ParseString(c.Input, c.AST)
	if err != nil {
		return fmt.Errorf("parse error: %v", err.Error())
	}

	// extract type declarations
	for _, declaration := range c.AST.Declarations {
		// get name
		var name string
		if declaration.Enum != nil {
			name = declaration.Enum.Name
		} else if declaration.Type != nil {
			name = declaration.Type.Name
		}

		// check title starts with uppercase
		if !unicode.IsUpper(rune(name[0])) {
			return fmt.Errorf("type name '%v' should start with an uppercase letter", name)
		}

		// check it's not a primitive type name
		if isPrimitiveTypeName(name) {
			return fmt.Errorf("declaration of '%v' at %v overlaps with primitive type name", name, declaration.Pos.String())
		}

		// check doesn't exist
		if c.Declarations[name] != nil {
			return fmt.Errorf("name '%v' at %v has already been declared", name, declaration.Pos.String())
		}

		// save declaration
		c.Declarations[name] = declaration

		// parse stream and save
		s, err := c.parseStream(declaration)
		if s != nil {
			c.Streams[s.TypeName] = s
		}
		if err != nil {
			return err
		}
	}

	// check there's at least one stream
	if len(c.Streams) == 0 {
		return fmt.Errorf("no streams declared in input")
	}

	// check no stream names are declared twice
	streamNamesSeen := make(map[string]bool, len(c.Streams))
	for _, s := range c.Streams {
		if streamNamesSeen[s.Name] {
			return fmt.Errorf("stream name '%v' used twice", s.Name)
		}
		streamNamesSeen[s.Name] = true
	}

	// (TEMPORARY) require that there's only one stream in the input
	if len(c.Streams) != 1 {
		return fmt.Errorf("more than one schema declared in input")
	}

	// type check declarations
	seen := make(map[string]bool)
	path := set.New()
	for _, declaration := range c.AST.Declarations { // not c.Declarations to ensure deterministic order
		err := c.checkDeclaration(declaration, seen, path)
		if err != nil {
			return err
		}
	}

	// check field names of types and enums
	for _, declaration := range c.Declarations {
		// check enum names not declared twice
		if declaration.Enum != nil {
			members := make(map[string]bool)
			for _, name := range declaration.Enum.Members {
				if members[name] {
					return fmt.Errorf("member '%v' declared twice in enum '%v'", name, declaration.Enum.Name)
				}
				members[name] = true
			}
		}

		// check type field names and stream keys
		if declaration.Type != nil {
			// check not declared twice
			fields := make(map[string]*TypeRef, len(declaration.Type.Fields))
			for _, field := range declaration.Type.Fields {
				if fields[field.Name] != nil {
					return fmt.Errorf("field '%v' declared twice in type '%v'", field.Name, declaration.Type.Name)
				}
				fields[field.Name] = field.Type
			}

			// if it's a stream, check key fields a) exist, b) not used twice, c) are primitive, d) not optional
			stream := c.Streams[declaration.Type.Name]
			if stream != nil {
				keysSeen := make(map[string]bool, len(stream.KeyFields))
				for _, key := range stream.KeyFields {
					// check it exists
					if fields[key] == nil {
						return fmt.Errorf("field '%v' in key doesn't exist in type '%v'", key, declaration.Type.Name)
					}

					// check not used twice
					if keysSeen[key] {
						return fmt.Errorf("field '%v' used twice in key for type '%v'", key, declaration.Type.Name)
					}
					keysSeen[key] = true

					// check it has a primitive type
					if !isIndexableType(fields[key]) {
						return fmt.Errorf("field '%v' in type '%v' cannot be used as key", key, declaration.Type.Name)
					}

					// check not optional
					if !fields[key].Required {
						return fmt.Errorf("field '%v' in type '%v' cannot be used as key because it is optional", key, declaration.Type.Name)
					}
				}
			}
		}
	}

	// done
	return nil
}

// type checks declaration
func (c *Compiler) checkDeclaration(d *Declaration, seen map[string]bool, path *set.Set) error {
	// separate check for enum
	if d.Enum != nil {
		// check has more than one member
		if len(d.Enum.Members) == 0 {
			return fmt.Errorf("enum '%v' must have at least one member", d.Enum.Name)
		}

		// done for enums
		return nil
	}

	// proceeding means d.Type != nil

	// check circular type
	if path.Has(d.Type.Name) {
		return fmt.Errorf("type '%v' is circular, which is not supported", d.Type.Name)
	}

	// check seen
	if seen[d.Type.Name] {
		return nil
	}
	seen[d.Type.Name] = true

	// check fields not empty
	if len(d.Type.Fields) == 0 {
		return fmt.Errorf("type '%v' does not define any fields", d.Type.Name)
	}

	// check types
	path.Insert(d.Type.Name)
	for _, field := range d.Type.Fields {
		e := c.checkFieldName(field.Name)
		if e != nil {
			return e
		}

		e = c.checkTypeRef(field.Type, seen, path)
		if e != nil {
			return e
		}
	}
	path.Remove(d.Type.Name)

	// done
	return nil
}

// type checks typeRef
func (c *Compiler) checkTypeRef(tr *TypeRef, seen map[string]bool, path *set.Set) error {
	// if array, check for nested arrays and double-optional, then recurse
	if tr.Array != nil {
		if tr.Array.Array != nil {
			return fmt.Errorf("nested lists are not allowed at %v", tr.Pos.String())
		}
		if !tr.Array.Required {
			return fmt.Errorf("type wrapped by list cannot be optional at %v", tr.Pos.String())
		}
		return c.checkTypeRef(tr.Array, seen, path)
	}

	// if primitive, approve
	if isPrimitiveType(tr) {
		return nil
	}

	// if not custom declared type, error
	if c.Declarations[tr.Type] == nil {
		return fmt.Errorf("unknown type '%v' at %v", tr.Type, tr.Pos.String())
	}

	// recurse on type
	return c.checkDeclaration(c.Declarations[tr.Type], seen, path)
}

// checks if field name is allowed
func (c *Compiler) checkFieldName(name string) error {
	if len(name) > maxFieldNameLen {
		return fmt.Errorf("field name '%v' exceeds limit of 127 characters", name)
	}

	if name == "__key" || name == "__timestamp" {
		return fmt.Errorf("field name '%v' is a reserved identifier", name)
	}

	return nil
}

// parse declaration as stream if it has a stream annotation
func (c *Compiler) parseStream(declaration *Declaration) (*StreamDef, error) {
	// if not a stream, return empty
	if declaration.Type == nil || len(declaration.Type.Annotations) == 0 {
		return nil, nil
	}

	// check there's only one annotation
	annotations := declaration.Type.Annotations
	if len(annotations) > 1 {
		return nil, fmt.Errorf("multiple annotations in declaration at %v", declaration.Pos.String())
	}

	// check it's a stream annotation
	a := annotations[0]
	if a.Name != "stream" {
		return nil, fmt.Errorf("annotation @%v not supported", a.Name)
	}

	// read params
	var streamName string
	var streamKey []string
	var streamExternal bool
	for _, arg := range a.Arguments {
		switch arg.Name {
		case "name":
			if arg.Value.String == "" {
				return nil, fmt.Errorf("stream arg 'name' at %v is not a string", arg.Pos.String())
			}
			streamName = arg.Value.String
		case "key":
			err := fmt.Errorf("stream arg 'key' at %v is not a string or array of strings", arg.Pos.String())
			if arg.Value.String != "" {
				streamKey = []string{arg.Value.String}
			} else if arg.Value.Array != nil && len(arg.Value.Array) > 0 {
				streamKey = make([]string, len(arg.Value.Array))
				for idx, val := range arg.Value.Array {
					if val.String == "" {
						return nil, err
					}
					streamKey[idx] = val.String
				}
			} else {
				return nil, err
			}
		case "external":
			streamExternal = (arg.Value.Symbol == "true")
		default:
			return nil, fmt.Errorf("unknown @stream arg '%v' at %v", arg.Name, arg.Pos.String())
		}
	}

	// check params
	if len(streamName) == 0 {
		return nil, fmt.Errorf("missing arg 'name' in @stream at %v", declaration.Pos.String())
	}
	if len(streamKey) == 0 {
		return nil, fmt.Errorf("missing arg 'key' in @stream at %v", declaration.Pos.String())
	}

	// done
	return &StreamDef{
		Name:        streamName,
		Description: strings.TrimSpace(declaration.Type.Doc),
		TypeName:    declaration.Type.Name,
		KeyFields:   streamKey,
		External:    streamExternal,
		Compiler:    c,
	}, nil
}
