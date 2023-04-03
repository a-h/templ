package generator

import (
	"io"
	"strings"

	"github.com/a-h/templ/parser/v2"
)

func NewRangeWriter(w io.Writer) *RangeWriter {
	return &RangeWriter{
		w: w,
	}
}

type RangeWriter struct {
	Current parser.Position
	w       io.Writer
}

func (rw *RangeWriter) WriteIndent(level int, s string) (r parser.Range, err error) {
	_, err = rw.Write(strings.Repeat("\t", level))
	if err != nil {
		return
	}
	return rw.Write(s)
}

func (rw *RangeWriter) Write(s string) (r parser.Range, err error) {
	r.From = parser.Position{
		Index: rw.Current.Index,
		Line:  rw.Current.Line,
		Col:   rw.Current.Col,
	}
	var n int
	for _, c := range s {
		rw.Current.Col++
		if c == '\n' {
			rw.Current.Line++
			rw.Current.Col = 0
		}
		n, err = io.WriteString(rw.w, string(c))
		rw.Current.Index += int64(n)
		if err != nil {
			return r, err
		}
	}
	r.To = rw.Current
	return r, err
}
