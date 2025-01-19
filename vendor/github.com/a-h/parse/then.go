package parse

// Then matches a sequence of two parsers. For multiples of the same type, use Times, Repeat, AtLeast, AtMost, ZeroOrMore, OneOrMore.
func Then[A any, B any](a Parser[A], b Parser[B]) Parser[Tuple2[A, B]] {
	return sequenceOf2Parser[A, B]{
		A: a,
		B: b,
	}
}
