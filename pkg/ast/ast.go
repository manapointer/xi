package ast

import "github.com/manapointer/xi/pkg/token"

type Node interface {
}

type Decl interface {
	Node
	declNode()
}

type Expr interface {
	Node
	exprNode()
}

type Stmt interface {
	Node
	stmtNode()
}

type Type interface {
	typeNode()
}

type (
	PrimitiveType struct {
		Kind token.TokenType
	}

	ArrayType struct {
		Elt Type
	}
)

func (*PrimitiveType) typeNode() {}
func (*ArrayType) typeNode()     {}

type (
	Ident struct {
		Name string
	}

	BasicLit struct {
		Kind  token.TokenType
		Value string
	}

	ArrayLit struct {
		Elts []Expr
	}

	CallExpr struct {
		Func *Ident
		Args []Expr
	}

	SubscriptExpr struct {
		Lhs       Expr
		Subscript Expr
	}

	UnaryExpr struct {
		Op  token.TokenType
		Rhs Expr
	}

	BinaryExpr struct {
		Op  token.TokenType
		Lhs Expr
		Rhs Expr
	}
)

func (*Ident) exprNode()         {}
func (*BasicLit) exprNode()      {}
func (*ArrayLit) exprNode()      {}
func (*CallExpr) exprNode()      {}
func (*CallExpr) stmtNode()      {}
func (*SubscriptExpr) exprNode() {}

type Spec struct {
	Name *Ident
	Type Type
}

type Lvalue interface {
	Node
	lvalueNode()
}

type Discard struct {
	Tok token.Token
}

func (*Ident) lvalueNode()         {}
func (*SubscriptExpr) lvalueNode() {}
func (*Discard) lvalueNode()       {}

type (
	AssignStmt struct {
		Lhs Lvalue
		Rhs Expr
	}

	IfStmt struct {
		Cond Expr
		Then Stmt
		Else Stmt
	}

	WhileStmt struct {
		Cond Expr
		Body Stmt
	}

	ReturnStmt struct {
		Values []Expr
	}

	BlockStmt struct {
		List []Stmt
	}

	SingleDeclStmt struct {
		Spec *Spec
		Init Expr
	}

	MultiDeclStmt struct {
		Specs []*Spec
		Init  *CallExpr
	}
)

func (*AssignStmt) stmtNode()     {}
func (*IfStmt) stmtNode()         {}
func (*WhileStmt) stmtNode()      {}
func (*ReturnStmt) stmtNode()     {}
func (*BlockStmt) stmtNode()      {}
func (*SingleDeclStmt) stmtNode() {}
func (*MultiDeclStmt) stmtNode()  {}

type (
	FuncDecl struct {
		Name    *Ident
		Args    []*Spec
		Body    *BlockStmt
		Results []Type
	}

	UseDecl struct {
		Lib *Ident
	}
)

func (*FuncDecl) declNode() {}
func (*UseDecl) declNode()  {}

type File struct {
	FuncDecls []*FuncDecl
	UseDecls  []*UseDecl
}
