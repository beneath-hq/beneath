package transpilers

import (
	"fmt"

	"github.com/beneath-hq/beneath/pkg/schemalang"
	"github.com/beneath-hq/beneath/pkg/schemalang/graphql"
)

// FromGraphQL converts a GraphQL schema to an Avro schema
func FromGraphQL(gql string) (schemalang.Schema, schemalang.Indexes, error) {
	// parse ast
	parsed, err := graphql.ParseSDL(gql)
	if err != nil {
		return nil, nil, err
	}

	// transpile to record
	t := fromGraphQL{
		gqlSchema:    parsed,
		definedNames: make(map[string]bool),
	}

	// transpile
	res, err := t.fromType(parsed.Root.Type) // root is always a Type
	if err != nil {
		return nil, nil, err
	}

	// add table info
	indexes := make(schemalang.Indexes, len(parsed.Indexes)+1)
	indexes[0] = t.fromIndex(parsed.Key, true)
	for i, index := range parsed.Indexes {
		indexes[i+1] = t.fromIndex(index, false)
	}

	return res, indexes, nil
}

type fromGraphQL struct {
	gqlSchema    *graphql.Schema
	definedNames map[string]bool
}

func (t fromGraphQL) fromIndex(index graphql.Index, key bool) schemalang.Index {
	return schemalang.Index{
		Fields:    index.Fields,
		Key:       key,
		Normalize: index.Normalize,
	}
}

func (t fromGraphQL) fromDeclaration(ast *graphql.Declaration) (schemalang.Schema, error) {
	if ast.Enum != nil {
		return t.fromEnum(ast.Enum)
	}

	if ast.Type != nil {
		return t.fromType(ast.Type)
	}

	panic(fmt.Errorf("declaration for type '%v' is neither enum nor record", ast))
}

func (t fromGraphQL) fromEnum(e *graphql.Enum) (schemalang.Schema, error) {
	name := e.Name

	if t.definedNames[name] {
		return &schemalang.Ref{Name: name}, nil
	}
	t.definedNames[name] = true

	return &schemalang.Enum{
		Name:    name,
		Doc:     e.Doc,
		Symbols: e.Members,
	}, nil
}

func (t fromGraphQL) fromType(ast *graphql.Type) (schemalang.Schema, error) {
	name := ast.Name

	if t.definedNames[name] {
		return &schemalang.Ref{Name: name}, nil
	}
	t.definedNames[name] = true

	fields := make([]*schemalang.RecordField, len(ast.Fields))

	for idx, gqlField := range ast.Fields {
		fieldSchema, err := t.fromTypeRef(gqlField.Type)
		if err != nil {
			return nil, err
		}

		fields[idx] = &schemalang.RecordField{
			Name: gqlField.Name,
			Doc:  gqlField.Doc,
			Type: fieldSchema,
		}
	}

	return &schemalang.Record{
		Name:   name,
		Doc:    ast.Doc,
		Fields: fields,
	}, nil
}

func (t fromGraphQL) fromTypeRef(ast *graphql.TypeRef) (schemalang.Schema, error) {
	var res schemalang.Schema
	if ast.Array != nil {
		itemType, err := t.fromTypeRef(ast.Array)
		if err != nil {
			return nil, err
		}
		res = &schemalang.Array{ItemType: itemType}
	} else if ast.Type != "" {
		typ, err := t.fromTypeName(ast.Type)
		if err != nil {
			return nil, err
		}
		res = typ
	}

	if !ast.Required {
		res = &schemalang.Nullable{NonNullType: res}
	}

	return res, nil
}

func (t fromGraphQL) fromTypeName(name string) (schemalang.Schema, error) {
	primitive, arg, err := graphql.ParsePrimitive(name)

	if err == graphql.ErrNotPrimitive {
		declaration := t.gqlSchema.Declarations[name]
		if declaration == nil {
			return nil, fmt.Errorf("unknown type '%s' (neither a primitive nor a declared type)", name)
		}
		return t.fromDeclaration(declaration)
	}

	switch primitive {
	case graphql.BooleanPrimitiveType:
		return &schemalang.Primitive{Type: schemalang.BooleanType}, nil
	case graphql.BytesPrimitiveType:
		if arg != 0 {
			return &schemalang.Fixed{Size: arg}, nil
		}
		return &schemalang.Primitive{Type: schemalang.BytesType}, nil
	case graphql.FloatPrimitiveType:
		if arg == 32 {
			return &schemalang.Primitive{Type: schemalang.FloatType}, nil
		}
		return &schemalang.Primitive{Type: schemalang.DoubleType}, nil
	case graphql.IntPrimitiveType:
		if arg == 32 {
			return &schemalang.Primitive{Type: schemalang.IntType}, nil
		}
		return &schemalang.Primitive{Type: schemalang.LongType}, nil
	case graphql.NumericPrimitiveType:
		return &schemalang.Primitive{
			Type:        schemalang.BytesType,
			LogicalType: schemalang.NumericLogicalType,
		}, nil
	case graphql.StringPrimitiveType:
		return &schemalang.Primitive{Type: schemalang.StringType}, nil
	case graphql.TimestampPrimitiveType:
		return &schemalang.Primitive{
			Type:        schemalang.LongType,
			LogicalType: schemalang.TimestampMillisLogicalType,
		}, nil
	case graphql.UUIDPrimitiveType:
		return &schemalang.Primitive{
			Type:        schemalang.StringType,
			LogicalType: schemalang.UUIDLogicalType,
		}, nil
	default:
		panic(fmt.Errorf("unhandled primitive type '%s'", primitive))
	}
}
