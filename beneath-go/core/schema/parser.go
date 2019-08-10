package schema

import (
	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"
)

var globalParser *participle.Parser

// GetParser returns a concurrency-safe global SDL parser
func GetParser() *participle.Parser {
	if globalParser == nil {
		globalParser = NewParser()
	}
	return globalParser
}

// NewParser creates a new SDL parser
func NewParser() *participle.Parser {
	sdlLexer := lexer.Must(ebnf.New(`
		Comment = ("#" | "//") { "\u0000"…"\uffff"-"\n"-"\r" } .
		String = 
				("'" { "\u0000"…"\uffff"-"\\"-"'" | "\\" any } "'")
			| ("\"" { "\u0000"…"\uffff"-"\\"-"\"" | "\\" any } "\"")
			.
		Bool = "true" | "false" .
		Ident = (alpha | "_") { "_" | alpha | digit } .
		Number = [ "-" | "+" ] ("." | digit) { "." | digit } .
		Punct = "!"…"/" | ":"…"@" | "["…` + "\"`\"" + ` | "{"…"~" .
		Whitespace = " " | "\t" | "\n" | "\r" .

		alpha = "a"…"z" | "A"…"Z" .
		digit = "0"…"9" .
		any = "\u0000"…"\uffff" .
	`))

	return participle.MustBuild(&File{},
		participle.Lexer(sdlLexer),
		participle.Elide("Comment", "Whitespace"),
		participle.Unquote("String"),
	)
}

// File is top-level input
type File struct {
	Declarations []*Declaration `@@*`
}

// Declaration is a "type ..." or "enum ..."
type Declaration struct {
	Pos  lexer.Position
	Enum *Enum `  @@`
	Type *Type `| @@`
}

// Enum declaration
type Enum struct {
	Name    string   `"enum" @Ident`
	Members []string `"{" @Ident* "}"`
}

// Type declaration
type Type struct {
	Doc         string        "@String?"
	Name        string        `"type" @Ident`
	Annotations []*Annotation `@@*`
	Fields      []*Field      `"{" @@* "}"`
}

// Field is member of Type
type Field struct {
	Doc  string   "@String?"
	Name string   `@Ident`
	Type *TypeRef `":" @@`
}

// Annotation on Type
type Annotation struct {
	Name      string      `"@" @Ident`
	Arguments []*Argument `("(" (@@ ("," @@)*)? ")")?`
}

// Argument to an Annotation
type Argument struct {
	Pos   lexer.Position
	Name  string `@Ident`
	Value *Value `":" @@`
}

// TypeRef is type of Field
type TypeRef struct {
	Pos      lexer.Position
	Array    *TypeRef `(   "[" @@ "]"`
	Type     string   `  | @Ident )`
	Required bool     `@"!"?`
}

// Value is passed to an Argument in an Annotation
type Value struct {
	String string   "  @String"
	Number float64  "| @Number"
	Symbol string   `| @Bool`
	Array  []*Value `| "[" @@ ("," @@)* "]"`
}
