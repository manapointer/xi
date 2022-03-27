package parser

import (
	"fmt"
	"os"

	"github.com/manapointer/xi/pkg/ast"
	"github.com/manapointer/xi/pkg/scanner"
	"github.com/manapointer/xi/pkg/token"
)

type parser struct {
	scanner *scanner.Scanner
	tok     token.Token
}

func (p *parser) init(filename string, src []byte) {
	p.scanner = scanner.NewScanner(src, nil)
	p.next()
}

func (p *parser) next() {
	p.tok = p.scanner.Scan()
}

func (p *parser) expect(typ token.TokenType) token.Position {
	pos := p.tok.Pos
	p.next()
	return pos
}

func (p *parser) parseIdent() *ast.Ident {
	name := "_"
	if p.tok.Typ == token.Ident {
		name = p.tok.Lit
		p.next()
	} else {
		p.expect(token.Ident)
	}
	return &ast.Ident{Name: name}
}

func (p *parser) parseArrayLit() *ast.ArrayLit {
	p.expect(token.Lbrace)

	elts := []ast.Expr{}

	for p.tok.Typ != token.Rbrace {
		elts = append(elts, p.parseExpr())
	}

	p.next()
	return &ast.ArrayLit{Elts: elts}
}

func (p *parser) parseType() ast.Type {
	fmt.Fprintf(os.Stderr, "type\n")
	if p.tok.Typ != token.Int && p.tok.Typ != token.Char {
		panic("invalid type")
	}
	var typ ast.Type = &ast.PrimitiveType{Kind: p.tok.Typ}
	p.next()
	for p.tok.Typ == token.Lbrack {
		fmt.Fprintf(os.Stderr, "lbrack")
		p.next()
		p.expect(token.Rbrack)
		typ = &ast.ArrayType{
			Elt: typ,
		}
	}

	fmt.Fprintln(os.Stderr, "done type", typ)
	return typ
}

func (p *parser) parseCallExpr(ident0 *ast.Ident) *ast.CallExpr {
	p.expect(token.Rparen)

	var args = make([]ast.Expr, 0)

	for {
		args = append(args, p.parseExpr())
		if p.tok.Typ != token.Comma {
			break
		}
		p.next()
	}

	p.expect(token.Lparen)

	fmt.Fprintf(os.Stderr, "return call expr\n")

	return &ast.CallExpr{
		Func: ident0,
		Args: args,
	}
}

func (p *parser) parseSubscriptExpr(lhs ast.Expr) *ast.SubscriptExpr {
	p.expect(token.Lbrack)
	subscript := p.parseExpr()
	p.expect(token.Rbrack)

	expr := &ast.SubscriptExpr{Lhs: lhs, Subscript: subscript}

	for p.tok.Typ == token.Lbrack {
		p.next()
		subscript := p.parseExpr()
		p.expect(token.Rbrack)
		expr = &ast.SubscriptExpr{Lhs: lhs, Subscript: subscript}
	}

	return expr
}

func (p *parser) parseExpr() ast.Expr {
	fmt.Fprintf(os.Stderr, "expr\n")
	switch p.tok.Typ {
	case token.Ident:
		ident0 := p.parseIdent()
		switch p.tok.Typ {
		case token.Lparen:
			return p.parseCallExpr(ident0)
		case token.Lbrack:
			return p.parseSubscriptExpr(ident0)
		default:
			return ident0
		}
	case token.Int, token.String:
		lit := ast.BasicLit{Kind: p.tok.Typ, Value: p.tok.Lit}
		p.next()
		return &lit
	case token.Lbrace:
		return p.parseArrayLit()
	}
	return nil
}

func (p *parser) parseLvalue(ident0 *ast.Ident) ast.Lvalue {
	if p.tok.Typ == token.Lbrack {
		return p.parseSubscriptExpr(ident0)
	}

	return ident0
}

// either single or multi
func (p *parser) parseDeclStmt(lvalue0 ast.Lvalue) ast.Stmt {

	// switch t := lvalue0.(type) {
	// case *ast.Discard:

	// default:
	// 	panic("unreachable")
	// }
	return nil
}

func (p *parser) parseAssignStmt(ident0 *ast.Ident) *ast.AssignStmt {
	lvalue := p.parseLvalue(ident0)
	p.expect(token.Assign)
	expr := p.parseExpr()
	return &ast.AssignStmt{Lhs: lvalue, Rhs: expr}
}

func (p *parser) parseIf() *ast.IfStmt {
	p.expect(token.If)
	cond := p.parseExpr()
	then := p.parseStmt()
	p.expect(token.Else)
	els := p.parseStmt()
	return &ast.IfStmt{Cond: cond, Then: then, Else: els}
}

func (p *parser) parseWhile() *ast.WhileStmt {
	p.expect(token.While)
	cond := p.parseExpr()
	body := p.parseStmt()
	return &ast.WhileStmt{Cond: cond, Body: body}
}

func (p *parser) parseReturn() *ast.ReturnStmt {
	p.expect(token.Return)

	vals := make([]ast.Expr, 0)
	val := p.parseExpr()

	if val != nil {
		vals = append(vals, val)
		for {
			if p.tok.Typ != token.Comma {
				break
			}

			vals = append(vals, p.parseExpr())
		}
	}

	return &ast.ReturnStmt{Values: vals}
}

func (p *parser) parseBlock() *ast.BlockStmt {
	p.expect(token.Lbrace)

	list := make([]ast.Stmt, 0)

	for p.tok.Typ != token.Rbrace && p.tok.Typ != token.Eof {
		list = append(list, p.parseStmt())

		if p.tok.Typ == token.Semicolon {
			p.next()
		}
	}

	p.expect(token.Rbrace)
	return &ast.BlockStmt{List: list}
}

func (p *parser) parseDiscard() *ast.Discard {
	tok := p.tok
	p.expect(token.Underscore)
	return &ast.Discard{Tok: tok}
}

func (p *parser) parseStmt() ast.Stmt {
	fmt.Fprintf(os.Stderr, "stmt\n")
	switch p.tok.Typ {
	case token.Ident:
		ident0 := p.parseIdent()
		switch p.tok.Typ {
		case token.Colon:
			return p.parseDeclStmt(ident0)
		case token.Assign:
			return p.parseAssignStmt(ident0)
		case token.Lparen:
			return p.parseCallExpr(ident0)
		default:
			panic(fmt.Sprintf("unexpected token in stmt: %s", p.tok.Typ))
		}
	case token.Underscore:
		return p.parseDeclStmt(p.parseDiscard())
	case token.If:
		return p.parseIf()
	case token.While:
		return p.parseWhile()
	case token.Return:
		return p.parseReturn()
	case token.Lbrace:
		return p.parseBlock()
	default:
		panic(fmt.Sprintf("unknown token: %+v", p.tok))
	}
}

func (p *parser) parseParameters() []*ast.Spec {
	l := make([]*ast.Spec, 0)

	for p.tok.Typ != token.Eof && p.tok.Typ != token.Rparen {
		ident := p.parseIdent()
		p.expect(token.Colon)
		typ := p.parseType()

		l = append(l, &ast.Spec{Name: ident, Type: typ})
		if p.tok.Typ != token.Comma {
			break
		}

		p.next()

	}

	fmt.Fprintf(os.Stderr, "params\n")

	p.expect(token.Rparen)
	return l
}

func (p *parser) parseFuncBody() *ast.BlockStmt {
	fmt.Fprintf(os.Stderr, "parse func body\n")
	fmt.Fprintf(os.Stderr, "token %s\n", p.tok.Typ)
	p.expect(token.Lbrace)

	var l []ast.Stmt
	for p.tok.Typ != token.Rbrace {
		l = append(l, p.parseStmt())
	}

	p.next()

	fmt.Fprintf(os.Stderr, "done func body\n")

	return &ast.BlockStmt{
		List: l,
	}
}

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	ident := p.parseIdent()
	p.expect(token.Lparen)
	args := p.parseParameters()

	var results []ast.Type

	fmt.Fprintf(os.Stderr, "parse results")

	for p.tok.Typ == token.Ident {
		results = append(results, p.parseType())
	}

	body := p.parseFuncBody()

	return &ast.FuncDecl{
		Name:    ident,
		Body:    body,
		Args:    args,
		Results: results,
	}
}

func (p *parser) parseUseDecl() *ast.UseDecl {
	p.expect(token.Use)
	return &ast.UseDecl{
		Lib: p.parseIdent(),
	}
}

func (p *parser) parseFile() *ast.File {
	var (
		useDecls  []*ast.UseDecl
		funcDecls []*ast.FuncDecl
	)

	for p.tok.Typ == token.Use {
		useDecls = append(useDecls, p.parseUseDecl())
	}

	for p.tok.Typ != token.Eof {
		funcDecls = append(funcDecls, p.parseFuncDecl())
	}

	return &ast.File{
		FuncDecls: funcDecls,
		UseDecls:  useDecls,
	}
}
