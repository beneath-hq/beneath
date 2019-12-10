package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/golang-collections/collections/set"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
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

var (
	snakeCaseRegex *regexp.Regexp
)

func init() {
	snakeCaseRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")
}

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
			// index fields by name -> type
			fields := make(map[string]*TypeRef, len(declaration.Type.Fields))
			for _, field := range declaration.Type.Fields {
				// check not declared twice
				if fields[field.Name] != nil {
					return fmt.Errorf("field '%v' declared twice in type '%v'", field.Name, declaration.Type.Name)
				}

				// check is snake case
				if !snakeCaseRegex.MatchString(field.Name) {
					return fmt.Errorf("field name '%v' in type '%v' is not underscore case", field.Name, declaration.Type.Name)
				}

				// store it
				fields[field.Name] = field.Type
			}

			// if it's a stream, check key fields a) exist, b) not used twice, c) are primitive, d) not optional
			stream := c.Streams[declaration.Type.Name]
			if stream != nil {
				err := c.checkStreamDef(stream, declaration, fields)
				if err != nil {
					return err
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

// checks well-formedness of a stream definition
func (c *Compiler) checkStreamDef(def *StreamDef, declaration *Declaration, fields map[string]*TypeRef) error {
	indexes := append([]Index{def.KeyIndex}, def.SecondaryIndexes...)

	for _, index := range indexes {
		err := c.checkIndex(index, declaration, fields)
		if err != nil {
			return err
		}
	}

	return c.checkIndexesMutuallyExclusive(indexes, declaration)
}

// checks that no index can substitute another
func (c *Compiler) checkIndexesMutuallyExclusive(indexes []Index, declaration *Declaration) error {
	for i := 0; i < len(indexes)-1; i++ {
		x := indexes[i].Fields
		for j := i + 1; j < len(indexes); j++ {
			y := indexes[j].Fields
			for k := 0; true; k++ {
				// if got to the end of an index, not mutually exclusive
				if len(x) <= k || len(y) <= k {
					return fmt.Errorf("the indexes on type '%v' are not mutually exclusive", declaration.Type.Name)
				}

				// if the fields differ, they're mutually exclusive
				if x[k] != y[k] {
					break
				}
				// continue
			}
		}
	}
	return nil
}

// checks well-formedness of an index
func (c *Compiler) checkIndex(index Index, declaration *Declaration, fields map[string]*TypeRef) error {
	seen := make(map[string]bool, len(index.Fields))
	for _, field := range index.Fields {
		// check it exists
		if fields[field] == nil {
			return fmt.Errorf("index field '%v' doesn't exist in type '%v'", field, declaration.Type.Name)
		}

		// check not used twice
		if seen[field] {
			return fmt.Errorf("field '%v' used twice in index for type '%v'", field, declaration.Type.Name)
		}
		seen[field] = true

		// check it has a primitive type
		if !isIndexableType(fields[field]) {
			return fmt.Errorf("field '%v' in type '%v' cannot be used as index", field, declaration.Type.Name)
		}

		// check not optional
		if !fields[field].Required {
			return fmt.Errorf("field '%v' in type '%v' cannot be used as index because it is optional", field, declaration.Type.Name)
		}
	}
	return nil
}

// parse declaration as stream if it has a stream annotation
func (c *Compiler) parseStream(declaration *Declaration) (*StreamDef, error) {
	// if not a stream, return empty
	if declaration.Type == nil || len(declaration.Type.Annotations) == 0 {
		return nil, nil
	}

	// prepare data to extract from annotations
	var streamName string
	var keyIndex Index
	var secondaryIndexes []Index

	// parse annotations
	for _, ann := range declaration.Type.Annotations {
		switch ann.Name {
		case "stream":
			name, err := c.parseStreamAnnotation(declaration, ann)
			if err != nil {
				return nil, err
			}
			streamName = name
		case "key":
			index, err := c.parseIndexAnnotation(declaration, ann)
			if err != nil {
				return nil, err
			}
			keyIndex = index
		case "index":
			index, err := c.parseIndexAnnotation(declaration, ann)
			if err != nil {
				return nil, err
			}
			secondaryIndexes = append(secondaryIndexes, index)
		default:
			return nil, fmt.Errorf("unknown annotation '@%v' at %v", ann.Name, declaration.Pos.String())
		}
	}

	// if no streamName set, @stream annotation not found
	if streamName == "" {
		return nil, fmt.Errorf("cannot have @key or @index annotations on non-stream declaration at %v -- are you missing an @stream annotation?", declaration.Pos.String())
	}

	// check params
	if len(keyIndex.Fields) == 0 {
		return nil, fmt.Errorf("missing annotation '@key' with 'fields' arg in stream declaration at %v", declaration.Pos.String())
	}

	// done
	return &StreamDef{
		Name:             streamName,
		Description:      strings.TrimSpace(declaration.Type.Doc),
		TypeName:         declaration.Type.Name,
		KeyIndex:         keyIndex,
		SecondaryIndexes: secondaryIndexes,
		Compiler:         c,
	}, nil
}

func (c *Compiler) parseStreamAnnotation(declaration *Declaration, annotation *Annotation) (string, error) {
	// default name is snake cased plural
	streamName := strcase.ToSnake(inflection.Plural(declaration.Type.Name))

	// parse for "name" arg
	for _, arg := range annotation.Arguments {
		switch arg.Name {
		case "name":
			if arg.Value.String == "" {
				return "", fmt.Errorf("stream arg 'name' at %v must be a non-empty string", arg.Pos.String())
			}

			streamName = strings.ReplaceAll(arg.Value.String, "-", "_")
			if !snakeCaseRegex.MatchString(streamName) {
				return "", fmt.Errorf("stream name '%v' at %v is not a valid stream name (only alphanumeric characters, '-' and '_' allowed)", arg.Value.String, arg.Pos.String())
			}
		default:
			return "", fmt.Errorf("unknown arg '%v' for annotation '@%v' at %v", arg.Name, annotation.Name, arg.Pos.String())
		}
	}

	return streamName, nil
}

func (c *Compiler) parseIndexAnnotation(declaration *Declaration, annotation *Annotation) (Index, error) {
	// parse "fields" and "normalize" args
	var index Index
	for _, arg := range annotation.Arguments {
		switch arg.Name {
		case "normalize":
			// must be boolean symbol
			if arg.Value.Symbol == "" {
				return index, fmt.Errorf("key arg 'normalize' at %v must be a boolean", arg.Pos.String())
			}
			// parse as "true" or "false"
			index.Denormalize = arg.Value.Symbol != "true"
		case "fields":
			err := fmt.Errorf("arg 'fields' at %v is not a string or array of strings", arg.Pos.String())
			if arg.Value.String != "" {
				index.Fields = []string{arg.Value.String}
			} else if arg.Value.Array != nil && len(arg.Value.Array) > 0 {
				index.Fields = make([]string, len(arg.Value.Array))
				for idx, val := range arg.Value.Array {
					if val.String == "" {
						return index, err
					}
					index.Fields[idx] = val.String
				}
			} else {
				return index, err
			}
		default:
			return index, fmt.Errorf("unknown arg '%v' for annotation '@%v' at %v", arg.Name, annotation.Name, arg.Pos.String())
		}
	}

	return index, nil
}
