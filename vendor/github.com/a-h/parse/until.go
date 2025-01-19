package parse

type untilParser[T, D any] struct {
	Parser    Parser[T]
	Delimiter Parser[D]
	AllowEOF  bool
}

func (p untilParser[T, D]) Parse(in *Input) (match []T, ok bool, err error) {
	start := in.Index()
	if _, ok = in.Peek(1); !ok && p.AllowEOF {
		ok = true
		return
	}
	var m T
	m, ok, err = p.Parser.Parse(in)
	if err != nil {
		return
	}
	if !ok {
		return
	}
	match = append(match, m)
	for {
		beforeDelimiter := in.Index()
		_, ok, err = p.Delimiter.Parse(in)
		if err != nil {
			in.Seek(start)
			return
		}
		if ok {
			in.Seek(beforeDelimiter)
			return
		}
		if _, ok = in.Peek(1); !ok && p.AllowEOF {
			ok = true
			return
		}
		var m T
		m, ok, err = p.Parser.Parse(in)
		if err != nil {
			in.Seek(start)
			return
		}
		if !ok {
			in.Seek(start)
			return
		}
		match = append(match, m)
	}
}

// Until matches until the delimiter is reached.
func Until[T, D any](parser Parser[T], delimiter Parser[D]) Parser[[]T] {
	return untilParser[T, D]{
		Parser:    parser,
		Delimiter: delimiter,
	}
}

// UntilEOF matches until the delimiter or the end of the file is reached.
func UntilEOF[T, D any](parser Parser[T], delimiter Parser[D]) Parser[[]T] {
	return untilParser[T, D]{
		Parser:    parser,
		Delimiter: delimiter,
		AllowEOF:  true,
	}
}
