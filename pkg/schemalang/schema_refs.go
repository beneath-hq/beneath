package schemalang

import "fmt"

// ExtractRefs returns a map of sub-schemas referenced with a Ref in schema
func ExtractRefs(schema Schema) map[string]Schema {
	e := extractRefs{
		NamedTypes: make(map[string]Schema),
		Refs:       make(map[string]bool),
	}
	e.extract(schema)

	result := make(map[string]Schema)
	for name := range e.Refs {
		t := e.NamedTypes[name]
		if t == nil {
			panic(fmt.Errorf("found ref '%s' to undefined type in schema", name))
		}
		result[name] = t
	}

	return result
}

type extractRefs struct {
	NamedTypes map[string]Schema
	Refs       map[string]bool
}

func (e extractRefs) extract(schema Schema) {
	switch s := schema.(type) {
	case *Primitive:
	case *Array:
		e.extract(s.ItemType)
	case *Nullable:
		e.extract(s.NonNullType)
	case *Fixed:
	case *Enum:
		e.NamedTypes[s.Name] = s
	case *Record:
		e.NamedTypes[s.Name] = s
		for _, field := range s.Fields {
			e.extract(field.Type)
		}
	case *Ref:
		e.Refs[s.Name] = true
	default:
		panic(fmt.Errorf("unexpected schema type %T", s))
	}
}
