package transpilers

import (
	"testing"

	"github.com/beneath-hq/beneath/pkg/schemalang"
	"github.com/stretchr/testify/assert"
)

const UserSchemaGraphQLWithDoc = `
" User test "
type user @stream @key(fields: "id") {
	id: UUID!
	password_hash: Bytes32!
	password_salt: Bytes32!
	favorite_number: Int32!
	favorite_color: String
	home: coords!
	work: coords!
}

" Coords test "
type coords {
	latitude: Float!
	longitude: Float!
}
`

func TestGraphQL(t *testing.T) {
	schema, spec, err := FromGraphQL(UserSchemaGraphQLWithDoc)
	assert.Nil(t, err)

	err = schemalang.Check(schema)
	assert.Nil(t, err)

	err = spec.Check(schema)
	assert.Nil(t, err)

	avro := ToAvro(schema, true)
	assert.Equal(t, UserSchemaAvroWithDoc, avro)
}

func TestChecksWithGraphQL(t *testing.T) {
	parseCheckAndAssertRegex := func(t *testing.T, gql string, regex string) {
		schema, spec, err := FromGraphQL(gql)
		if err == nil {
			err = schemalang.Check(schema)
			if err == nil {
				err = spec.Check(schema)
			}
		}
		assert.NotNil(t, err)
		assert.Regexp(t, regex, err.Error())
	}

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "a") {
			a: String!
			k: Int!
			k: String
		}
	`, "field 'k' declared twice in")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "k") {
			k: Int!
			a: TestB!
		}
		
		type TestB {
			b: TestC
		}
		
		type TestC {
			c: TestA
		}
	`, "found cyclic type 'TestA' \\(not supported\\)")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "k") {
			k: Int!
			a: TestB
		}
		
		type TestB {
		}
	`, "type 'TestB' does not define any fields")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "k") {
			k: Int!
			b: TestB
		}
	`, "unknown type 'TestB'")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: ["a", "b"]) {
			a: Int!
			b: [Int!]
		}
	`, "field 'b' in type 'TestA' cannot be used as index")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "a") {
			a: Int
		}
	`, "field 'a' in type 'TestA' cannot be used as index")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "a") {
			a: TestB!
		}

		type TestB {
			b: Int!
		}
	`, "field 'a' in type 'TestA' cannot be used as index")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "b") {
			a: Int!
		}
	`, "index field 'b' doesn't exist in type 'TestA'")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: ["a", "b"]) {
			a: Int!
			b: Int
		}
	`, "field 'b' in type 'TestA' cannot be used as index because it is optional")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: ["a", "a"]) {
			a: Int!
		}
	`, "field 'a' used twice in index for type 'TestA'")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "a") {
			a: Int!
			b: TestE
		}

		enum TestE {
		}
	`, "enum 'TestE' must have at least one symbol")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "a") {
			a: Int!
			b: TestE
		}
		
		enum TestE {
			TestA
			TestA
		}
	`, "symbol 'TestA' declared twice in enum 'TestE'")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "a") {
			a: Int!
			b: [Int]!
		}
	`, "type wrapped by list cannot be nullable")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "a") {
			a: Int!
			b: [[Int!]!]
		}
	`, "nested lists are not allowed")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "a") {
			a: Int!
			bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb: Int!
		}
	`, "field name 'b+' exceeds limit of 64 characters")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "a") {
			a: Int!
			__timestamp: Int!
		}
	`, "field name '__timestamp' is a reserved identifier")

	parseCheckAndAssertRegex(t, `
		type TestA @stream @key(fields: "a") {
			a: Int!
			aA: String!
		}
	`, "field name 'aA' in record 'TestA' is not underscore case")

	parseCheckAndAssertRegex(t, `
		type TestA
			@stream
			@key(fields: "a")
			@index(fields: ["a", "b"])
		{
			a: Int!
			b: String!
		}
	`, "the indexes on type 'TestA' are not mutually exclusive")

	parseCheckAndAssertRegex(t, `
		type TestA
			@stream
			@key(fields: "a")
			@index(fields: ["a", "b"])
		{
			a: Int!
			b: String!
		}
	`, "the indexes on type 'TestA' are not mutually exclusive")

	parseCheckAndAssertRegex(t, `
		type TestA @stream(name: "Hey") @key(fields: "a") {
			a: Int!
			b: String!
		}
	`, "stream name 'Hey' at .* is not a valid stream name")
}
