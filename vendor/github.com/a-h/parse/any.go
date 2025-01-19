package parse

type anyParser[T any] struct {
	Parsers []Parser[T]
}

func (p anyParser[T]) Parse(in *Input) (match T, ok bool, err error) {
	for _, parser := range p.Parsers {
		match, ok, err = parser.Parse(in)
		if err != nil || ok {
			return
		}
	}
	return
}

// Any parses any one of the parsers in the list.
func Any[T any](parsers ...Parser[T]) Parser[T] {
	return anyParser[T]{
		Parsers: parsers,
	}
}
