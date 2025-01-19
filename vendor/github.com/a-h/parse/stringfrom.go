package parse

type stringFromParser[T any] struct {
	Parsers []Parser[T]
}

func (p stringFromParser[T]) Parse(in *Input) (match string, ok bool, err error) {
	start := in.Index()
	for _, parser := range p.Parsers {
		_, ok, err = parser.Parse(in)
		if err != nil {
			return
		}
		if !ok {
			in.Seek(start)
			return
		}
	}
	end := in.Index()
	in.Seek(start)
	match, ok = in.Take(end - start)
	return
}

// StringFrom returns the string range captured by the given parsers.
func StringFrom[T any](parsers ...Parser[T]) Parser[string] {
	return stringFromParser[T]{
		Parsers: parsers,
	}
}
