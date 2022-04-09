package types

import "github.com/manapointer/xi/pkg/token"

var Universe *Scope

var PredeclaredTyp = []*Basic{
	Bool: {kind: Bool, name: "bool"},
	Int:  {kind: Int, name: "int"},
}

var PredeclaredFunc = [...]struct {
	name string
	sig  *Signature
}{}

func init() {
	Universe = NewScope(nil)
}

func defPredeclaredTypes() {
	for _, typ := range PredeclaredTyp {
		Universe.Insert(NewTypeName(token.Position{}, typ.name, typ))
	}
}

// func defPredeclaredFuncs() {
// 	for _, f := range PredeclaredFunc {
// 		Universe.Insert()
// 	}
// }

// func newFunc(name string, sig *Signature) *Func {
// 	return &Func{object{name, nil, token.Position{}, PredeclaredTyp[Bool]}}
// }
