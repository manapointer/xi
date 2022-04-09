package types

import (
	"github.com/manapointer/xi/pkg/ast"
	"github.com/manapointer/xi/pkg/token"
)

type resultMode int

const (
	invalid resultMode = iota
	none
	unknown
	ok
)

type result struct {
	mode resultMode
	typ  Type
}

func isInt(typ Type) bool        { return isBasic(typ, Int) }
func isBool(typ Type) bool       { return isBasic(typ, Bool) }
func isIntOrArray(typ Type) bool { return isBasic(typ, Int) || isArray(typ) }
func any(typ Type) bool          { return true }

func isArray(typ Type) bool {
	switch typ.(type) {
	case *Array:
		return true
	default:
		return false
	}
}

func underlying(typ Type) Type {
	switch t := typ.(type) {
	case *Array:
		return t.elem
	default:
		return nil
	}
}

type OpPredicates map[token.TokenType]func(Type) bool

var unopPredicates = OpPredicates{
	token.Sub: isInt,
	token.Not: isBool,
}

var binopPredicates = OpPredicates{
	token.Add: isIntOrArray,
	token.Sub: isInt,
	token.Mul: isInt,
	token.Div: isInt,

	token.Le: isInt,
	token.Lt: isInt,
	token.Ge: isInt,
	token.Gt: isInt,

	token.Eq:  any,
	token.Neq: any,
	token.And: isBool,
	token.Or:  isBool,
}

func (c *Checker) predicate(r *result, predicates OpPredicates, op token.TokenType, typ Type) {
	if pred := predicates[op]; pred != nil {
		if !pred(typ) {
			panic("cannot apply operation")
		}
	} else {
		panic("unknown op")
	}
}

func isBasic(typ Type, kind BasicKind) bool {
	switch t := typ.(type) {
	case *Basic:
		return t.kind == kind
	default:
		return false
	}
}

type Checker struct {
	scope *Scope
}

func (check *Checker) openScope() {
	check.scope = NewScope(check.scope)
}

func (check *Checker) closeScope() {
	check.scope = check.scope.parent
}

func (c *Checker) decl() {

}

func (c *Checker) stmt() {

}

func (c *Checker) expr(r *result, expr ast.Expr) {
	switch t := expr.(type) {
	case *ast.Ident:
		c.ident(r, t)
	case *ast.BasicLit:
		c.basicLit(r, t.Kind)
	case *ast.LengthExpr:
		c.lengthExpr(r, t.Arg)
	case *ast.UnaryExpr:
		c.unaryExpr(r, t.Rhs, t.Op)
	case *ast.BinaryExpr:
		c.binaryExpr(r, t.Lhs, t.Rhs, t.Op)
	case *ast.SubscriptExpr:
		c.subscriptExpr(r, t.Lhs, t.Subscript)
	case *ast.ArrayLit:
		c.arrayLit(r, t.Elts)
	}
}

func (c *Checker) basicLit(r *result, kind token.TokenType) {
	switch kind {
	case token.String:
		r.typ = makeArrayType(PredeclaredTyp[Int], 0)
	case token.Int, token.Char:
		r.typ = PredeclaredTyp[Int]
	case token.True, token.False:
		r.typ = PredeclaredTyp[Bool]
	default:
		panic("invalid type for basic literal")
	}
}

func (c *Checker) lengthExpr(r *result, expr ast.Expr) {
	c.expr(r, expr)

	if !isArray(r.typ) {
		panic("can't take length of non-array value")
	}

	r.typ = PredeclaredTyp[Int]
	r.mode = ok
}

func (c *Checker) ident(r *result, ident *ast.Ident) {
	obj := c.scope.Lookup(ident.Name)
	if obj == nil {
		panic("not defined")
	}

	r.typ = obj.Type()
	r.mode = ok
}

func (c *Checker) unaryExpr(r *result, expr ast.Expr, op token.TokenType) {
	c.expr(r, expr)
	c.predicate(r, unopPredicates, op, r.typ)
	r.mode = ok
}

func (c *Checker) binaryExpr(r *result, lhs, rhs ast.Expr, op token.TokenType) {
	var r2 result

	c.expr(r, lhs)
	c.expr(&r2, rhs)

	if !TypeEqual(r.typ, r2.typ) {
		panic("types in binary expr not equal")
	}

	c.predicate(r, binopPredicates, op, r.typ)
	r.mode = ok
}

func (c *Checker) subscriptExpr(r *result, lhs, subscript ast.Expr) {
	c.expr(r, subscript)
	if !isInt(r.typ) {
		panic("cannot subscript an array with a non-integer value")
	}

	c.expr(r, lhs)
	if !isArray(r.typ) {
		panic("cannot subscript an non-array")
	}

	r.typ = underlying(r.typ)
	r.mode = ok
}

func (c *Checker) arrayLit(r *result, elts []ast.Expr) {
	if len(elts) == 0 {
		r.typ = nil
		r.mode = unknown
		return
	}

	c.expr(r, elts[0])

	var y result
	for _, elt := range elts[1:] {
		c.expr(&y, elt)
		if !TypeEqual(r.typ, y.typ) {
			panic("mismatched array elements")
		}
	}

	r.mode = ok
}
