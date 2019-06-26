package sdl

import (
	"fmt"
	"unicode"

	"github.com/golang-collections/collections/set"
)

type Compiler struct {
	input        string
	ast          *File
	declarations map[string]*Declaration
	streams      map[string]*stream
}

func NewCompiler(schema string) Compiler {
	return Compiler{
		input:        schema,
		ast:          &File{},
		declarations: make(map[string]*Declaration),
		streams:      make(map[string]*stream),
	}
}

type stream struct {
	name        string
	key         []string
	declaration *Declaration
}

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
			c.streams[s.name] = s
		}
		if err != nil {
			return err
		}
	}

	// check there's at least one stream
	if len(c.streams) == 0 {
		return fmt.Errorf("no streams declared in input")
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

	// done
	return nil
}

// type checks declaration
func (c *Compiler) checkDeclaration(d *Declaration, seen map[string]bool, path *set.Set) error {
	// no checks for enums
	if d.Type == nil {
		return nil
	}

	// check circular type
	if path.Has(d.Type.Name) {
		return fmt.Errorf("type '%v' is circular, which is not supported", d.Type.Name)
	}

	// check seen
	if seen[d.Type.Name] {
		return nil
	}
	seen[d.Type.Name] = true

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
	if tr.Type == "Boolean" ||
		tr.Type == "Int" ||
		tr.Type == "Int32" ||
		tr.Type == "Int64" ||
		tr.Type == "Float" ||
		tr.Type == "Float32" ||
		tr.Type == "Float64" ||
		tr.Type == "Bytes" ||
		tr.Type == "String" {
		return nil
	}

	// if not custom declared type, error
	if c.declarations[tr.Type] == nil {
		return fmt.Errorf("unknown type '%v' at %v", tr.Type, tr.Pos.String())
	}

	// recurse on type
	return c.checkDeclaration(c.declarations[tr.Type], seen, path)
}

// parse declaration as stream if it has a stream annotation
func (c *Compiler) parseStream(declaration *Declaration) (*stream, error) {
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
	return &stream{
		name:        streamName,
		key:         streamKey,
		declaration: declaration,
	}, nil
}
