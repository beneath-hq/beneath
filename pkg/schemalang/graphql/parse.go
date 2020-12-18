package graphql

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
)

var (
	snakeCaseRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")
)

// Schema represents a parsed and checked GraphQL schema where
// the object annotated with "@schema" is assigned to Root (and
// its "@key" and "@index" annotations parsed accordingly).
type Schema struct {
	Root         *Declaration
	Name         string
	Key          Index
	Indexes      []Index
	AST          *File
	Declarations map[string]*Declaration
}

// Index represents an index on one or more columns
type Index struct {
	Fields    []string
	Normalize bool
}

// ParseSDL parses a Beneath-flavoured GraphQL SDL schema
func ParseSDL(sdl string) (*Schema, error) {
	schema := &Schema{
		AST:          &File{},
		Declarations: make(map[string]*Declaration),
	}
	err := schema.parse(sdl)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func (s *Schema) parse(sdl string) error {
	// parse ast
	err := getParser().ParseString(sdl, s.AST)
	if err != nil {
		return fmt.Errorf("parse error: %v", err.Error())
	}

	// extract type declarations
	for _, declaration := range s.AST.Declarations {
		err := s.extractDeclaration(declaration)
		if err != nil {
			return err
		}
	}

	// check there's a schema
	if s.Root == nil {
		return fmt.Errorf("no schema declared in input (use @schema annotation on the root type)")
	}

	// done
	return nil
}

func (s *Schema) extractDeclaration(declaration *Declaration) error {
	// get name and trim doc
	var name string
	if declaration.Enum != nil {
		name = declaration.Enum.Name
		declaration.Enum.Doc = strings.TrimSpace(declaration.Enum.Doc)
	} else if declaration.Type != nil {
		name = declaration.Type.Name
		declaration.Type.Doc = strings.TrimSpace(declaration.Type.Doc)
	}

	// check it's not a primitive type name
	_, _, err := ParsePrimitive(name)
	if err != ErrNotPrimitive {
		return fmt.Errorf("declaration of '%v' at %v overlaps with primitive type name", name, declaration.Pos.String())
	}

	// check doesn't exist
	if s.Declarations[name] != nil {
		return fmt.Errorf("name '%v' at %v has already been declared", name, declaration.Pos.String())
	}

	// save declaration
	s.Declarations[name] = declaration

	// if it's a schema (is a Type with annotations), parse it
	if declaration.Type != nil && len(declaration.Type.Annotations) != 0 {
		err := s.parseRoot(declaration)
		if err != nil {
			return err
		}
	}

	return nil
}

// parse declaration as root if it has an @schema annotation
func (s *Schema) parseRoot(declaration *Declaration) error {
	// if we already parsed a root, error
	if s.Root != nil {
		return fmt.Errorf("found multiple types with '@schema' annotation - you can only define one root schema")
	}

	// prepare data to extract from annotations
	var schemaName string
	var keyIndex *Index
	var secondaryIndexes []Index

	// parse type annotations
	for _, ann := range declaration.Type.Annotations {
		switch ann.Name {
		case "stream":
			fallthrough
		case "schema":
			name, err := s.parseSchemaAnnotation(declaration, ann)
			if err != nil {
				return err
			}
			schemaName = name
		case "key":
			index, err := s.parseIndexAnnotationOnType(declaration, ann)
			if err != nil {
				return err
			}
			keyIndex = index
		case "index":
			index, err := s.parseIndexAnnotationOnType(declaration, ann)
			if err != nil {
				return err
			}
			secondaryIndexes = append(secondaryIndexes, *index)
		default:
			return fmt.Errorf("unknown annotation '@%v' at %v", ann.Name, declaration.Pos.String())
		}
	}

	// extract key based on @key annotations on specific fields
	keyIndexFromFields, err := s.parseIndexAnnotationsOnFields(declaration.Type.Fields)
	if err != nil {
		return err
	}

	// we only allow @key on either the type OR individual fields (not both)
	if keyIndex == nil {
		keyIndex = keyIndexFromFields
	} else if keyIndexFromFields != nil {
		return fmt.Errorf("cannot use @key annotation on both type and individual fields for declaration at %v", declaration.Pos.String())
	}

	// if no schema name set, @schema annotation not found
	if schemaName == "" {
		return fmt.Errorf("cannot have '@key' or '@index' annotations on non-schema declaration at %v -- are you missing an @schema annotation?", declaration.Pos.String())
	}

	// check key is correctly set
	if keyIndex == nil || len(keyIndex.Fields) == 0 {
		return fmt.Errorf("missing or incomplete '@key' annotation(s) in schema declaration at %v", declaration.Pos.String())
	}

	// done
	s.Root = declaration
	s.Name = schemaName
	s.Key = *keyIndex
	s.Indexes = secondaryIndexes

	return nil
}

func (s *Schema) parseSchemaAnnotation(declaration *Declaration, annotation *Annotation) (string, error) {
	// default name is snake cased plural
	schemaName := strcase.ToSnake(inflection.Plural(declaration.Type.Name))

	// parse for "name" arg
	for _, arg := range annotation.Arguments {
		switch arg.Name {
		case "name":
			if arg.Value.String == "" {
				return "", fmt.Errorf("@schema argument 'name' at %v must be a non-empty string", arg.Pos.String())
			}

			schemaName = strings.ReplaceAll(arg.Value.String, "-", "_")
			if !snakeCaseRegex.MatchString(schemaName) {
				return "", fmt.Errorf("schema name='%v' at %v is not a valid schema name (only alphanumeric characters, '-' and '_' allowed)", arg.Value.String, arg.Pos.String())
			}
		default:
			return "", fmt.Errorf("unknown arg '%v' for annotation '@%v' at %v", arg.Name, annotation.Name, arg.Pos.String())
		}
	}

	return schemaName, nil
}

func (s *Schema) parseIndexAnnotationOnType(declaration *Declaration, annotation *Annotation) (*Index, error) {
	// parse "fields" and "normalize" args
	var index Index
	for _, arg := range annotation.Arguments {
		switch arg.Name {
		case "normalize":
			// must be boolean symbol
			if arg.Value.Symbol == "" {
				return nil, fmt.Errorf("key arg 'normalize' at %v must be a boolean", arg.Pos.String())
			}
			// parse as "true" or "false"
			index.Normalize = arg.Value.Symbol == "true"
		case "fields":
			err := fmt.Errorf("arg 'fields' at %v is not a string or array of strings", arg.Pos.String())
			if arg.Value.String != "" {
				index.Fields = []string{arg.Value.String}
			} else if arg.Value.Array != nil && len(arg.Value.Array) > 0 {
				index.Fields = make([]string, len(arg.Value.Array))
				for idx, val := range arg.Value.Array {
					if val.String == "" {
						return nil, err
					}
					index.Fields[idx] = val.String
				}
			} else {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown arg '%v' for annotation '@%v' at %v", arg.Name, annotation.Name, arg.Pos.String())
		}
	}

	return &index, nil
}

func (s *Schema) parseIndexAnnotationsOnFields(fields []*Field) (*Index, error) {
	// extracts key based on @key annotations on fields in the order they appear
	var key []string
	for _, field := range fields {
		for _, ann := range field.Annotations {
			if ann.Name == "key" {
				if len(ann.Arguments) != 0 {
					return nil, fmt.Errorf("unexpected arguments to @key annotation at %v", field.Pos.String())
				}
				key = append(key, field.Name)
			} else {
				return nil, fmt.Errorf("unknown annotation '@%v' at %v", ann.Name, field.Pos.String())
			}
		}
	}

	if len(key) > 0 {
		return &Index{Fields: key}, nil
	}

	return nil, nil
}
