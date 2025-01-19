package parse

type sequenceOf5Parser[A, B, C, D, E any] struct {
	A Parser[A]
	B Parser[B]
	C Parser[C]
	D Parser[D]
	E Parser[E]
}

func (p sequenceOf5Parser[A, B, C, D, E]) Parse(in *Input) (match Tuple5[A, B, C, D, E], ok bool, err error) {
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
	return
}

func SequenceOf5[A, B, C, D, E any](a Parser[A], b Parser[B], c Parser[C], d Parser[D], e Parser[E]) Parser[Tuple5[A, B, C, D, E]] {
	return sequenceOf5Parser[A, B, C, D, E]{
		A: a,
		B: b,
		C: c,
		D: d,
		E: e,
	}
}
