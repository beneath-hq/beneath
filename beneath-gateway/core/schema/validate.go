package schema

import "fmt"

// ValidateSchema ensures that the given Avro schema conforms to the special
// schema requirements enforced by Beneath. It assumes that the input
// schema is a valid, canonical avro schema. The special requirements are:
//   - the top-level type must be a record
//   - the top-level record must specify indexes as an array of arrays of field names
//   - indexes only operate on non-nullable fields of type int, long, bytes, fixed, string or timestamp-millis
//   - maps and enums not supported
//   - aliases and namespaces not used
//   - default values and orderings not supported
//   - only unions of the type [null, XXX] are allowed
//   - recursive types not allowed (i.e., no reusing custom types)
//   - the only supported logicalTypes are
//       - bytes.decimal
//       - long.timestamp-millis
func ValidateSchema(schema interface{}) error {
	// check it's a record field
	schemaRecord, ok := schema.(map[string]interface{})
	if !ok {
		return fmt.Errorf("schema must be a record at the top level")
	}

	// validate that it adheres to our special avro rules
	err := validateAvroSchema(schema)
	if err != nil {
		return err
	}

	// check it provides at least a primary index
	indexes, ok := schemaRecord["indexes"].([]interface{})
	if !ok {
		return fmt.Errorf("schema must provide indexes as an array")
	}
	if len(indexes) == 0 {
		return fmt.Errorf("schema must specify at least one index")
	}

	// check it only indexes fields allowed fields
	indexableField := extractIndexableFields(schemaRecord)
	for _, indexI := range indexes {
		index, ok := indexI.([]interface{})
		if !ok {
			return fmt.Errorf("an index must be an array of field names")
		}

		if len(index) == 0 {
			return fmt.Errorf("cannot have index without fields")
		}

		for _, fieldI := range index {
			field, ok := fieldI.(string)
			if !ok {
				return fmt.Errorf("index element must be string, got %v", fieldI)
			}

			if !indexableField[field] {
				return fmt.Errorf("field <%v> cannot be used in an index", field)
			}
		}
	}

	// passed
	return nil
}

func extractIndexableFields(schema map[string]interface{}) map[string]bool {
	res := make(map[string]bool)

	fields := schema["fields"].([]interface{})
	for _, fieldI := range fields {
		field := fieldI.(map[string]interface{})

		name := field["name"].(string)
		indexable := false

		if t, ok := field["type"].(string); ok {
			indexable = t == "int" || t == "long" || t == "bytes" || t == "string"
		}

		if t, ok := field["type"].(map[string]interface{}); ok {
			indexable = t["type"] == "fixed"
		}

		res[name] = indexable
	}

	return res
}

func validateAvroSchema(schema interface{}) error {
	switch schemaType := schema.(type) {
	case string:
		if schemaType == "null" ||
			schemaType == "boolean" ||
			schemaType == "int" ||
			schemaType == "long" ||
			schemaType == "float" ||
			schemaType == "double" ||
			schemaType == "bytes" ||
			schemaType == "string" {
			return nil
		}
		return fmt.Errorf("type not supported: %v", schemaType)
	case []interface{}:
		if len(schemaType) == 2 && schemaType[0] == "null" {
			return validateAvroSchema(schemaType[1])
		}
		return fmt.Errorf("unions can have only two types and the first must be 'null'")
	case map[string]interface{}:
		t := schemaType["type"]

		if schemaType["default"] != nil {
			return fmt.Errorf("default not supported: %v", schemaType)
		}

		if schemaType["order"] != nil {
			return fmt.Errorf("order not supported: %v", schemaType)
		}

		if lt := schemaType["logicalType"]; lt != nil {
			if t == "bytes" && lt == "decimal" {
				return nil
			}
			if t == "long" && lt == "timestamp-millis" {
				return nil
			}
			return fmt.Errorf("logicalType not supported: %v.%v", t, lt)
		}

		if t == "fixed" {
			return nil
		}

		if t == "array" {
			return validateAvroSchema(schemaType["items"])
		}

		if t == "record" {
			fields, ok := schemaType["fields"].([]interface{})
			if !ok {
				return fmt.Errorf("record type doesn't have fields: %v", schemaType)
			}

			if len(fields) == 0 {
				return fmt.Errorf("record fields can't be empty: %v", schemaType)
			}

			for _, field := range fields {
				err := validateAvroSchema(field)
				if err != nil {
					return err
				}
			}

			return nil
		}

		return validateAvroSchema(t)
	default:
		return fmt.Errorf("unknown schema type: %T", schema)
	}
}
