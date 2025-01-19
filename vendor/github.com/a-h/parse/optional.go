package parse

type optionalParser[T any] struct {
	Parser      Parser[T]
	Insensitive bool
}

func (p optionalParser[T]) Parse(in *Input) (match Match[T], ok bool, err error) {
	match.Value, match.OK, err = p.Parser.Parse(in)
	if err != nil {
		return
	}
	return match, true, nil
}

type Match[T any] struct {
	Value T
	OK    bool
}

// Optional converts the given parser into an optional parser.
func Optional[T any](parser Parser[T]) Parser[Match[T]] {
	return optionalParser[T]{
		Parser: parser,
	}
}
