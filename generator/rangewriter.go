package generator

import (
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/a-h/templ/parser/v2"
)

func NewRangeWriter(w io.Writer) *RangeWriter {
	return &RangeWriter{
		w: w,
		literalWriter: &literalWriter{
			builder: &strings.Builder{},
		},
	}
}

type RangeWriter struct {
	Current   parser.Position
	inLiteral bool
	w         io.Writer

	// Extract strings.
	literalWriter *literalWriter
}

type literalWriter struct {
	index    int
	builder  *strings.Builder
	Literals []string
}

func (w *literalWriter) closeLiteral(indent int) string {
	w.index++
	var sb strings.Builder
	sb.WriteString(strings.Repeat("\t", indent))
	sb.WriteString(`templ_7745c5c3_Err = templ.WriteString(templ_7745c5c3_Buffer, `)
	sb.WriteString(strconv.Itoa(w.index))
	sb.WriteString(`, "`)
	literal := w.builder.String()
	defer w.builder.Reset()
	w.Literals = append(w.Literals, literal)
	sb.WriteString(literal)
	sb.WriteString(`")`)
	sb.WriteString("\n")
	return sb.String()
}

func (w *literalWriter) writeLiteral(s string) {
	w.builder.WriteString(s)
}

func (rw *RangeWriter) closeLiteral(indent int) (r parser.Range, err error) {
	rw.inLiteral = false
	if _, err := rw.write(rw.literalWriter.closeLiteral(indent)); err != nil {
		return r, err
	}

	err = rw.writeErrorHandler(indent)
	return
}

func (rw *RangeWriter) WriteIndent(level int, s string) (r parser.Range, err error) {
	if rw.inLiteral {
		if _, err = rw.closeLiteral(level); err != nil {
			return
		}
	}
	_, err = rw.write(strings.Repeat("\t", level))
	if err != nil {
		return
	}
	return rw.write(s)
}

func (rw *RangeWriter) WriteStringLiteral(level int, s string) (r parser.Range, err error) {
	rw.literalWriter.writeLiteral(s)
	rw.inLiteral = true
	return
}

func (rw *RangeWriter) Write(s string) (r parser.Range, err error) {
	if rw.inLiteral {
		if _, err = rw.closeLiteral(0); err != nil {
			return
		}
	}
	return rw.write(s)
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
