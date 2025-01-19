package parse

type sequenceOf8Parser[A, B, C, D, E, F, G, H any] struct {
	A Parser[A]
	B Parser[B]
	C Parser[C]
	D Parser[D]
	E Parser[E]
	F Parser[F]
	G Parser[G]
	H Parser[H]
}

func (p sequenceOf8Parser[A, B, C, D, E, F, G, H]) Parse(in *Input) (match Tuple8[A, B, C, D, E, F, G, H], ok bool, err error) {
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
	return
}

func SequenceOf8[A, B, C, D, E, F, G, H any](a Parser[A], b Parser[B], c Parser[C], d Parser[D], e Parser[E], f Parser[F], g Parser[G], h Parser[H]) Parser[Tuple8[A, B, C, D, E, F, G, H]] {
	return sequenceOf8Parser[A, B, C, D, E, F, G, H]{
		A: a,
		B: b,
		C: c,
		D: d,
		E: e,
		F: f,
		G: g,
		H: h,
	}
}
