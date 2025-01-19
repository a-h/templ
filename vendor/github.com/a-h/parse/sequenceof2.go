package parse

type sequenceOf2Parser[A, B any] struct {
	A Parser[A]
	B Parser[B]
}

func (p sequenceOf2Parser[A, B]) Parse(in *Input) (match Tuple2[A, B], ok bool, err error) {
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
	return
}

func SequenceOf2[A, B any](a Parser[A], b Parser[B]) Parser[Tuple2[A, B]] {
	return sequenceOf2Parser[A, B]{
		A: a,
		B: b,
	}
}
