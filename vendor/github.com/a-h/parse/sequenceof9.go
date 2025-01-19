package parse

type sequenceOf9Parser[A, B, C, D, E, F, G, H, I any] struct {
	A Parser[A]
	B Parser[B]
	C Parser[C]
	D Parser[D]
	E Parser[E]
	F Parser[F]
	G Parser[G]
	H Parser[H]
	I Parser[I]
}

func (p sequenceOf9Parser[A, B, C, D, E, F, G, H, I]) Parse(in *Input) (match Tuple9[A, B, C, D, E, F, G, H, I], ok bool, err error) {
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
	match.H, ok, err = p.H.Parse(in)
	if err != nil || !ok {
		in.Seek(start)
		return
	}
	match.I, ok, err = p.I.Parse(in)
	if err != nil || !ok {
		in.Seek(start)
		return
	}
	return
}

func SequenceOf9[A, B, C, D, E, F, G, H, I any](a Parser[A], b Parser[B], c Parser[C], d Parser[D], e Parser[E], f Parser[F], g Parser[G], h Parser[H], i Parser[I]) Parser[Tuple9[A, B, C, D, E, F, G, H, I]] {
	return sequenceOf9Parser[A, B, C, D, E, F, G, H, I]{
		A: a,
		B: b,
		C: c,
		D: d,
		E: e,
		F: f,
		G: g,
		H: h,
		I: i,
	}
}
