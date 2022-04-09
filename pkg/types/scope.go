package types

import "github.com/manapointer/xi/pkg/token"

type Scope struct {
	parent *Scope
	elems  map[string]Object
}

func NewScope(parent *Scope) *Scope {
	return &Scope{parent: parent, elems: make(map[string]Object)}
}

func (s *Scope) Lookup(name string) Object {
	return s.elems[name]
}

func (s *Scope) LookupParent(name string, pos token.Position) (*Scope, Object) {
	for ; s != nil; s = s.parent {
		if obj, ok := s.elems[name]; ok && obj.Position().Compare(pos) != 1 {
			return s, obj
		}
	}

	return nil, nil
}

func (s *Scope) Insert(obj Object) bool {
	name := obj.Name()
	if alt := s.Lookup(name); alt != nil {
		return false
	}

	s.elems[name] = obj
	if obj.Parent() == nil {
		obj.setParent(s)
	}
	return true
}
