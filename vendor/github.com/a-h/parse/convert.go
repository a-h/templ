package parse

// Convert a parser's output type using the given conversion function.
func Convert[A, B any](parser Parser[A], converter func(a A) (B, error)) Parser[B] {
	return Func(func(in *Input) (match B, ok bool, err error) {
		var m A
		m, ok, err = parser.Parse(in)
		if err != nil || !ok {
			return
		}
		match, err = converter(m)
		return
	})
}
