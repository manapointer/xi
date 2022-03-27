package token

import (
	"strconv"
)

type TokenType int

const (
	Error TokenType = iota
	Illegal
	Eof
	Comment

	If
	Else
	While
	Return
	Length
	Use
	Int
	Bool
	True
	False
	Underscore

	Ident
	Integer
	Char
	String

	Add
	Sub
	Mul
	Div
	Rem

	Assign
	Not
	Eq
	Neq
	Lt
	Le
	Gt
	Ge
	And
	Or

	Lparen
	Lbrack
	Lbrace
	Rparen
	Rbrack
	Rbrace

	Comma
	Colon
	Semicolon
)

var tokens = [...]string{
	Error:   "ERROR",
	Illegal: "ILLEGAL",
	Eof:     "EOF",
	Comment: "COMMENT",

	If:         "if",
	Else:       "else",
	While:      "while",
	Return:     "return",
	Length:     "length",
	Use:        "use",
	Int:        "int",
	Bool:       "bool",
	True:       "true",
	False:      "false",
	Underscore: "_",

	Ident:   "IDENT",
	Integer: "INTEGER",
	Char:    "CHAR",
	String:  "STRING",

	Add: "+",
	Sub: "-",
	Mul: "*",
	Div: "/",
	Rem: "%",

	Assign: "=",
	Not:    "!",
	Eq:     "==",
	Neq:    "!=",
	Lt:     "<",
	Le:     "<=",
	Gt:     ">",
	Ge:     ">=",
	And:    "&",
	Or:     "|",

	Lparen: "(",
	Lbrack: "[",
	Lbrace: "{",
	Rparen: ")",
	Rbrack: "]",
	Rbrace: "}",

	Comma:     ",",
	Colon:     ":",
	Semicolon: ";",
}

func (typ TokenType) String() string {
	if Illegal <= typ && typ <= Colon {
		return tokens[typ]
	}

	return "Token(" + strconv.Itoa(int(typ)) + ")"
}

type Token struct {
	Typ TokenType
	Pos Position
	Lit string
}
