package schema

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/linkedin/goavro/v2"
)

// BuildAvroSchema compiles the stream into an Avro schema
func (s *StreamDef) BuildAvroSchema() (string, error) {
	return s.buildAvroSchema(true)
}

// BuildCanonicalAvroSchema compiles the stream into an Avro schema
// in canonical form (compact and without doc)
// NOTE: We're just returning the avro schema without doc fields.
// Canonical Avro isn't actually well defined and goavro's
// function for transforming to canonical drops logicalType fields,
// which won't do for us.
func (s *StreamDef) BuildCanonicalAvroSchema() (string, error) {
	return s.buildAvroSchema(false)
}

func (s *StreamDef) buildAvroSchema(doc bool) (string, error) {
	decl := s.Compiler.Declarations[s.TypeName]

	definedNames := make(map[string]bool)
	avro := s.buildAvroRecord(decl.Type, doc, definedNames)

	json, err := json.Marshal(avro)
	if err != nil {
		return "", fmt.Errorf("cannot marshal avro schema: %v", err.Error())
	}

	_, err = goavro.NewCodec(string(json))
	if err != nil {
		return "", err
	}

	return string(json), nil
}

func (s *StreamDef) buildAvroRecord(t *Type, doc bool, definedNames map[string]bool) interface{} {
	if definedNames[t.Name] {
		return t.Name
	}
	definedNames[t.Name] = true

	fields := make([]map[string]interface{}, len(t.Fields))
	for idx, field := range t.Fields {
		fields[idx] = map[string]interface{}{
			"name": field.Name,
			"type": s.buildAvroTypeRef(field.Type, doc, definedNames),
		}
		if doc {
			content := strings.TrimSpace(field.Doc)
			if content != "" {
				fields[idx]["doc"] = field.Doc
			}
		}
	}

	record := map[string]interface{}{
		"type":   "record",
		"name":   t.Name,
		"fields": fields,
	}

	if doc {
		content := strings.TrimSpace(t.Doc)
		if content != "" {
			record["doc"] = content
		}
	}

	return record
}

func (s *StreamDef) buildAvroEnum(e *Enum, doc bool, definedNames map[string]bool) interface{} {
	if definedNames[e.Name] {
		return e.Name
	}
	definedNames[e.Name] = true

	avro := map[string]interface{}{
		"type":    "enum",
		"name":    e.Name,
		"symbols": e.Members,
	}
	return avro
}

func (s *StreamDef) buildAvroTypeRef(tr *TypeRef, doc bool, definedNames map[string]bool) interface{} {
	var avro interface{}
	if tr.Array != nil {
		avro = map[string]interface{}{
			"type":  "array",
			"items": s.buildAvroTypeRef(tr.Array, doc, definedNames),
		}
	} else {
		avro = s.buildAvroTypeName(tr.Type, doc, definedNames)
	}

	if !tr.Required {
		return []interface{}{
			"null",
			avro,
		}
	}

	return avro
}

func (s *StreamDef) buildAvroTypeName(name string, doc bool, definedNames map[string]bool) interface{} {
	if isPrimitiveTypeName(name) {
		return s.buildAvroPrimitiveTypeName(name, doc, definedNames)
	}

	decl := s.Compiler.Declarations[name]
	if decl == nil {
		panic(fmt.Errorf("type '%v' is neither primitive nor declared", name))
	}

	if decl.Enum != nil {
		return s.buildAvroEnum(decl.Enum, doc, definedNames)
	}

	if decl.Type != nil {
		return s.buildAvroRecord(decl.Type, doc, definedNames)
	}

	panic(fmt.Errorf("declaration for type '%v' is neither enum nor record", name))
}

func (s *StreamDef) buildAvroPrimitiveTypeName(name string, doc bool, definedNames map[string]bool) interface{} {
	base, arg := splitPrimitiveName(name)
	switch base {
	case "Boolean":
		return "boolean"
	case "Bytes":
		if arg == 0 {
			return "bytes"
		}

		// the fixed type is unfortunately named in Avro
		if definedNames[name] {
			return name
		}
		definedNames[name] = true

		return map[string]interface{}{
			"type": "fixed",
			"size": arg,
			"name": name,
		}
	case "Float":
		if arg == 32 {
			return "float"
		}
		return "double"
	case "Int":
		if arg == 32 {
			return "int"
		}
		return "long"
	case "Numeric":
		return map[string]interface{}{
			"type":        "bytes",
			"logicalType": "decimal",
			"precision":   128,
			"scale":       0,
		}
	case "String":
		return "string"
	case "Timestamp":
		return map[string]interface{}{
			"type":        "long",
			"logicalType": "timestamp-millis",
		}
	}

	panic(fmt.Errorf("type '%v' is not a primitive", name))
}
