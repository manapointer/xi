package scanner

import (
	"fmt"
	"unicode/utf8"

	"github.com/manapointer/xi/pkg/token"
)

type (
	state func(*Scanner) state

	ErrorHandler func(pos token.Position, message string)

	Scanner struct {
		src    []byte
		tokens chan token.Token
		err    ErrorHandler

		ch      rune // current character
		pos     int  // character position
		rpos    int  // next read position
		start   int  // start of next token
		line    int  // scanner's line number
		linepos int  // start of the current line; used to calculate pos column

		ErrorCount int
	}
)

const (
	eof = -1
)

func (s *Scanner) next() {
	if s.rpos < len(s.src) {
		s.pos = s.rpos

		r, w := rune(s.src[s.rpos]), 1
		switch {
		case r == 0:
			// TODO: error handling
		case r >= utf8.RuneSelf:
			r, w = utf8.DecodeRune(s.src[s.rpos:])
			if r == utf8.RuneError && w == 1 {
				// TODO: error handling
			}

		}
		s.rpos += w
		s.ch = r
	} else {
		s.pos = len(s.src)
		s.ch = eof
	}
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAlpha(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isHexDigit(r rune) bool {
	return ('A' <= r && r <= 'F') || ('a' <= r && r < 'f') || isDigit(r)
}

func (s *Scanner) bump() {
	s.start++
}

func (s *Scanner) switch2(tok0, tok1 token.TokenType) token.TokenType {
	if s.ch == '=' {
		s.next()
		return tok1
	}

	return tok0
}

func (s *Scanner) scanEscape(quote rune) bool {
	s.next()
	switch s.ch {
	case '\\', quote, 'n', 't':
		s.next()
	case 'x':
		s.next()
		if s.ch != '{' {
			s.errorf("expected { in escape sequence")
			return false
		}

		s.next()
		if !isHexDigit(s.ch) {
			s.errorf("illegal character %#U in escape sequence", s.ch)
			return false
		}

		s.next()
		for i := 0; i < 3; i++ {
			if !isHexDigit(s.ch) {
				return false
			}
			s.next()
		}

		if s.ch != '}' {
			s.errorf("expected } in escape sequence")
			return false
		}
	default:
		s.errorf("unknown escape sequence")
		return false
	}

	return true
}

func scanCharacter(s *Scanner) state {
	s.next()
	switch s.ch {
	case '\\':
		s.scanEscape('\'')
	case '\'':
		s.errorf("illegal character literal")
	}

	s.next()
	if s.ch == '\'' {
		s.next()
		s.emit(token.Char)
		return scanDefault
	}

	s.errorf("illegal character literal")
	for s.ch != '\'' && s.ch != eof {
		s.next()
	}

	s.emit(token.Char)
	return scanDefault
}

func scanString(s *Scanner) state {
	s.next()

	for {
		switch s.ch {
		case '\n', eof:
			s.errorf("string literal not terminated")
			return scanDefault
		case '\\':
			s.scanEscape('"')
		case '"':
			s.next()
			s.emit(token.String)
			return scanDefault
		default:
			s.next()
		}
	}
}

func scanIdent(s *Scanner) state {
	var typ token.TokenType

	for {
		s.next()
		ch := s.ch
		if !(isAlpha(ch) || isDigit(ch) || ch == '_' || ch == '\'') {
			break
		}
	}

	switch s.lexeme() {
	case "if":
		typ = token.If
	case "else":
		typ = token.Else
	case "while":
		typ = token.While
	case "return":
		typ = token.Return
	case "length":
		typ = token.Length
	case "use":
		typ = token.Use
	case "int":
		typ = token.Int
	case "bool":
		typ = token.Bool
	case "true":
		typ = token.True
	case "false":
		typ = token.False
	default:
		typ = token.Ident
	}
	s.emit(typ)

	return scanDefault
}

func scanInt(s *Scanner) state {
	for {
		s.next()

		if isAlpha(s.ch) {
			s.errorf("unexpected token: %#U", s.ch)
		}

		if !isDigit(s.ch) {
			break
		}
	}

	s.emit(token.Integer)
	return scanDefault
}

func (s *Scanner) errorf(format string, args ...interface{}) {
	s.tokens <- token.Token{Typ: token.Error, Lit: fmt.Sprintf(format, args...)}
}

func scanDefault(s *Scanner) state {
	var typ token.TokenType

	ch := s.ch

	switch {
	case isAlpha(ch):
		return scanIdent(s)
	case isDigit(ch):
		return scanInt(s)
	case ch == '"':
		return scanString(s)
	case ch == '\'':
		return scanCharacter(s)
	default:
		s.next()
		switch ch {
		case eof:
			s.tokens <- token.Token{Typ: token.Eof, Lit: ""}
			return nil
		case '+':
			typ = token.Add
		case '-':
			typ = token.Sub
		case '*':
			typ = token.Mul
		case '/':
			if s.ch == '/' {
				s.bump() // bump one to ignore 'ch'
				for s.ch != '\n' && s.ch != eof {
					s.bump() // bump
					s.next()
				}
				return scanDefault
			}
			typ = token.Div
		case '%':
			typ = token.Rem
		case '&':
			typ = token.And
		case '|':
			typ = token.Or
		case '(':
			typ = token.Lparen
		case '[':
			typ = token.Lbrack
		case '{':
			typ = token.Lbrace
		case ')':
			typ = token.Rparen
		case ']':
			typ = token.Rbrack
		case '}':
			typ = token.Rbrace
		case ',':
			typ = token.Comma
		case ':':
			typ = token.Colon
		case ';':
			typ = token.Semicolon
		case '=':
			typ = s.switch2(token.Assign, token.Eq)
		case '!':
			typ = s.switch2(token.Not, token.Neq)
		case '<':
			typ = s.switch2(token.Lt, token.Le)
		case '>':
			typ = s.switch2(token.Gt, token.Ge)
		case '\n':
			s.line += 1
			s.linepos = s.pos
			fallthrough
		case ' ', '\t':
			s.bump()
			return scanDefault
		default:
			s.errorf("unexpected token: %#U", ch)
		}
	}

	s.emit(typ)
	return scanDefault
}

func (s *Scanner) run() {
	for state := scanDefault; state != nil; {
		state = state(s)
	}

	close(s.tokens)
}

func NewScanner(src []byte, err ErrorHandler) *Scanner {
	s := &Scanner{
		src:    src,
		err:    err,
		tokens: make(chan token.Token),
		ch:     ' ',
		line:   1,
	}

	s.next()

	go s.run()
	return s
}

func (s *Scanner) Scan() token.Token {
	var tok token.Token

	for {
		tok = <-s.tokens

		if tok.Typ == token.Error {
			if s.err != nil {
				s.err(tok.Pos, tok.Lit)
			}
			s.ErrorCount += 1

			return tok
		} else {
			break
		}
	}

	return tok
}

func (s *Scanner) lexeme() string {
	return string(s.src[s.start:s.pos])
}

func (s *Scanner) emit(typ token.TokenType) {
	s.tokens <- token.Token{Typ: typ, Lit: s.lexeme(), Pos: token.Position{Filename: "", Line: s.line, Column: s.start - s.linepos + 1}}
	s.start = s.pos
}
