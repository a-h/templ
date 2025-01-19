package parse

type sequenceOf7Parser[A, B, C, D, E, F, G any] struct {
	A Parser[A]
	B Parser[B]
	C Parser[C]
	D Parser[D]
	E Parser[E]
	F Parser[F]
	G Parser[G]
}

func (p sequenceOf7Parser[A, B, C, D, E, F, G]) Parse(in *Input) (match Tuple7[A, B, C, D, E, F, G], ok bool, err error) {
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
	match.G, ok, err = p.G.Parse(in)
	if err != nil || !ok {
		in.Seek(start)
		return
	}
	return
}

func SequenceOf7[A, B, C, D, E, F, G any](a Parser[A], b Parser[B], c Parser[C], d Parser[D], e Parser[E], f Parser[F], g Parser[G]) Parser[Tuple7[A, B, C, D, E, F, G]] {
	return sequenceOf7Parser[A, B, C, D, E, F, G]{
		A: a,
		B: b,
		C: c,
		D: d,
		E: e,
		F: f,
		G: g,
	}
}
