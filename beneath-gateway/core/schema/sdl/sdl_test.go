package sdl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSDL1(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream(name: "test-example-1", key: ["aAaa", "aBbb"]) {
			aAaa: String!
			aBbb: Timestamp!
			aCcc: [TestB!]
		}
		type TestB {
			bAaa: Int
			bBbb: Bytes
			bCcc: TestC!
		}
		enum TestC {
			Aaa
			Bbb
			Ccc
		}
		type TestD {
			aAaa: String!
			aBbb: Timestamp!
			aCcc: [TestA!]
		}
	`).Compile()
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
	assert.Regexp(t, "missing arg 'name' in @stream", err.Error())
}

func TestSDL6(t *testing.T) {
	err := NewCompiler(`
		type Test @stream(testA: "test", testB: "test") {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "unknown @stream arg 'testA'", err.Error())
}

func TestSDL7(t *testing.T) {
	err := NewCompiler(`
		type Test @stream(name: 20) {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "stream arg 'name' at .* is not a string", err.Error())
}

func TestSDL8(t *testing.T) {
	err := NewCompiler(`
		type Test @stream(name: "test") {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "missing arg 'key' in @stream", err.Error())
}

func TestSDL9(t *testing.T) {
	err := NewCompiler(`
		type Test @stream(name: "test", key: [0, 1]) {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "stream arg 'key' at .* is not a string or array of strings", err.Error())
}

func TestSDL10(t *testing.T) {
	err := NewCompiler(`
		type Test @stream(name: "test", key: ["a", "b"], external: whatever) {
			a: Int!
			b: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "parse error: .* unexpected \"whatever\" \\(expected .*\\)", err.Error())
}

func TestSDL11(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream(name: "test", key: "a") {
			a: Int!
		}
		type TestB @stream(name: "test", key: "a") {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "stream name 'test' used twice", err.Error())
}

func TestSDL12(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream(name: "test", key: "k") {
			k: Int!
			k: String
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'k' declared twice in type 'TestA'", err.Error())
}

func TestSDL13(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream(name: "test", key: "k") {
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
		type TestA @stream(name: "test", key: "k") {
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
		type TestA @stream(name: "test", key: "k") {
			k: Int!
			b: TestB
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "unknown type 'TestB'", err.Error())
}

func TestSDL16(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream(name: "test", key: ["a", "b"]) {
			a: Int!
			b: [Int!]
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'b' in type 'TestA' cannot be used as key", err.Error())
}

func TestSDL17(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream(name: "test", key: "a") {
			a: Int
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'a' in type 'TestA' cannot be used as key", err.Error())
}

func TestSDL18(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream(name: "test", key: "a") {
			a: TestB!
		}
		type TestB {
			b: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'a' in type 'TestA' cannot be used as key", err.Error())
}

func TestSDL19(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream(name: "test", key: "b") {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'b' in key doesn't exist in type 'TestA'", err.Error())
}

func TestSDL20(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream(name: "test", key: ["a", "b"]) {
			a: Int!
			b: Int
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'b' in type 'TestA' cannot be used as key because it is optional", err.Error())
}

func TestSDL21(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream(name: "test", key: ["a", "a"]) {
			a: Int!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "field 'a' used twice in key for type 'TestA'", err.Error())
}

func TestSDL22(t *testing.T) {
	err := NewCompiler(`
		type TestA @stream(name: "test", key: "a") {
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
		type TestA @stream(name: "test", key: "a") {
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
		type TestA @stream(name: "test", key: "a") {
			a: Int!
			b: [Int]!
		}
	`).Compile()
	assert.NotNil(t, err)
	assert.Regexp(t, "type wrapped by list cannot be optional at .*", err.Error())
}
