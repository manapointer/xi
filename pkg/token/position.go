package token

type Position struct {
	Filename string
	Line     int
	Column   int
}

func compareInt(lhs, rhs int) int {
	switch {
	case lhs < rhs:
		return -1
	case rhs < lhs:
		return 1
	default:
		return 0
	}
}

func (pos Position) Compare(other Position) int {
	cmp := compareInt(pos.Line, other.Line)
	if cmp == 0 {
		return compareInt(pos.Column, other.Column)
	}
	return cmp
}
