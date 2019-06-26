package sdl

import (
	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"
)

var globalParser *participle.Parser

func GetParser() *participle.Parser {
	if globalParser == nil {
		globalParser = NewParser()
	}
	return globalParser
}

func NewParser() *participle.Parser {
	sdlLexer := lexer.Must(ebnf.New(`
		Comment = ("#" | "//") { "\u0000"…"\uffff"-"\n"-"\r" } .
		String = 
				("'" { "\u0000"…"\uffff"-"\\"-"'" | "\\" any } "'")
			| ("\"" { "\u0000"…"\uffff"-"\\"-"\"" | "\\" any } "\"")
			.
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

type File struct {
	Declarations []*Declaration `@@*`
}

type Declaration struct {
	Pos  lexer.Position
	Enum *Enum `  @@`
	Type *Type `| @@`
}

type Enum struct {
	Name    string   `"enum" @Ident`
	Members []string `"{" @Ident* "}"`
}

type Type struct {
	Doc         string        "@String?"
	Name        string        `"type" @Ident`
	Annotations []*Annotation `@@*`
	Fields      []*Field      `"{" @@* "}"`
}

type Field struct {
	Name string   `@Ident`
	Type *TypeRef `":" @@`
}

type Annotation struct {
	Name      string      `"@" @Ident`
	Arguments []*Argument `("(" (@@ ("," @@)*)? ")")?`
}

type Argument struct {
	Pos   lexer.Position
	Name  string `@Ident`
	Value *Value `":" @@`
}

type TypeRef struct {
	Pos      lexer.Position
	Array    *TypeRef `(   "[" @@ "]"`
	Type     string   `  | @Ident )`
	Required bool     `@"!"?`
}

type Value struct {
	String string   "  @String"
	Number float64  "| @Number"
	Symbol string   `| @Ident`
	Array  []*Value `| "[" @@ ("," @@)* "]"`
}
