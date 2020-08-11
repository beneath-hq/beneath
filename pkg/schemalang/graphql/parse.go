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
// the object annotated with "@stream" is assigned to Root (and
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

	// check there's at least one stream
	if s.Root == nil {
		return fmt.Errorf("no streams declared in input")
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

	// if it's a stream (is a Type with annotations), parse it
	if declaration.Type != nil && len(declaration.Type.Annotations) != 0 {
		err := s.parseStream(declaration)
		if err != nil {
			return err
		}
	}

	return nil
}

// parse declaration as stream if it has a stream annotation
func (s *Schema) parseStream(declaration *Declaration) error {
	// if we already parsed a stream, error
	if s.Root != nil {
		return fmt.Errorf("found multiple objects in schema with '@stream' annotation - you can only define one stream in a schema")
	}

	// prepare data to extract from annotations
	var streamName string
	var keyIndex Index
	var secondaryIndexes []Index

	// parse annotations
	for _, ann := range declaration.Type.Annotations {
		switch ann.Name {
		case "stream":
			name, err := s.parseStreamAnnotation(declaration, ann)
			if err != nil {
				return err
			}
			streamName = name
		case "key":
			index, err := s.parseIndexAnnotation(declaration, ann)
			if err != nil {
				return err
			}
			keyIndex = index
		case "index":
			index, err := s.parseIndexAnnotation(declaration, ann)
			if err != nil {
				return err
			}
			secondaryIndexes = append(secondaryIndexes, index)
		default:
			return fmt.Errorf("unknown annotation '@%v' at %v", ann.Name, declaration.Pos.String())
		}
	}

	// if no streamName set, @stream annotation not found
	if streamName == "" {
		return fmt.Errorf("cannot have '@key' or '@index' annotations on non-stream declaration at %v -- are you missing an @stream annotation?", declaration.Pos.String())
	}

	// check params
	if len(keyIndex.Fields) == 0 {
		return fmt.Errorf("missing annotation '@key' with 'fields' arg in stream declaration at %v", declaration.Pos.String())
	}

	// done
	s.Root = declaration
	s.Name = streamName
	s.Key = keyIndex
	s.Indexes = secondaryIndexes

	return nil
}

func (s *Schema) parseStreamAnnotation(declaration *Declaration, annotation *Annotation) (string, error) {
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

func (s *Schema) parseIndexAnnotation(declaration *Declaration, annotation *Annotation) (Index, error) {
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
			index.Normalize = arg.Value.Symbol == "true"
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
