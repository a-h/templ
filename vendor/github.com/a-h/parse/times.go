package parse

type timesParser[T any] struct {
	P   Parser[T]
	Min int
	Max func(i int) bool
}

func (p timesParser[T]) Parse(in *Input) (match []T, ok bool, err error) {
	start := in.Index()
	for i := 0; p.Max(i); i++ {
		var m T
		m, ok, err = p.P.Parse(in)
		if err != nil {
			return match, false, err
		}
		if !ok {
			break
		}
		match = append(match, m)
	}
	ok = len(match) >= p.Min && p.Max(len(match)-1)
	if !ok {
		in.Seek(start)
		return match, false, nil
	}
	return match, true, nil
}

// Times matches the given parser n times.
func Times[T any](n int, p Parser[T]) Parser[[]T] {
	return timesParser[T]{
		P:   p,
		Min: n,
		Max: func(i int) bool { return i < n },
	}
}

// Repeat matches the given parser between min and max times.
func Repeat[T any](min, max int, p Parser[T]) Parser[[]T] {
	return timesParser[T]{
		P:   p,
		Min: min,
		Max: func(i int) bool { return i < max },
	}
}

// AtLeast matches the given parser at least min times.
func AtLeast[T any](min int, p Parser[T]) Parser[[]T] {
	return timesParser[T]{
		P:   p,
		Min: min,
		Max: func(i int) bool { return true },
	}
}

// AtMost matches the given parser at most max times.
// It is equivalent to ZeroOrMore.
func AtMost[T any](max int, p Parser[T]) Parser[[]T] {
	return timesParser[T]{
		P:   p,
		Min: 0,
		Max: func(i int) bool { return i < max },
	}
}

// ZeroOrMore matches the given parser zero or more times.
func ZeroOrMore[T any](p Parser[T]) Parser[[]T] {
	return timesParser[T]{
		P:   p,
		Min: 0,
		Max: func(i int) bool { return true },
	}
}

// OneOrMore matches the given parser at least once.
func OneOrMore[T any](p Parser[T]) Parser[[]T] {
	return timesParser[T]{
		P:   p,
		Min: 1,
		Max: func(i int) bool { return true },
	}
}
