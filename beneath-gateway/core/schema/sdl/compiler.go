package sdl

import (
	"fmt"
	"unicode"

	"github.com/golang-collections/collections/set"
)

// Compiler represents a single compilation input
type Compiler struct {
	input        string
	ast          *File
	declarations map[string]*Declaration
	streams      map[string]*streamInfo
}

// NewCompiler creates a new Compiler for schema -- don't forget to call Compile()
func NewCompiler(schema string) *Compiler {
	return &Compiler{
		input:        schema,
		ast:          &File{},
		declarations: make(map[string]*Declaration),
		streams:      make(map[string]*streamInfo),
	}
}

type streamInfo struct {
	name        string
	typeName    string
	key         []string
	external    bool
	declaration *Declaration
}

// Compile parses the schema (given in NewCompiler)
func (c *Compiler) Compile() error {
	// parse ast
	err := GetParser().ParseString(c.input, c.ast)
	if err != nil {
		return fmt.Errorf("parse error: %v", err.Error())
	}

	// extract type declarations
	for _, declaration := range c.ast.Declarations {
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

		// check doesn't exist
		if c.declarations[name] != nil {
			return fmt.Errorf("name '%v' at %v has already been declared", name, declaration.Pos.String())
		}

		// save declaration
		c.declarations[name] = declaration

		// parse stream and save
		s, err := c.parseStream(declaration)
		if s != nil {
			c.streams[s.typeName] = s
		}
		if err != nil {
			return err
		}
	}

	// check there's at least one stream
	if len(c.streams) == 0 {
		return fmt.Errorf("no streams declared in input")
	}

	// check no stream names are declared twice
	streamNamesSeen := make(map[string]bool, len(c.streams))
	for _, s := range c.streams {
		if streamNamesSeen[s.name] {
			return fmt.Errorf("stream name '%v' used twice", s.name)
		}
		streamNamesSeen[s.name] = true
	}

	// type check declarations
	seen := make(map[string]bool)
	path := set.New()
	for _, declaration := range c.declarations {
		err := c.checkDeclaration(declaration, seen, path)
		if err != nil {
			return err
		}
	}

	// check field names of types and enums
	for _, declaration := range c.declarations {
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
			stream := c.streams[declaration.Type.Name]
			if stream != nil {
				keysSeen := make(map[string]bool, len(stream.key))
				for _, key := range stream.key {
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
					if !c.isIndexableType(fields[key]) {
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
		e := c.checkTypeRef(field.Type, seen, path)
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
	// if array, recurse
	if tr.Array != nil {
		return c.checkTypeRef(tr.Array, seen, path)
	}

	// if primitive, approve
	if c.isPrimitiveType(tr) {
		return nil
	}

	// if not custom declared type, error
	if c.declarations[tr.Type] == nil {
		return fmt.Errorf("unknown type '%v' at %v", tr.Type, tr.Pos.String())
	}

	// recurse on type
	return c.checkDeclaration(c.declarations[tr.Type], seen, path)
}

// returns true iff primitive
func (c *Compiler) isPrimitiveType(tr *TypeRef) bool {
	if tr.Array != nil {
		return false
	}

	// TODO

	return (tr.Type == "Boolean" ||
		tr.Type == "Int" ||
		tr.Type == "Int32" ||
		tr.Type == "Int64" ||
		tr.Type == "Float" ||
		tr.Type == "Float32" ||
		tr.Type == "Float64" ||
		tr.Type == "Numeric" ||
		tr.Type == "Timestamp" ||
		tr.Type == "Bytes" ||
		tr.Type == "Bytes20" ||
		tr.Type == "String")
}

// returns true iff can be used in an index or as a key
func (c *Compiler) isIndexableType(tr *TypeRef) bool {
	return c.isPrimitiveType(tr) && (tr.Type == "Int" ||
		tr.Type == "Int32" ||
		tr.Type == "Int64" ||
		tr.Type == "Timestamp" ||
		tr.Type == "Bytes" ||
		tr.Type == "String")
}

// parse declaration as stream if it has a stream annotation
func (c *Compiler) parseStream(declaration *Declaration) (*streamInfo, error) {
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
	return &streamInfo{
		name:        streamName,
		typeName:    declaration.Type.Name,
		key:         streamKey,
		external:    streamExternal,
		declaration: declaration,
	}, nil
}
