package scanner

import (
	"testing"

	"github.com/manapointer/xi/pkg/token"
)

type scannerTest struct {
	name   string
	input  string
	tokens []token.Token
}

func makeToken(typ token.TokenType, text string) token.Token {
	return token.Token{
		Typ: typ,
		Lit: text,
	}
}

var (
	tokLparen = makeToken(token.Lparen, "(")
	tokRparen = makeToken(token.Rparen, ")")
	tokSub    = makeToken(token.Sub, "-")
	tokAdd    = makeToken(token.Add, "+")
	tokDiv    = makeToken(token.Div, "/")
	tokRem    = makeToken(token.Rem, "%")
	tokEq     = makeToken(token.Eq, "==")
	tokBangEq = makeToken(token.Neq, "!=")
	tokLt     = makeToken(token.Lt, "<")
	tokLe     = makeToken(token.Le, "<=")
	tokGt     = makeToken(token.Gt, ">")
	tokGe     = makeToken(token.Ge, ">=")
	tokEof    = makeToken(token.Eof, "")
	tokColon  = makeToken(token.Colon, ":")
	tokAssign = makeToken(token.Assign, "=")
	tokLbrack = makeToken(token.Lbrack, "[")
	tokRbrack = makeToken(token.Rbrack, "]")
	tokLbrace = makeToken(token.Lbrace, "{")
	tokRbrace = makeToken(token.Rbrace, "}")
	tokWhile  = makeToken(token.While, "while")
	tokIf     = makeToken(token.If, "if")
	tokLength = makeToken(token.Length, "length")
	tokRet    = makeToken(token.Return, "return")
	tokInt    = makeToken(token.Int, "int")
)

var lexTests = []scannerTest{
	{"operators", "-+/%==!=<<=>>=", []token.Token{
		tokSub,
		tokAdd,
		tokDiv,
		tokRem,
		tokEq,
		tokBangEq,
		tokLt,
		tokLe,
		tokGt,
		tokGe,
		tokEof,
	}},
	{"number", "1337", []token.Token{
		makeToken(token.Integer, "1337"),
		tokEof,
	}},
	{"assign", "a: int = 4", []token.Token{
		makeToken(token.Ident, "a"),
		tokColon,
		tokInt,
		tokAssign,
		makeToken(token.Integer, "4"),
		tokEof,
	}},
	{"string", `a: int[] = "Hello, world"`, []token.Token{
		makeToken(token.Ident, "a"),
		tokColon,
		tokInt,
		tokLbrack,
		tokRbrack,
		tokAssign,
		makeToken(token.String, `"Hello, world"`),
		tokEof,
	}},
	{"escape", `a: int[] = "Hello, world\n"`, []token.Token{
		makeToken(token.Ident, "a"),
		tokColon,
		tokInt,
		tokLbrack,
		tokRbrack,
		tokAssign,
		makeToken(token.String, `"Hello, world\n"`),
		tokEof,
	}},
	{"unknown escape sequence", `"Hello, world\d"`, []token.Token{
		makeToken(token.Error, "unknown escape sequence"),
	}},
	{"unterminated string", "\"Hello, world\n", []token.Token{
		makeToken(token.Error, "string literal not terminated"),
	}},
	{"sort", `
sort(a: int[]) {
	i:int = 0
	n:int = length(a)
	while (i < n) {
		j:int = i
		while (j > 0) {
			if (a[j-1] > a[j]) {
				swap:int = a[j]
				a[j] = a[j-1]
				a[j-1] = swap
			}
			j = j-1
		}
		i = i+1
	}
}
`, []token.Token{
		makeToken(token.Ident, "sort"),
		tokLparen,
		makeToken(token.Ident, "a"),
		tokColon,
		tokInt,
		tokLbrack,
		tokRbrack,
		tokRparen,
		tokLbrace,
		makeToken(token.Ident, "i"),
		tokColon,
		tokInt,
		tokAssign,
		makeToken(token.Integer, "0"),
		makeToken(token.Ident, "n"),
		tokColon,
		tokInt,
		tokAssign,
		tokLength,
		tokLparen,
		makeToken(token.Ident, "a"),
		tokRparen,
		tokWhile,
		tokLparen,
		makeToken(token.Ident, "i"),
		tokLt,
		makeToken(token.Ident, "n"),
		tokRparen,
		tokLbrace,
		makeToken(token.Ident, "j"),
		tokColon,
		tokInt,
		tokAssign,
		makeToken(token.Ident, "i"),
		tokWhile,
		tokLparen,
		makeToken(token.Ident, "j"),
		tokGt,
		makeToken(token.Integer, "0"),
		tokRparen,
		tokLbrace,
		tokIf,
		tokLparen,
		makeToken(token.Ident, "a"),
		tokLbrack,
		makeToken(token.Ident, "j"),
		tokSub,
		makeToken(token.Integer, "1"),
		tokRbrack,
		tokGt,
		makeToken(token.Ident, "a"),
		tokLbrack,
		makeToken(token.Ident, "j"),
		tokRbrack,
		tokRparen,
		tokLbrace,
		makeToken(token.Ident, "swap"),
		tokColon,
		tokInt,
		tokAssign,
		makeToken(token.Ident, "a"),
		tokLbrack,
		makeToken(token.Ident, "j"),
		tokRbrack,
		makeToken(token.Ident, "a"),
		tokLbrack,
		makeToken(token.Ident, "j"),
		tokRbrack,
		tokAssign,
		makeToken(token.Ident, "a"),
		tokLbrack,
		makeToken(token.Ident, "j"),
		tokSub,
		makeToken(token.Integer, "1"),
		tokRbrack,
		makeToken(token.Ident, "a"),
		tokLbrack,
		makeToken(token.Ident, "j"),
		tokSub,
		makeToken(token.Integer, "1"),
		tokRbrack,
		tokAssign,
		makeToken(token.Ident, "swap"),
		tokRbrace,
		makeToken(token.Ident, "j"),
		tokAssign,
		makeToken(token.Ident, "j"),
		tokSub,
		makeToken(token.Integer, "1"),
		tokRbrace,
		makeToken(token.Ident, "i"),
		tokAssign,
		makeToken(token.Ident, "i"),
		tokAdd,
		makeToken(token.Integer, "1"),
		tokRbrace,
		tokRbrace,
		tokEof,
	}},
}

func TestScan(t *testing.T) {
	for _, test := range lexTests {
		tokens := runTest(test)
		if !tokensEqual(tokens, test.tokens) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, tokens, test.tokens)
		}
	}
}

func runTest(test scannerTest) (tokens []token.Token) {
	s := NewScanner([]byte(test.input), nil)

	for {
		tok := s.Scan()
		tokens = append(tokens, tok)
		if tok.Typ == token.Eof || tok.Typ == token.Error {
			return
		}
	}
}

func tokensEqual(t1, t2 []token.Token) bool {
	if len(t1) != len(t2) {
		return false
	}

	for i := range t1 {
		if t1[i].Typ != t2[i].Typ {
			return false
		}

		if t1[i].Lit != t2[i].Lit {
			return false
		}
	}

	return true
}
