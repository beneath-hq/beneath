package schema

import (
	"encoding/json"
	"fmt"
	"log"
)

// BuildBigQuerySchema compiles the stream into a BigQuery schema
func (s *StreamDef) BuildBigQuerySchema() (string, error) {
	decl := s.Compiler.Declarations[s.TypeName]
	schema := s.buildBigQueryFields(decl.Type.Fields)

	json, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("cannot marshal bigquery schema: %v", err.Error())
	}

	return string(json), nil
}

func (s *StreamDef) buildBigQueryFields(fields []*Field) interface{} {
	schema := make([]interface{}, len(fields))
	for idx, field := range fields {
		schema[idx] = s.buildBigQueryField(field.Type, field.Name)
	}
	return schema
}

func (s *StreamDef) buildBigQueryField(tr *TypeRef, name string) interface{} {
	schema := map[string]interface{}{
		"name": name,
	}

	// set mode
	if tr.Required {
		schema["mode"] = "REQUIRED"
	} else {
		schema["mode"] = "NULLABLE"
	}

	// handle if array
	if tr.Array != nil {
		// override mode and tr
		schema["mode"] = "REPEATED"
		tr = tr.Array

		// nested arrays not allowed
		if tr.Array != nil {
			log.Fatalf("bigquery schema builder: encountered nested array at %v", tr.Pos)
			return nil
		}
	}

	// set type
	if isPrimitiveTypeName(tr.Type) {
		// handle primitive type
		schema["type"] = s.buildBigQueryPrimitiveTypeName(tr.Type)
	} else {
		// handle declaration
		decl := s.Compiler.Declarations[tr.Type]
		if decl == nil {
			log.Fatalf("bigquery schema builder: type '%v' is neither primitive nor declared", tr.Type)
			return nil
		}

		// enums coerced to string
		if decl.Enum != nil {
			schema["type"] = "STRING"
		}

		// recurse on record fields
		if decl.Type != nil {
			schema["type"] = "RECORD"
			schema["fields"] = s.buildBigQueryFields(decl.Type.Fields)
		}
	}

	// done
	return schema
}

func (s *StreamDef) buildBigQueryPrimitiveTypeName(name string) interface{} {
	base, _ := splitPrimitiveName(name)
	switch base {
	case "Boolean":
		return "BOOLEAN"
	case "Bytes":
		// no special handling for arg
		return "BYTES"
	case "Float":
		// no special handling for arg
		return "FLOAT"
	case "Int":
		// no special handling for arg
		return "INTEGER"
	case "Numeric":
		// the built-in NUMERIC type is too limited, so we use strings
		return "STRING"
	case "String":
		return "STRING"
	case "Timestamp":
		return "TIMESTAMP"
	}

	log.Fatalf("type '%v' is not a primitive", name)
	return nil
}
