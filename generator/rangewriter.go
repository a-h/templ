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
		w:             w,
		literalWriter: prodLiteralWriter{},
	}
}

type RangeWriter struct {
	Current   parser.Position
	inLiteral bool
	w         io.Writer

	// Extract strings.
	literalWriter literalWriter
}

type literalWriter interface {
	writeLiteral(inLiteral bool, s string) string
	closeLiteral(indent int) string
	literals() string
}

type watchLiteralWriter struct {
	index   int
	builder *strings.Builder
}

func (w *watchLiteralWriter) closeLiteral(indent int) string {
	w.index++
	w.builder.WriteString("\n")
	return ""
}

func (w *watchLiteralWriter) writeLiteral(inLiteral bool, s string) string {
	w.builder.WriteString(s)
	if inLiteral {
		return ""
	}

	return "templ_7745c5c3_Err = templ.WriteWatchModeString(templ_7745c5c3_Buffer, " + strconv.Itoa(w.index+1) + ")\n"
}

func (w *watchLiteralWriter) literals() string {
	return w.builder.String()
}

type prodLiteralWriter struct{}

func (prodLiteralWriter) closeLiteral(indent int) string {
	return "\")\n"
}

func (prodLiteralWriter) writeLiteral(inLiteral bool, s string) string {
	if inLiteral {
		return s
	}
	return `_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("` + s
}

func (prodLiteralWriter) literals() string {
	return ""
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
	if !rw.inLiteral {
		_, err = rw.write(strings.Repeat("\t", level))
		if err != nil {
			return
		}
	}

	if _, err := rw.write(rw.literalWriter.writeLiteral(rw.inLiteral, s)); err != nil {
		return r, err
	}

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
