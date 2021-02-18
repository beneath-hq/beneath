package codec

import (
	"bytes"
	"errors"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/pkg/codec/ext/tuple"
	"gitlab.com/beneath-hq/beneath/pkg/queryparse"
	"gitlab.com/beneath-hq/beneath/pkg/schemalang"
)

var (
	// ErrIndexMiss is returned by newKeyRange when a query doesn't match the index fields
	ErrIndexMiss = errors.New("you can only query indexed fields (composite keys are indexed starting with the leftmost field)")
)

// KeyRange represents a range of keys
type KeyRange struct {
	Base     []byte
	RangeEnd []byte
}

// IsNil returns true if the key range is uninitialized
func (r KeyRange) IsNil() bool {
	return r.Base == nil && r.RangeEnd == nil
}

// IsPrefix is true iff key range identifies a prefix of keys
func (r KeyRange) IsPrefix() bool {
	if len(r.RangeEnd) > len(r.Base) {
		return false
	}
	// approach: moving backwards, if Base is 0xFF, then RangeEnd should be empty, else RangeEnd should be Base[idx]+1
	for i := len(r.Base) - 1; i >= 0; i-- {
		if r.Base[i] == 0xFF {
			if len(r.RangeEnd)-1 >= i {
				return false
			}
			continue
		}
		if len(r.RangeEnd)-1 == i {
			if r.Base[i]+1 == r.RangeEnd[i] {
				return true
			}
		}
		return false
	}
	return false
}

// Contains is true if the KeyRange contains a key
func (r KeyRange) Contains(key []byte) bool {
	if r.RangeEnd == nil {
		return bytes.Compare(key, r.Base) >= 0
	}
	return bytes.Compare(key, r.Base) >= 0 && bytes.Compare(key, r.RangeEnd) < 0
}

// newKeyRange builds a new key range based on a where query and a key codec
func newKeyRange(c *Codec, index Index, q queryparse.Query) (r KeyRange, err error) {
	// handle empty
	if q == nil || len(q) == 0 {
		return KeyRange{}, nil
	}

	// prepare key
	key := make(tuple.Tuple, len(q))

	// iterate through key fields
	for idx, field := range index.GetFields() {
		// get condition for field
		cond := q[field]
		if cond == nil {
			if len(index.GetFields()) == 1 {
				return KeyRange{}, ErrIndexMiss
			}
			return KeyRange{}, ErrIndexMiss
		}

		// get avro type
		fieldType := c.getFieldTypes()[field].Type

		// parse arg1 value
		arg1, err := parseJSONValue(fieldType, cond.Arg1)
		if err != nil {
			return KeyRange{}, err
		}

		// convert UUID because package tuple has its own UUID type
		if uuidVal, ok := arg1.(uuid.UUID); ok {
			arg1 = tuple.UUID(uuidVal)
		}

		// set val in key we're building
		key[idx] = arg1

		// _eq is the only condition that is potentially not terminal
		// so handle it separately
		if cond.Op == queryparse.ConditionOpEq {
			// next step depends on how far we've come
			if idx+1 == len(index.GetFields()) {
				// we've added _eq constraints for every key field, so we're done
				// if it's the primary index, it identifies a single row, but for secondary indexes, it's still only a prefix
				packed := key.Pack()
				return KeyRange{
					Base:     packed,
					RangeEnd: tuple.PrefixSuccessor(packed),
				}, nil
			} else if idx+1 == len(q) {
				// we've added _eq constraints for some subset of leftmost key fields, so we're done and it's a prefix key
				packed := key.Pack()
				return KeyRange{
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
			return KeyRange{}, ErrIndexMiss
		}

		// pack key (and a copy with the last element removed)
		packed := key.Pack()
		packedParent := key[:len(key)-1].Pack()

		// handle ops that only rely on cond.Arg1 (which we've already parsed)
		switch cond.Op {
		case queryparse.ConditionOpPrefix:
			// prefix should only work on strings, bytes and fixed
			if !canPrefixLookup(fieldType) {
				return KeyRange{}, fmt.Errorf("cannot use '_prefix' on field '%s' because it only works on string and byte types", field)
			}
			base := tuple.TruncateBytesTypeForPrefixSuccessor(packed)
			return KeyRange{
				Base:     base,
				RangeEnd: tuple.PrefixSuccessor(base),
			}, nil
		case queryparse.ConditionOpGt:
			return KeyRange{
				Base:     tuple.PrefixSuccessor(packed),
				RangeEnd: tuple.PrefixSuccessor(packedParent),
			}, nil
		case queryparse.ConditionOpGte:
			return KeyRange{
				Base:     packed,
				RangeEnd: tuple.PrefixSuccessor(packedParent),
			}, nil
		case queryparse.ConditionOpLt:
			return KeyRange{
				Base:     packedParent,
				RangeEnd: packed,
			}, nil
		case queryparse.ConditionOpLte:
			return KeyRange{
				Base:     packedParent,
				RangeEnd: tuple.PrefixSuccessor(packed),
			}, nil
		}

		// parse cond.Arg2
		arg2, err := parseJSONValue(fieldType, cond.Arg2)
		if err != nil {
			return KeyRange{}, err
		}

		// pack with arg2
		key[len(key)-1] = arg2
		packedEnd := key.Pack()

		// handle ops that also rely on cond.Arg2
		switch cond.Op {
		case queryparse.ConditionOpGtLt:
			return KeyRange{
				Base:     tuple.PrefixSuccessor(packed),
				RangeEnd: packedEnd,
			}, nil
		case queryparse.ConditionOpGtLte:
			return KeyRange{
				Base:     tuple.PrefixSuccessor(packed),
				RangeEnd: tuple.PrefixSuccessor(packedEnd),
			}, nil
		case queryparse.ConditionOpGteLt:
			return KeyRange{
				Base:     packed,
				RangeEnd: packedEnd,
			}, nil
		case queryparse.ConditionOpGteLte:
			return KeyRange{
				Base:     packed,
				RangeEnd: tuple.PrefixSuccessor(packedEnd),
			}, nil
		}

		// if we got here, something went terribly wrong
		panic(fmt.Errorf("RangeFromQuery: impossible state"))
	}

	// if we got here, something went terribly wrong
	panic(fmt.Errorf("RangeFromQuery: impossible state"))
}

// converts a query arg to a native value
func parseJSONValue(schema schemalang.Schema, val interface{}) (interface{}, error) {
	return jsonConverter{}.convert(schema, val, true)
}

// canPrefixLookup returns true iff type is string, bytes or fixed
func canPrefixLookup(schema schemalang.Schema) bool {
	t := schema.GetType()
	if t == schemalang.StringType || t == schemalang.BytesType || t == schemalang.FixedType {
		return true
	}
	return false
}
