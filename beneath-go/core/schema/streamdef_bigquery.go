package schema

import (
	"encoding/json"
	"fmt"
)

// BuildBigQuerySchema compiles the stream into a BigQuery schema
func (s *StreamDef) BuildBigQuerySchema() (string, error) {
	// generate bigquery schema
	decl := s.Compiler.Declarations[s.TypeName]
	schema := s.buildBigQueryFields(decl.Type.Fields)

	// inject internal fields
	schema = append(
		schema.([]interface{}),
		map[string]interface{}{
			"name": "__key",
			"type": "BYTES",
			"mode": "REQUIRED",
		},
		map[string]interface{}{
			"name": "__timestamp",
			"type": "TIMESTAMP",
			"mode": "REQUIRED",
		},
	)

	// convert to json
	json, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("cannot marshal bigquery schema: %v", err.Error())
	}

	// finito
	return string(json), nil
}

func (s *StreamDef) buildBigQueryFields(fields []*Field) interface{} {
	schema := make([]interface{}, len(fields))
	for idx, field := range fields {
		schema[idx] = s.buildBigQueryField(field.Type, field.Name, field.Doc)
	}
	return schema
}

func (s *StreamDef) buildBigQueryField(tr *TypeRef, name string, doc string) interface{} {
	schema := map[string]interface{}{
		"name":        name,
		"description": doc,
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
			panic(fmt.Errorf("bigquery schema builder: encountered nested array at %v", tr.Pos))
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
			panic(fmt.Errorf("bigquery schema builder: type '%v' is neither primitive nor declared", tr.Type))
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
		// BigQuery doesn't allow clustering on bytes
		return "STRING"
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

	panic(fmt.Errorf("type '%v' is not a primitive", name))
}
