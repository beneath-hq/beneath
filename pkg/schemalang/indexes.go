package schemalang

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Indexes represents a set of indexes for a schema. The first index should have Key == true.
type Indexes []Index

// Index represents a data index on a schema fields. It should only be used in KeySpec.
type Index struct {
	Fields    []string `json:"fields"`
	Key       bool     `json:"key,omitempty"`
	Normalize bool     `json:"normalize,omitempty"`
}

// Sort sorts indexes in-place, such that the key comes first, then sorted by Fields lexicographically
func (s Indexes) Sort() {
	sort.Slice(s, func(i, j int) bool {
		// key always comes first
		if s[i].Key {
			return true
		}
		if s[j].Key {
			return false
		}

		// then sort by fields
		for x := 0; true; x++ {
			if len(s[i].Fields) == x {
				return true
			}

			if len(s[j].Fields) == x {
				return false
			}

			comp := strings.Compare(s[i].Fields[x], s[j].Fields[x])
			if comp < 0 {
				return true
			} else if comp > 0 {
				return false
			}
		}
		return false // can't happen
	})
}

// CanonicalJSON returns a canonical representation of the indexes (i.e. it can be used for comparisons)
func (s Indexes) CanonicalJSON() string {
	s.Sort()
	// marshal (preserves key order)
	json, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(json)
}

// Check checks that key spec is valid for schema
func (s Indexes) Check(schema Schema) error {
	record, ok := schema.(*Record)
	if !ok {
		return fmt.Errorf("cannot check key spec against non-record schema (actual schema type: '%s')", schema.GetType())
	}

	if len(s) == 0 || !s[0].Key {
		return fmt.Errorf("no key specified for schema '%s'", record.Name)
	}

	fields := make(map[string]*RecordField, len(record.Fields))
	for _, field := range record.Fields {
		fields[field.Name] = field
	}

	for i, index := range s {
		err := s.checkIndex(record, index, fields)
		if err != nil {
			return err
		}

		if index.Key && i != 0 {
			return fmt.Errorf("found multiple key indexes for schema '%s'", record.Name)
		}
	}

	err := s.checkMutuallyExclusive(record)
	if err != nil {
		return err
	}

	return nil
}

func (s Indexes) checkMutuallyExclusive(record *Record) error {
	if len(s) == 0 {
		return nil
	}

	for i := 0; i < len(s); i++ {
		x := s[i].Fields
		for j := i + 1; j < len(s); j++ {
			y := s[j].Fields
			for k := 0; true; k++ {
				// if got to the end of an index, not mutually exclusive
				if len(x) <= k || len(y) <= k {
					return fmt.Errorf("the indexes on type '%v' are not mutually exclusive", record.Name)
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

func (s Indexes) checkIndex(record *Record, index Index, fields map[string]*RecordField) error {
	if len(index.Fields) == 0 {
		return fmt.Errorf("found index with no fields for schema '%s'", record.Name)
	}

	seen := make(map[string]bool, len(index.Fields))
	for _, fieldName := range index.Fields {
		// check it exists
		if fields[fieldName] == nil {
			return fmt.Errorf("index field '%v' doesn't exist in type '%v'", fieldName, record.Name)
		}

		// check not used twice
		if seen[fieldName] {
			return fmt.Errorf("field '%v' used twice in index for type '%v'", fieldName, record.Name)
		}
		seen[fieldName] = true

		// check type
		err := s.checkIndexField(record, fields[fieldName])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s Indexes) checkIndexField(record *Record, field *RecordField) error {
	schema := field.Type

	// special case for the common error of having a nullable index field
	if schema.GetType() == NullableType {
		return fmt.Errorf("field '%s' in type '%s' cannot be used as index because it is optional", field.Name, record.Name)
	}

	if schema.GetType() == FixedType {
		return nil
	}

	primitive, ok := schema.(*Primitive)
	if ok && s.checkPrimitive(primitive) {
		return nil
	}

	return fmt.Errorf("field '%s' in type '%s' cannot be used as index", field.Name, record.Name)
}

func (s Indexes) checkPrimitive(primitive *Primitive) bool {
	if primitive.LogicalType != "" {
		switch primitive.LogicalType {
		case TimestampMillisLogicalType:
			return true
		case UUIDLogicalType:
			return true
		default:
			return false
		}
	}

	switch primitive.Type {
	case StringType:
		return true
	case BytesType:
		return true
	case IntType:
		return true
	case LongType:
		return true
	default:
		return false
	}
}
