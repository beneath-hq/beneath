package codec

import (
	"bytes"
	"fmt"
	"log"

	"github.com/beneath-core/beneath-go/core/codec/ext/tuple"
	"github.com/beneath-core/beneath-go/core/queryparse"
)

// KeyRange represents a range of keys
type KeyRange struct {
	Base     []byte
	RangeEnd []byte
	Unique   bool
}

// Contains is true if the KeyRange contains a key
func (r *KeyRange) Contains(key []byte) bool {
	if r.Unique {
		return bytes.Equal(r.Base, key)
	} else if r.RangeEnd == nil {
		return bytes.Compare(key, r.Base) >= 0
	}
	return bytes.Compare(key, r.Base) >= 0 && bytes.Compare(key, r.RangeEnd) < 0
}

// CheckUnique is true iff key range identifies one exact key
func (r *KeyRange) CheckUnique() bool {
	return r != nil && r.Unique
}

// NewKeyRange builds a new key range based on a query and a key codec
func NewKeyRange(q queryparse.Query, c *KeyCodec) (r *KeyRange, err error) {
	// prepare key
	key := make(tuple.Tuple, len(q))

	// iterate through key fields
	for idx, field := range c.fields {
		// get condition for field
		cond := q[field]
		if cond == nil {
			if len(c.fields) == 1 {
				return nil, fmt.Errorf("expected lookup on and only on key field '%s'", field)
			}
			return nil, fmt.Errorf("expected field '%s' in query (composite keys are indexed starting with the leftmost field)", field)
		}

		// get avro type
		avroType := c.avroTypes[idx]

		// parse arg1 value
		arg1, err := parseJSONValue(avroType, cond.Arg1)
		if err != nil {
			return nil, err
		}

		// set val in key we're building
		key[idx] = arg1

		// _eq is the only condition that is potentially not terminal
		// so handle it separately
		if cond.Op == queryparse.ConditionOpEq {
			// next step depends on how far we've come
			if idx+1 == len(c.fields) {
				// we've added _eq constraints for every key field, so we're done and it's a unique key
				return &KeyRange{
					Base:   key.Pack(),
					Unique: true,
				}, nil
			} else if idx+1 == len(q) {
				// we've added _eq constraints for some subset of leftmost key fields, so we're done and it's a prefix key
				packed := key.Pack()
				return &KeyRange{
					Base:     packed,
					RangeEnd: tuple.PrefixSuccessor(packed),
				}, nil
			} else {
				// continue for loop (skipping code below)
				continue
			}
		}

		// now we know we're not handling a ConditionOpEq and therefore we have
		// to return from the function now (every other condition is terminal)

		// every other condition is terminal, so check there's no more conditions
		if idx+1 != len(q) {
			return nil, fmt.Errorf("cannot use '%s' on field '%s' because you have constraints on fields that appear later in the key", cond.Op.String(), field)
		}

		// pack key (and a copy with the last element removed)
		packed := key.Pack()
		packedParent := key[:len(key)-1].Pack()

		// handle ops that only rely on cond.Arg1 (which we've already parsed)
		switch cond.Op {
		case queryparse.ConditionOpPrefix:
			// prefix should only work on strings, bytes and fixed
			if !canPrefixLookup(avroType) {
				return nil, fmt.Errorf("cannot use '_prefix' on field '%s' because it only works on string and byte types", field)
			}
			return &KeyRange{
				Base:     packed,
				RangeEnd: tuple.BytesTypePrefixSuccessor(packed),
			}, nil
		case queryparse.ConditionOpGt:
			return &KeyRange{
				Base:     tuple.Successor(packed),
				RangeEnd: tuple.PrefixSuccessor(packedParent),
			}, nil
		case queryparse.ConditionOpGte:
			return &KeyRange{
				Base:     packed,
				RangeEnd: tuple.PrefixSuccessor(packedParent),
			}, nil
		case queryparse.ConditionOpLt:
			return &KeyRange{
				Base:     packedParent,
				RangeEnd: packed,
			}, nil
		case queryparse.ConditionOpLte:
			return &KeyRange{
				Base:     packedParent,
				RangeEnd: tuple.Successor(packed),
			}, nil
		}

		// parse cond.Arg2
		arg2, err := parseJSONValue(avroType, cond.Arg2)
		if err != nil {
			return nil, err
		}

		// pack with arg2
		key[len(key)-1] = arg2
		packedEnd := key.Pack()

		// handle ops that also rely on cond.Arg2
		switch cond.Op {
		case queryparse.ConditionOpGtLt:
			return &KeyRange{
				Base:     tuple.Successor(packed),
				RangeEnd: packedEnd,
			}, nil
		case queryparse.ConditionOpGtLte:
			return &KeyRange{
				Base:     tuple.Successor(packed),
				RangeEnd: tuple.Successor(packedEnd),
			}, nil
		case queryparse.ConditionOpGteLt:
			return &KeyRange{
				Base:     packed,
				RangeEnd: packedEnd,
			}, nil
		case queryparse.ConditionOpGteLte:
			return &KeyRange{
				Base:     packed,
				RangeEnd: tuple.Successor(packedEnd),
			}, nil
		}

		// if we got here, something went terribly wrong
		log.Panicf("RangeFromQuery: impossible state")
		return nil, nil
	}

	// if we got here, something went terribly wrong
	log.Panicf("RangeFromQuery: impossible state")
	return nil, nil
}

// converts a query arg to a native value
func parseJSONValue(avroType interface{}, val interface{}) (interface{}, error) {
	return jsonNativeToAvroNative(avroType, val, nil)
}

// canPrefixLookup returns true iff type is string, bytes or fixed
func canPrefixLookup(avroType interface{}) bool {
	switch t := avroType.(type) {
	case string:
		return t == "string" || t == "bytes" || t == "fixed"
	case map[string]interface{}:
		return canPrefixLookup(t["type"])
	default:
		return false
	}
}
