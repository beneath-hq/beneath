package sdl

import (
	"encoding/json"
	"fmt"
	"log"
)

// BuildAvroSchema compiles the stream into an Avro schema
func (s *StreamDef) BuildAvroSchema(doc bool) (string, error) {
	decl := s.Compiler.Declarations[s.TypeName]

	definedNames := make(map[string]bool)
	avro := s.buildAvroRecord(decl.Type, definedNames)

	json, err := json.Marshal(avro)
	if err != nil {
		return "", fmt.Errorf("cannot marshal avro schema: %v", err.Error())
	}

	return string(json), nil
}

func (s *StreamDef) buildAvroRecord(t *Type, definedNames map[string]bool) interface{} {
	if definedNames[t.Name] {
		return t.Name
	}
	definedNames[t.Name] = true

	fields := make([]map[string]interface{}, len(t.Fields))
	for idx, field := range t.Fields {
		fields[idx] = map[string]interface{}{
			"name": field.Name,
			"type": s.buildAvroTypeRef(field.Type, definedNames),
		}
	}

	record := map[string]interface{}{
		"type":   "record",
		"name":   t.Name,
		"doc":    t.Doc,
		"fields": fields,
	}

	return record
}

func (s *StreamDef) buildAvroEnum(e *Enum, definedNames map[string]bool) interface{} {
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

func (s *StreamDef) buildAvroTypeRef(tr *TypeRef, definedNames map[string]bool) interface{} {
	var avro interface{}
	if tr.Array != nil {
		avro = map[string]interface{}{
			"type":  "array",
			"items": s.buildAvroTypeRef(tr.Array, definedNames),
		}
	} else {
		avro = s.buildAvroTypeName(tr.Type, definedNames)
	}

	if tr.Required {
		return []interface{}{
			"null",
			avro,
		}
	}

	return avro
}

func (s *StreamDef) buildAvroTypeName(name string, definedNames map[string]bool) interface{} {
	if isPrimitiveTypeName(name) {
		return s.buildAvroPrimitiveTypeName(name, definedNames)
	}

	decl := s.Compiler.Declarations[name]
	if decl == nil {
		log.Fatalf("type '%v' is neither primitive nor declared", name)
		return nil
	}

	if decl.Enum != nil {
		return s.buildAvroEnum(decl.Enum, definedNames)
	}

	if decl.Type != nil {
		return s.buildAvroRecord(decl.Type, definedNames)
	}

	log.Fatalf("declaration for type '%v' is neither enum nor record", name)
	return nil
}

func (s *StreamDef) buildAvroPrimitiveTypeName(name string, definedNames map[string]bool) interface{} {
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

	log.Fatalf("type '%v' is not a primitive", name)
	return nil
}
