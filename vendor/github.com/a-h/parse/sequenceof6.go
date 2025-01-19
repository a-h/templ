package parse

type sequenceOf6Parser[A, B, C, D, E, F any] struct {
	A Parser[A]
	B Parser[B]
	C Parser[C]
	D Parser[D]
	E Parser[E]
	F Parser[F]
}

func (p sequenceOf6Parser[A, B, C, D, E, F]) Parse(in *Input) (match Tuple6[A, B, C, D, E, F], ok bool, err error) {
	start := in.Index()
	match.A, ok, err = p.A.Parse(in)
	if err != nil || !ok {
		return
	}
	match.B, ok, err = p.B.Parse(in)
	if err != nil || !ok {
		in.Seek(start)
		return
	}
	match.C, ok, err = p.C.Parse(in)
	if err != nil || !ok {
		in.Seek(start)
		return
	}
	match.D, ok, err = p.D.Parse(in)
	if err != nil || !ok {
		in.Seek(start)
		return
	}
	match.E, ok, err = p.E.Parse(in)
	if err != nil || !ok {
		in.Seek(start)
		return
	}
	match.F, ok, err = p.F.Parse(in)
	if err != nil || !ok {
		in.Seek(start)
		return
	}
	return
}

func SequenceOf6[A, B, C, D, E, F any](a Parser[A], b Parser[B], c Parser[C], d Parser[D], e Parser[E], f Parser[F]) Parser[Tuple6[A, B, C, D, E, F]] {
	return sequenceOf6Parser[A, B, C, D, E, F]{
		A: a,
		B: b,
		C: c,
		D: d,
		E: e,
		F: f,
	}
}
