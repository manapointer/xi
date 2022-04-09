package parser

import (
	"fmt"

	"github.com/manapointer/xi/pkg/ast"
	"github.com/manapointer/xi/pkg/scanner"
	"github.com/manapointer/xi/pkg/token"
)

type parser struct {
	scanner *scanner.Scanner
	indent  int
	trace   bool

	pos token.Position
	tok token.TokenType
	lit string
}

// Copied from go/parser. Credit to the Go team!
func (p *parser) printTrace(a ...interface{}) {
	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = len(dots)
	fmt.Printf("%5d:%3d: ", p.pos.Line, p.pos.Column)
	i := 2 * p.indent
	for i > n {
		fmt.Print(dots)
		i -= n
	}
	// i <= n
	fmt.Print(dots[0:i])
	fmt.Println(a...)
}

func trace(p *parser, msg string) *parser {
	p.printTrace(msg, "(")
	p.indent++
	return p
}

// Usage pattern: defer un(trace(p, "..."))
func un(p *parser) {
	p.indent--
	p.printTrace(")")
}

func (p *parser) init(filename string, src []byte, mode Mode) {
	p.scanner = scanner.NewScanner(src, nil)
	p.trace = mode&Trace != 0
	p.next()
}

func (p *parser) next() {
	tok := p.scanner.Scan()
	if tok.Typ == token.Error {
		panic(fmt.Errorf(tok.Lit))
	}

	p.pos, p.tok, p.lit = tok.Pos, tok.Typ, tok.Lit
}

func (p *parser) expect(tok token.TokenType) token.Position {
	pos := p.pos

	if p.tok != tok {
		panic(fmt.Errorf("unexpected token: %s, wanted: %s", p.lit, tok))
	}

	p.next()
	return pos
}

func (p *parser) parseIdent() *ast.Ident {
	if p.trace {
		defer un(trace(p, "Ident"))
	}

	name := "_"
	if p.tok == token.Ident {
		name = p.lit
		p.next()
	} else {
		p.expect(token.Ident)
	}
	return &ast.Ident{Name: name}
}

func (p *parser) parseArrayLit() *ast.ArrayLit {
	if p.trace {
		defer un(trace(p, "ArrayLit"))
	}

	p.expect(token.Lbrace)

	elts := []ast.Expr{}

	for p.tok != token.Rbrace && p.tok != token.Eof {
		elts = append(elts, p.parseExpr())
		if p.tok != token.Comma {
			break
		}
		p.next()
	}

	p.expect(token.Rbrace)
	return &ast.ArrayLit{Elts: elts}
}

func (p *parser) parseType() ast.Type {
	if p.trace {
		defer un(trace(p, "Type"))
	}

	if p.tok != token.Int && p.tok != token.Bool {
		panic(fmt.Errorf("unexpected token: %s", p.lit))
	}

	var typ ast.Type = &ast.PrimitiveType{Kind: p.tok}

	p.next()

	for p.tok == token.Lbrack {
		p.next()

		var size ast.Expr
		if p.tok != token.Rbrack {
			size = p.parseExpr()
		}

		p.expect(token.Rbrack)
		typ = &ast.ArrayType{
			Elt:  typ,
			Size: size,
		}
	}

	return typ
}

func (p *parser) parseCallExpr(ident0 *ast.Ident) *ast.CallExpr {
	if p.trace {
		defer un(trace(p, "CallExpr"))
	}

	p.expect(token.Lparen)

	var args []ast.Expr

	if p.tok != token.Rparen {
		for {
			args = append(args, p.parseExpr())
			if p.tok != token.Comma {
				break
			}
			p.next()
		}
	}

	p.expect(token.Rparen)

	return &ast.CallExpr{
		Func: ident0,
		Args: args,
	}
}

func (p *parser) parseSubscriptExpr(lhs ast.Expr) *ast.SubscriptExpr {
	if p.trace {
		defer un(trace(p, "SubscriptExpr"))
	}

	p.expect(token.Lbrack)
	subscript := p.parseExpr()
	p.expect(token.Rbrack)

	expr := &ast.SubscriptExpr{Lhs: lhs, Subscript: subscript}

	for p.tok == token.Lbrack {
		p.next()
		subscript := p.parseExpr()
		p.expect(token.Rbrack)
		expr = &ast.SubscriptExpr{Lhs: lhs, Subscript: subscript}
	}

	return expr
}

func (p *parser) parseLengthExpr(tok token.TokenType) ast.Expr {
	if p.trace {
		defer un(trace(p, "LengthExpr"))
	}

	p.expect(token.Lparen)

	arg := p.parseExpr()

	p.expect(token.Rparen)

	return &ast.LengthExpr{
		Tok: tok,
		Arg: arg,
	}
}

func (p *parser) parseExpr() ast.Expr {
	return p.parseExpr0(true)
}

func (p *parser) parseExpr0(strict bool) (expr ast.Expr) {
	if p.trace {
		defer un(trace(p, "Expr"))
	}

	if !strict {
		defer func() {
			if e := recover(); e != nil {
				switch e.(type) {
				case error:
					expr = nil
				default:
					panic(e)
				}
			}
		}()
	}

	return p.parseOrExpr()
}

func (p *parser) parseOrExpr() ast.Expr {
	var lhs ast.Expr = p.parseAndExpr()
	for p.tok == token.Or {
		tok := p.tok
		p.next()
		lhs = &ast.BinaryExpr{
			Lhs: lhs,
			Op:  tok,
			Rhs: p.parseAndExpr(),
		}
	}
	return lhs
}

func (p *parser) parseAndExpr() ast.Expr {
	var lhs ast.Expr = p.parseEqualityExpr()
	for p.tok == token.And {
		tok := p.tok
		p.next()
		lhs = &ast.BinaryExpr{
			Lhs: lhs,
			Op:  tok,
			Rhs: p.parseEqualityExpr(),
		}
	}
	return lhs
}

func (p *parser) parseEqualityExpr() ast.Expr {
	var lhs ast.Expr = p.parseComparisonExpr()
	for p.tok == token.Eq || p.tok == token.Neq {
		tok := p.tok
		p.next()
		lhs = &ast.BinaryExpr{
			Lhs: lhs,
			Op:  tok,
			Rhs: p.parseComparisonExpr(),
		}
	}
	return lhs
}

func (p *parser) parseComparisonExpr() ast.Expr {
	var lhs ast.Expr = p.parseTermExpr()
	for p.tok == token.Lt || p.tok == token.Le || p.tok == token.Gt || p.tok == token.Ge {
		tok := p.tok
		p.next()
		lhs = &ast.BinaryExpr{
			Lhs: lhs,
			Op:  tok,
			Rhs: p.parseTermExpr(),
		}
	}
	return lhs
}

func (p *parser) parseTermExpr() ast.Expr {
	var lhs ast.Expr = p.parseFactorExpr()
	for p.tok == token.Add || p.tok == token.Sub {
		tok := p.tok
		p.next()
		lhs = &ast.BinaryExpr{
			Lhs: lhs,
			Op:  tok,
			Rhs: p.parseFactorExpr(),
		}
	}
	return lhs
}

func (p *parser) parseFactorExpr() ast.Expr {
	var lhs ast.Expr = p.parseUnaryExpr()
	for p.tok == token.Mul || p.tok == token.Div || p.tok == token.Rem {
		tok := p.tok
		p.next()
		lhs = &ast.BinaryExpr{
			Lhs: lhs,
			Op:  tok,
			Rhs: p.parseUnaryExpr(),
		}
	}
	return lhs
}

func (p *parser) parseUnaryExpr() ast.Expr {
	if p.tok == token.Sub || p.tok == token.Not {
		tok := p.tok
		p.next()
		return &ast.UnaryExpr{
			Op:  tok,
			Rhs: p.parseCallOrSubscriptExpr(),
		}
	}
	return p.parseCallOrSubscriptExpr()
}

func (p *parser) parseCallOrSubscriptExpr() ast.Expr {
	if p.tok == token.Length {
		tok := p.tok
		p.next()
		return p.parseLengthExpr(tok)
	}

	lhs := p.parseBaseExpr()

loop:
	for {
		switch p.tok {
		case token.Lbrack:
			p.next()
			expr := p.parseOrExpr()
			p.expect(token.Rbrack)
			lhs = &ast.SubscriptExpr{
				Lhs:       lhs,
				Subscript: expr,
			}
		case token.Lparen:
			if _, ok := lhs.(*ast.Ident); !ok {
				panic(fmt.Errorf("can't call a non-identifier expression"))
			}
			lhs = p.parseCallExpr(lhs.(*ast.Ident))
		default:
			break loop
		}
	}

	return lhs
}

func (p *parser) parseBaseExpr() ast.Expr {
	if p.trace {
		defer un(trace(p, "BaseExpr"))
	}

	switch p.tok {
	case token.Ident:
		return p.parseIdent()
	case token.Integer, token.String, token.True, token.False, token.Char:
		lit := ast.BasicLit{Kind: p.tok, Value: p.lit}
		p.next()
		return &lit
	case token.Lbrace:
		return p.parseArrayLit()
	case token.Lparen:
		p.next()
		expr := p.parseOrExpr()
		p.expect(token.Rparen)
		return expr
	default:
		panic(fmt.Errorf("unexpected token: %s", p.tok))
	}
}

func (p *parser) parseLvalue(ident0 *ast.Ident) ast.Lvalue {
	if p.trace {
		defer un(trace(p, "Lvalue"))
	}

	if p.tok == token.Lbrack {
		return p.parseSubscriptExpr(ident0)
	}

	return ident0
}

func (p *parser) parseAssignable() ast.Assignable {
	if p.trace {
		defer un(trace(p, "Assignable"))
	}

	switch p.tok {
	case token.Underscore:
		return p.parseDiscard()
	case token.Ident:
		ident := p.parseIdent()
		p.expect(token.Colon)
		type_ := p.parseType()
		return &ast.Spec{Name: ident, Type: type_}
	default:
		panic("unreachable")
	}
}

func (p *parser) parseDeclStmtSpec(spec *ast.Spec) ast.Stmt {
	if p.trace {
		defer un(trace(p, "DeclStmtSpec"))
	}

	var assigns []ast.Assignable

	if p.tok == token.Comma {
		assigns = append(assigns, spec)
		for p.tok == token.Comma {
			p.next()
			assigns = append(assigns, p.parseAssignable())
		}
	} else {
		var init ast.Expr
		if p.tok == token.Assign {
			p.next()
			init = p.parseExpr()
		}
		return &ast.SingleDeclStmt{Spec: spec, Init: init}
	}

	return p.parseMultiDeclStmt(assigns)
}

func (p *parser) parseDeclStmtDiscard(discard *ast.Discard) ast.Stmt {
	if p.trace {
		defer un(trace(p, "DeclStmtDiscard"))
	}

	assigns := []ast.Assignable{discard}

	if p.tok == token.Comma {
		for p.tok == token.Comma {
			p.next()
			assigns = append(assigns, p.parseAssignable())
		}
	}

	return p.parseMultiDeclStmt(assigns)
}

func (p *parser) parseMultiDeclStmt(assigns []ast.Assignable) *ast.MultiDeclStmt {
	if p.trace {
		defer un(trace(p, "MultiDeclStmt"))
	}

	p.expect(token.Assign)
	ident := p.parseIdent()
	expr := p.parseCallExpr(ident)
	return &ast.MultiDeclStmt{Assignables: assigns, Init: expr}
}

// either single or multi
func (p *parser) parseDeclStmt(lvalue0 ast.Node) ast.Stmt {
	if p.trace {
		defer un(trace(p, "DeclStmt"))
	}

	switch t := lvalue0.(type) {
	case *ast.Discard:
		return p.parseDeclStmtDiscard(t)
	case *ast.Ident:
		p.expect(token.Colon)
		type_ := p.parseType()
		spec := &ast.Spec{Name: t, Type: type_}
		return p.parseDeclStmtSpec(spec)
	default:
		panic("unreachable")
	}
}

func (p *parser) parseAssignStmt(ident0 *ast.Ident) *ast.AssignStmt {
	if p.trace {
		defer un(trace(p, "AssignStmt"))
	}

	lvalue := p.parseLvalue(ident0)
	p.expect(token.Assign)
	expr := p.parseExpr()
	return &ast.AssignStmt{Lhs: lvalue, Rhs: expr}
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	if p.trace {
		defer un(trace(p, "If"))
	}

	p.expect(token.If)

	cond := p.parseExpr()
	then := p.parseStmt()

	var else_ ast.Stmt
	if p.tok == token.Else {
		p.next()
		else_ = p.parseStmt()
	}

	return &ast.IfStmt{Cond: cond, Then: then, Else: else_}
}

func (p *parser) parseWhileStmt() *ast.WhileStmt {
	if p.trace {
		defer un(trace(p, "While"))
	}

	p.expect(token.While)

	cond := p.parseExpr()
	body := p.parseStmt()

	return &ast.WhileStmt{Cond: cond, Body: body}
}

func (p *parser) parseReturn() *ast.ReturnStmt {
	if p.trace {
		defer un(trace(p, "Return"))
	}

	p.expect(token.Return)

	vals := make([]ast.Expr, 0)
	val := p.parseExpr0(false)

	if val != nil {
		vals = append(vals, val)
		for {
			if p.tok != token.Comma {
				break
			}

			p.next()

			vals = append(vals, p.parseExpr())
		}
	}

	return &ast.ReturnStmt{Values: vals}
}

func (p *parser) parseBlock() *ast.BlockStmt {
	if p.trace {
		defer un(trace(p, "Block"))
	}

	p.expect(token.Lbrace)

	list := make([]ast.Stmt, 0)

	for p.tok != token.Rbrace && p.tok != token.Eof {
		list = append(list, p.parseStmt())

		if p.tok == token.Semicolon {
			p.next()
		}
	}

	p.expect(token.Rbrace)
	return &ast.BlockStmt{List: list}
}

func (p *parser) parseDiscard() *ast.Discard {
	if p.trace {
		defer un(trace(p, "Discard"))
	}

	p.expect(token.Underscore)
	return &ast.Discard{}
}

func (p *parser) parseStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "Stmt"))
	}

	switch p.tok {
	case token.Ident:
		ident0 := p.parseIdent()
		switch p.tok {
		case token.Colon:
			return p.parseDeclStmt(ident0)
		case token.Assign, token.Lbrack:
			return p.parseAssignStmt(ident0)
		case token.Lparen:
			return p.parseCallExpr(ident0)
		default:
			panic(fmt.Errorf("unexpected token in stmt: %s", p.tok))
		}
	case token.Underscore:
		return p.parseDeclStmt(p.parseDiscard())
	case token.If:
		return p.parseIfStmt()
	case token.While:
		return p.parseWhileStmt()
	case token.Return:
		return p.parseReturn()
	case token.Lbrace:
		return p.parseBlock()
	default:
		panic(fmt.Errorf("unknown token: %+v", p.tok))
	}
}

func (p *parser) parseParameters() []*ast.Spec {
	if p.trace {
		defer un(trace(p, "Parameters"))
	}

	l := make([]*ast.Spec, 0)

	for p.tok != token.Eof && p.tok != token.Rparen {
		ident := p.parseIdent()
		p.expect(token.Colon)
		typ := p.parseType()

		l = append(l, &ast.Spec{Name: ident, Type: typ})
		if p.tok != token.Comma {
			break
		}

		p.next()
	}

	p.expect(token.Rparen)
	return l
}

func (p *parser) parseResults() []ast.Type {
	if p.trace {
		un(trace(p, "Results"))
	}

	p.expect(token.Colon)

	results := []ast.Type{p.parseType()}

	for p.tok == token.Comma {
		p.next()
		results = append(results, p.parseType())
	}

	return results
}

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	if p.trace {
		defer un(trace(p, "FuncDecl"))
	}

	ident := p.parseIdent()

	p.expect(token.Lparen)

	args := p.parseParameters()

	var results []ast.Type
	if p.tok == token.Colon {
		results = p.parseResults()
	}

	body := p.parseBlock()
	decl := &ast.FuncDecl{
		Name:    ident,
		Body:    body,
		Args:    args,
		Results: results,
	}
	return decl
}

func (p *parser) parseUseDecl() *ast.UseDecl {
	if p.trace {
		defer un(trace(p, "UseDecl"))
	}

	p.expect(token.Use)
	return &ast.UseDecl{
		Lib: p.parseIdent(),
	}
}

func (p *parser) parseFile() *ast.File {
	if p.trace {
		defer un(trace(p, "File"))
	}

	var (
		useDecls  []*ast.UseDecl
		funcDecls []*ast.FuncDecl
	)

	for p.tok == token.Use {
		useDecls = append(useDecls, p.parseUseDecl())
	}

	for p.tok != token.Eof {
		funcDecls = append(funcDecls, p.parseFuncDecl())
	}

	return &ast.File{
		FuncDecls: funcDecls,
		UseDecls:  useDecls,
	}
}
