package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSDL1(t *testing.T) {
	c := NewCompiler(`
		type TestExample1
			@stream(name: "test-example-1")
			@key(fields: ["a_aaa", "a_bbb"])
			@index(fields: ["a_bbb"], normalize: false)
		{
			a_aaa: String!
			a_bbb: Timestamp!
			a_ccc: [TestB!]
		}
		type TestB {
			b_aaa: Int
			b_bbb: Bytes
			b_ccc: TestC!
		}
		enum TestC {
			Aaa
			Bbb
			Ccc
		}
		type TestD {
			a_aaa: String!
			a_bbb: Timestamp!
			a_ccc: [TestExample1!]
		}
	`)
	err := c.Compile()
	assert.Nil(t, err)
}

func TestSDL2(t *testing.T) {
	err := NewCompiler(`
		type test {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "should start with an uppercase letter", err.Error())
}

func TestSDL3(t *testing.T) {
	err := NewCompiler(`
		type Test {
			a: Int!
		}
		type Test {
			b: Bool!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "has already been declared", err.Error())
}

func TestSDL4(t *testing.T) {
	err := NewCompiler(`
		type Test {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "no streams declared in input", err.Error())
}

func TestSDL5(t *testing.T) {
	err := NewCompiler(`
		type Test @stream {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "missing annotation '@key' with 'fields' arg in stream declaration at", err.Error())
}

func TestSDL6(t *testing.T) {
	err := NewCompiler(`
		type Test @stream(testA: "test", testB: "test") {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "unknown arg 'testA' for annotation '@stream'", err.Error())
}

func TestSDL9(t *testing.T) {
	err := NewCompiler(`
		type Test @stream @key(fields: [0, 1]) {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "arg 'fields' at .* is not a string or array of strings", err.Error())
}

func TestSDL10(t *testing.T) {
	err := NewCompiler(`
		type Test @stream @key(fields: ["a", "b"], external: whatever) {
			a: Int!
			b: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "parse error: .* unexpected \"whatever\" \\(expected .*\\)", err.Error())
}

func TestSDL11(t *testing.T) {
	err := NewCompiler(`
		type Test @stream @key(fields: true) {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "arg 'fields' at .* is not a string or array of strings", err.Error())
}

func TestSDL12(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "k") {
			k: Int!
			k: String
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'k' declared twice in type 'TestA'", err.Error())
}

func TestSDL13(t *testing.T) {
	err := NewCompiler(`
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
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "type 'TestA' is circular, which is not supported", err.Error())
}

func TestSDL14(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "k") {
			k: Int!
		}
		type TestB {
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "type 'TestB' does not define any fields", err.Error())
}

func TestSDL15(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "k") {
			k: Int!
			b: TestB
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "unknown type 'TestB'", err.Error())
}

func TestSDL16(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: ["a", "b"]) {
			a: Int!
			b: [Int!]
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'b' in type 'TestA' cannot be used as index", err.Error())
}

func TestSDL17(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: Int
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'a' in type 'TestA' cannot be used as index", err.Error())
}

func TestSDL18(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: TestB!
		}
		type TestB {
			b: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'a' in type 'TestA' cannot be used as index", err.Error())
}

func TestSDL19(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "b") {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "index field 'b' doesn't exist in type 'TestA'", err.Error())
}

func TestSDL20(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: ["a", "b"]) {
			a: Int!
			b: Int
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'b' in type 'TestA' cannot be used as index because it is optional", err.Error())
}

func TestSDL21(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: ["a", "a"]) {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'a' used twice in index for type 'TestA'", err.Error())
}

func TestSDL22(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: Int!
		}
		enum TestE {
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "enum 'TestE' must have at least one member", err.Error())
}

func TestSDL23(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: Int!
		}
		enum TestE {
			TestA
			TestA
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "member 'TestA' declared twice in enum 'TestE'", err.Error())
}

func TestSDL24(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: Int!
			b: [Int]!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "type wrapped by list cannot be optional at .*", err.Error())
}

func TestSDL25(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: Int!
		}
		enum Bytes20 {
			Aa
			Bb
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "declaration of 'Bytes20' at .* overlaps with primitive type name", err.Error())
}

func TestSDL26(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: Int!
			b: [[Int!]!]
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "nested lists are not allowed at .*", err.Error())
}

func TestSDL27(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: Int!
			bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field name 'b+' exceeds limit of 127 characters", err.Error())
}

func TestSDL28(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: Int!
			__timestamp: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field name '__timestamp' is a reserved identifier", err.Error())
}

func TestSDL29(t *testing.T) {
	c := NewCompiler(`
		type TestA @stream(name: "test") @key(fields: "a") {
			a: Int!
			b: Int!
		}
	`)
	err := c.Compile()
	assert.Nil(t, err)
	assert.Equal(t, "test", c.GetStream().Name)
}

func TestSDL30(t *testing.T) {
	c := NewCompiler(`
		type TestA @stream(name: "test") @key(fields: "a") {
			a: Int!
		}
		type TestB @stream(name: "test") @key(fields: "a") {
			a: Int!
		}
	`)
	err := c.Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "stream name 'test' used twice", err.Error())
}

func TestSDL31(t *testing.T) {
	c := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: Int!
		}
		type TestB @stream @key(fields: "a") {
			a: Int!
		}
	`)
	err := c.Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "more than one schema declared in input", err.Error())
}

func TestSDL32(t *testing.T) {
	c := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: Int!
			aA: String!
		}
	`)
	err := c.Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field name 'aA' in type 'TestA' is not underscore case", err.Error())
}

func TestSDL33(t *testing.T) {
	c := NewCompiler(`
		type TestA
			@stream
			@key(fields: "a")
			@index(fields: ["a", "b"])
		{
			a: Int!
			b: String!
		}
	`)
	err := c.Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "the indexes on type 'TestA' are not mutually exclusive", err.Error())
}

func TestSDL34(t *testing.T) {
	c := NewCompiler(`
		type TestA
			@stream
			@key(fields: "a")
			@index(fields: ["a", "b"])
		{
			a: Int!
			b: String!
		}
	`)
	err := c.Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "the indexes on type 'TestA' are not mutually exclusive", err.Error())
}

func TestSDL35(t *testing.T) {
	c := NewCompiler(`
		type TestA
			@stream
			@key(fields: "a")
			@hello
		{
			a: Int!
			b: String!
		}
	`)
	err := c.Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "unknown annotation '@hello' at ", err.Error())
}

func TestSDL36(t *testing.T) {
	c := NewCompiler(`
		type TestA @stream @key(fields: "a") {
			a: Int!
			b: String!
		}
		type TestB @key(fields: "b") {
			b: String!
		}
	`)
	err := c.Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "cannot have @key or @index annotations on non-stream declaration at", err.Error())
}

func TestSDL37(t *testing.T) {
	c := NewCompiler(`
		type TestA @stream(name: "") @key(fields: "a") {
			a: Int!
			b: String!
		}
	`)
	err := c.Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "stream arg 'name' at .* must be a non-empty string", err.Error())
}

func TestSDL38(t *testing.T) {
	c := NewCompiler(`
		type TestA @stream(name: "Hey") @key(fields: "a") {
			a: Int!
			b: String!
		}
	`)
	err := c.Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "stream name 'Hey' at .* is not a valid stream name", err.Error())
}

func TestSDL39(t *testing.T) {
	c := NewCompiler(`
		type TestA @stream @key(fields: "a", normalize: "true") {
			a: Int!
			b: String!
		}
	`)
	err := c.Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "key arg 'normalize' at .* must be a boolean", err.Error())
}

func TestSDL40(t *testing.T) {
	c := NewCompiler(`
		type TestA @stream @key(fields: "a", xxx: "true") {
			a: Int!
			b: String!
		}
	`)
	err := c.Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "unknown arg 'xxx' for annotation '@key' at", err.Error())
}
