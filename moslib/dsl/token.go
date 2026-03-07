package dsl

import "fmt"

// TokenType classifies lexer tokens.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenString
	TokenInteger
	TokenFloat
	TokenBool
	TokenDateTime
	TokenLBrace
	TokenRBrace
	TokenLBracket
	TokenRBracket
	TokenEquals
	TokenComma
	TokenComment
	TokenNewline

	// v3 keywords (all lowercase, brace-delimited)
	TokenFeature
	TokenBackground
	TokenScenario
	TokenGiven
	TokenWhen
	TokenThen
	TokenGroup
	TokenSpec
	TokenInclude
	TokenStepText
)

var tokenNames = map[TokenType]string{
	TokenEOF:        "EOF",
	TokenIdent:      "Ident",
	TokenString:     "String",
	TokenInteger:    "Integer",
	TokenFloat:      "Float",
	TokenBool:       "Bool",
	TokenDateTime:   "DateTime",
	TokenLBrace:     "{",
	TokenRBrace:     "}",
	TokenLBracket:   "[",
	TokenRBracket:   "]",
	TokenEquals:     "=",
	TokenComma:      ",",
	TokenComment:    "Comment",
	TokenNewline:    "Newline",
	TokenFeature:    "feature",
	TokenBackground: "background",
	TokenScenario:   "scenario",
	TokenGiven:      "given",
	TokenWhen:       "when",
	TokenThen:       "then",
	TokenGroup:      "group",
	TokenSpec:       "spec",
	TokenInclude:    "include",
	TokenStepText:   "StepText",
}

func (t TokenType) String() string {
	if s, ok := tokenNames[t]; ok {
		return s
	}
	return fmt.Sprintf("Token(%d)", int(t))
}

// Token is a single lexical unit.
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Col    int
	Offset int // byte offset in source
}
