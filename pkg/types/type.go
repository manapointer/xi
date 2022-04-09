package types

import "bytes"

type Type interface {
	String() string
	Hash() int
}

type BasicKind int

const (
	Invalid BasicKind = iota
	Bool
	Int
)

type Basic struct {
	kind BasicKind
	name string
}

func (t *Basic) Kind() BasicKind { return t.kind }
func (t *Basic) Name() string    { return t.name }
func (t *Basic) String() string  { return TypeString(t) }

func (t *Basic) Hash() int {
	switch t.kind {
	case Bool:
		return 0
	case Int:
		return 1
	default:
		return -1
	}
}

type Array struct {
	elem Type
}

func (t *Array) Elem() Type     { return t.elem }
func (t *Array) String() string { return TypeString(t) }

func (t *Array) Hash() int {
	return 2 + t.elem.Hash()
}

type Tuple struct {
	types []Type
}

func (t *Tuple) Types() []Type  { return t.types }
func (t *Tuple) String() string { return TypeString(t) }

func (t *Tuple) Hash() int { return -1 }

type Signature struct {
	parameters *Tuple
	returns    *Tuple
}

func (s *Signature) Parameters() *Tuple {
	return s.parameters
}

func (s *Signature) Returns() *Tuple {
	return s.returns
}

func (t *Signature) String() string { return TypeString(t) }

func (t *Signature) Hash() int { return -1 }

func TypeString(typ Type) string {
	var buf bytes.Buffer
	w := &typeWriter{&buf}
	w.typ(typ)
	return buf.String()
}

type typeWriter struct {
	buf *bytes.Buffer
}

func (w *typeWriter) str(s string) {
	w.buf.WriteString(s)
}

func (w *typeWriter) byte(b byte) {
	w.buf.WriteByte(b)
}

func (w *typeWriter) typ(typ Type) {
	switch t := typ.(type) {
	case *Basic:
		w.str(t.name)
	case *Array:
		w.typ(t.elem)
		w.byte('[')
		w.byte(']')
	case *Tuple:
		w.byte('(')
		if len(t.types) > 0 {
			w.typ(t.types[0])
			for _, typ := range t.types[1:] {
				w.byte(',')
				w.typ(typ)
			}
		}
		w.byte(')')
	case *Signature:
		w.str("function ")
		w.typ(t.parameters)
		w.byte(' ')
		w.typ(t.returns)
	}
}

func TypeEqual(t1, t2 Type) bool {
	return t1.Hash() == t2.Hash()
}

var arrayTypeCache = make(map[int]*Array)

func makeArrayType(basic *Basic, extraDimensions int) *Array {
	var hash int

	switch basic.kind {
	case Bool:
		hash = 2
	case Int:
		hash = 3
	default:
		panic("invalid type for array")
	}

	hash += extraDimensions * 2

	if typ, ok := arrayTypeCache[hash]; ok {
		return typ
	}

	typ := &Array{elem: basic}

	for i := 0; i < extraDimensions; i++ {
		typ = &Array{elem: typ}
	}

	arrayTypeCache[hash] = typ
	return typ
}
