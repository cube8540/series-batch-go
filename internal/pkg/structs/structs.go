package structs

type Pair[A any, B any] struct {
	First  A
	Second B
}

func NewPair[A any, B any](f A, s B) *Pair[A, B] {
	return &Pair[A, B]{First: f, Second: s}
}
