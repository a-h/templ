package parse

type sequenceOf4Parser[A, B, C, D any] struct {
	A Parser[A]
	B Parser[B]
	C Parser[C]
	D Parser[D]
}

func (p sequenceOf4Parser[A, B, C, D]) Parse(in *Input) (match Tuple4[A, B, C, D], ok bool, err error) {
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
	return
}

func SequenceOf4[A, B, C, D any](a Parser[A], b Parser[B], c Parser[C], d Parser[D]) Parser[Tuple4[A, B, C, D]] {
	return sequenceOf4Parser[A, B, C, D]{
		A: a,
		B: b,
		C: c,
		D: d,
	}
}
