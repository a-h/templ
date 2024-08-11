package generator

import (
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/a-h/templ/parser/v2"
)

func NewRangeWriter(w io.Writer, literalVarName string) *RangeWriter {
	return &RangeWriter{
		w:              w,
		literalVarName: literalVarName,
		buf:            &strings.Builder{},
	}
}

type RangeWriter struct {
	Current   parser.Position
	inLiteral bool
	w         io.Writer

	// Variable name used to store a slice of all string literals.
	literalVarName string

	// literal buffer.
	buf *strings.Builder
	// literal index.
	index int
	// literal string slice.
	ss []string
}

func (rw *RangeWriter) closeLiteral(indent int) (r parser.Range, err error) {
	rw.inLiteral = false
	rw.ss = append(rw.ss, rw.buf.String())
	defer func() {
		rw.index++
		rw.buf.Reset()
	}()
	s := "_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" + rw.literalVarName + "[" + strconv.Itoa(rw.index) + "]) // " + strconv.Quote(rw.buf.String()) + "\n"
	if _, err := rw.WriteIndent(indent, s); err != nil {
		return r, err
	}
	err = rw.writeErrorHandler(indent)
	return
}

func (rw *RangeWriter) WriteIndent(level int, s string) (r parser.Range, err error) {
	if err = rw.Close(); err != nil {
		return r, err
	}
	_, err = rw.write(strings.Repeat("\t", level))
	if err != nil {
		return
	}
	return rw.write(s)
}

func (rw *RangeWriter) WriteStringLiteral(level int, s string) (r parser.Range, err error) {
	rw.buf.WriteString(s)
	rw.inLiteral = true
	return
}

func (rw *RangeWriter) Write(s string) (r parser.Range, err error) {
	if err = rw.Close(); err != nil {
		return r, err
	}
	return rw.write(s)
}

func (rw *RangeWriter) Close() error {
	if rw.inLiteral {
		_, err := rw.closeLiteral(0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rw *RangeWriter) write(s string) (r parser.Range, err error) {
	r.From = parser.Position{
		Index: rw.Current.Index,
		Line:  rw.Current.Line,
		Col:   rw.Current.Col,
	}
	utf8Bytes := make([]byte, 4)
	for _, c := range s {
		rlen := utf8.EncodeRune(utf8Bytes, c)
		rw.Current.Col += uint32(rlen)
		if c == '\n' {
			rw.Current.Line++
			rw.Current.Col = 0
		}
		_, err = rw.w.Write(utf8Bytes[:rlen])
		rw.Current.Index += int64(rlen)
		if err != nil {
			return r, err
		}
	}
	r.To = rw.Current
	return r, err
}

func (rw *RangeWriter) writeErrorHandler(indentLevel int) (err error) {
	_, err = rw.WriteIndent(indentLevel, "if templ_7745c5c3_Err != nil {\n")
	if err != nil {
		return err
	}
	indentLevel++
	_, err = rw.WriteIndent(indentLevel, "return templ_7745c5c3_Err\n")
	if err != nil {
		return err
	}
	indentLevel--
	_, err = rw.WriteIndent(indentLevel, "}\n")
	if err != nil {
		return err
	}
	return err
}
