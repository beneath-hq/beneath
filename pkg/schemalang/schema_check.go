package schemalang

import (
	"fmt"
	"regexp"
)

const (
	maxFieldNameLen = 64
	maxTypeNameLen  = 64
)

var (
	snakeCaseRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")
	fixedNameRegex = regexp.MustCompile("^bytes[1-9][0-9]+$")
)

// Check checks that s is a valid schema.
func Check(s Schema) error {
	if _, ok := s.(*Record); !ok {
		return fmt.Errorf("expected 'record' type as schema root, got '%s'", s.GetType())
	}
	c := &checker{
		NamedTypes: make(map[string]Schema),
	}
	return c.checkSchema(s)
}

type checker struct {
	NamedTypes map[string]Schema // used to check valid Refs
	NamesPath  []string          // used to check for cycles
}

func (c *checker) pushName(name string) {
	c.NamesPath = append(c.NamesPath, name)
}

func (c *checker) popName() {
	c.NamesPath = c.NamesPath[:len(c.NamesPath)-1]
}

func (c *checker) trackName(name string, schema Schema) error {
	if len(name) > maxTypeNameLen {
		return fmt.Errorf("type name '%s' exceeds limit of %d characters", name, maxTypeNameLen)
	}

	if fixedNameRegex.Match([]byte(name)) {
		return fmt.Errorf("type name '%s' cannot be used on non-fixed type", name)
	}

	if c.NamedTypes[name] != nil {
		return fmt.Errorf("found multiple type definitions with name '%s'", name)
	}
	c.NamedTypes[name] = schema

	return nil
}

func (c *checker) checkSchema(s Schema) error {
	switch val := s.(type) {
	case *Primitive:
		return nil
	case *Nullable:
		return c.checkSchema(val.NonNullType)
	case *Fixed:
		return nil
	case *Array:
		return c.checkArray(val)
	case *Record:
		return c.checkRecord(val)
	case *RecordField:
		return fmt.Errorf("found definition of record field '%s' outside record", val.Name)
	case *Enum:
		return c.checkEnum(val)
	case *Ref:
		return c.checkRef(val)
	default:
		panic(fmt.Errorf("unexpected Avro schema type %T", s))
	}
}

func (c *checker) checkArray(val *Array) error {
	if _, ok := val.ItemType.(*Array); ok {
		return fmt.Errorf("nested lists are not allowed")
	}
	if _, ok := val.ItemType.(*Nullable); ok {
		return fmt.Errorf("type wrapped by list cannot be nullable")
	}
	return c.checkSchema(val.ItemType)
}

func (c *checker) checkEnum(val *Enum) error {
	// check has more than one member
	if len(val.Symbols) == 0 {
		return fmt.Errorf("enum '%v' must have at least one symbol", val.Name)
	}

	// check if symbol is unique
	seen := make(map[string]bool)
	for _, name := range val.Symbols {
		if seen[name] {
			return fmt.Errorf("symbol '%s' declared twice in enum '%s'", name, val.Name)
		}
		seen[name] = true
	}

	err := c.trackName(val.Name, val)
	if err != nil {
		return err
	}

	return nil
}

func (c *checker) checkRecord(val *Record) error {
	// check fields not empty
	if len(val.Fields) == 0 {
		return fmt.Errorf("type '%s' does not define any fields", val.Name)
	}

	err := c.trackName(val.Name, val)
	if err != nil {
		return err
	}

	c.pushName(val.Name)
	seen := make(map[string]bool)
	for _, field := range val.Fields {
		// check not declared twice
		if seen[field.Name] {
			return fmt.Errorf("field '%v' declared twice in record '%v'", field.Name, val.Name)
		}
		seen[field.Name] = true

		// recursive check
		err := c.checkRecordField(field, val)
		if err != nil {
			return err
		}
	}
	c.popName()

	return nil
}

func (c *checker) checkRecordField(field *RecordField, record *Record) error {
	if !snakeCaseRegex.MatchString(field.Name) {
		return fmt.Errorf("field name '%s' in record '%s' is not underscore case", field.Name, record.Name)
	}

	if len(field.Name) > maxFieldNameLen {
		return fmt.Errorf("field name '%s' exceeds limit of %d characters", field.Name, maxFieldNameLen)
	}

	if len(field.Name) >= 2 && field.Name[0:2] == "__" {
		return fmt.Errorf("field name '%s' is a reserved identifier", field.Name)
	}

	err := c.checkSchema(field.Type)
	if err != nil {
		return err
	}

	return nil
}

func (c *checker) checkRef(val *Ref) error {
	if c.NamedTypes[val.Name] == nil {
		return fmt.Errorf("found reference to unknown type '%s'", val.Name)
	}
	for _, name := range c.NamesPath {
		if val.Name == name {
			return fmt.Errorf("found cyclic type '%s' (not supported)", val.Name)
		}
	}
	return nil
}
