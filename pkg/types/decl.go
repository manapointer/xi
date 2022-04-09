package types

import (
	"fmt"
	"go/ast"

	"github.com/manapointer/xi/pkg/token"
)

func (c *Checker) declare(s *Scope, ident *ast.Ident, obj Object, pos token.Position) {
	if !s.Insert(obj) {
		fmt.Println("duplicate declaration")
	}
}
