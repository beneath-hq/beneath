package whereparser

import (
	"encoding/json"
	"sync"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/participle/v2/lexer/stateful"
)

var global *participle.Parser
var globalOnce sync.Once

func Parse(where string) (*AST, error) {
	globalOnce.Do(func() {
		global = newParser()
	})

	ast := &AST{}
	err := global.ParseString("", where, ast)
	if err != nil {
		return nil, err
	}

	return ast, nil
}

func newParser() *participle.Parser {
	whereLexer := lexer.Must(stateful.NewSimple([]stateful.Rule{
		{Name: `Keyword`, Pattern: `(?i)TRUE|FALSE|NULL|IS|AND|STARTS WITH`},
		{Name: `Ident`, Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: `Number`, Pattern: `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
		{Name: `String`, Pattern: `'[^']*'|"[^"]*"`},
		{Name: `Operator`, Pattern: `<>|!=|<=|>=|==|&&|[,()=<>&]`},
		{Name: `whitespace`, Pattern: `\s+`},
	}))
	return participle.MustBuild(
		&AST{},
		participle.Lexer(whereLexer),
		participle.Unquote("String"),
		participle.Upper("Keyword"),
		participle.CaseInsensitive("Keyword"),
	)
}

type AST struct {
	Conditions []*Condition `(@@ | ("(" @@ ")")) ( ("AND" | "," | "&" | "&&") (@@ | ("(" @@ ")")) )*`
}

type Condition struct {
	Field string `@Ident`
	Op    string `@( "IS" | "STARTS WITH" | "<=" | ">=" | "=" | "==" | "<" | ">" )`
	Value Value  `@@`
}

type Value struct {
	Number  *json.Number `(  @Number`
	String  *string      ` | @String`
	Boolean *Boolean     ` | @("TRUE" | "FALSE")`
	Null    bool         ` | @"NULL" )`
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "TRUE"
	return nil
}
