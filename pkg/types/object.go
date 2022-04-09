package types

import "github.com/manapointer/xi/pkg/token"

type Object interface {
	Parent() *Scope
	Name() string
	Position() token.Position
	Type() Type

	setParent(*Scope)
}

type object struct {
	name   string
	parent *Scope
	pos    token.Position
	typ    Type
}

func (obj *object) Parent() *Scope           { return obj.parent }
func (obj *object) Name() string             { return obj.name }
func (obj *object) Position() token.Position { return obj.pos }
func (obj *object) Type() Type               { return obj.typ }

func (obj *object) setParent(parent *Scope) {
	obj.parent = parent
}

func (obj *object) setPosition(pos token.Position) {
	obj.pos = pos
}

type Func struct {
	object
}

type Builtin struct {
	object
}

type Var struct {
	object
}

type TypeName struct {
	object
}

func NewTypeName(pos token.Position, name string, typ Type) *TypeName {
	return &TypeName{object{name, nil, pos, typ}}
}
