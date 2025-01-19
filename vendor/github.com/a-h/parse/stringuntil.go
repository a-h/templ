package parse

type stringUntilParser[T any] struct {
	Delimiter Parser[T]
	AllowEOF  bool
}

func (p stringUntilParser[T]) Parse(in *Input) (match string, ok bool, err error) {
	start := in.Index()
	for {
		beforeDelimiter := in.Index()
		_, ok, err = p.Delimiter.Parse(in)
		if err != nil {
			in.Seek(start)
			return
		}
		if ok {
			in.Seek(beforeDelimiter)
			break
		}
		_, chompOK := in.Take(1)
		if !chompOK {
			if p.AllowEOF {
				break
			}
			in.Seek(start)
			return "", false, nil
		}
	}
	end := in.Index()
	in.Seek(start)
	match, ok = in.Take(end - start)
	return
}

// StringUntil matches until the delimiter is reached.
func StringUntil[T any](delimiter Parser[T]) Parser[string] {
	return stringUntilParser[T]{
		Delimiter: delimiter,
	}
}

// StringUntilEOF matches until the delimiter or the end of the file is reached.
func StringUntilEOF[T any](delimiter Parser[T]) Parser[string] {
	return stringUntilParser[T]{
		Delimiter: delimiter,
		AllowEOF:  true,
	}
}
