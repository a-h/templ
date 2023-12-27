package generator

import (
	"io"
	"strconv"
	"strings"

	"github.com/a-h/templ/parser/v2"
)

func NewRangeWriter(w io.Writer) *RangeWriter {
	return &RangeWriter{
		w: w,
	}
}

type RangeWriter struct {
	Current   parser.Position
	inLiteral bool
	w         io.Writer

	// Extract strings
	extractStrings bool
	strings        []string
}

func (rw *RangeWriter) closeLiteral(indent int) (r parser.Range, err error) {
	rw.inLiteral = false
	if rw.extractStrings {
		rw.strings = append(rw.strings, "")
	} else {
		_, err = rw.write("\")\n")
		if err != nil {
			return
		}
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

		if rw.extractStrings {
			index := len(rw.strings)
			if _, err = rw.WriteIndent(level, "templ_7745c5c3_Err = templ.WriteExtractedString(templ_7745c5c3_Buffer, "+strconv.Itoa(index)+")\n"); err != nil {
				return
			}
		} else {
			if _, err = rw.WriteIndent(level, `_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("`); err != nil {
				return
			}
		}
	}

	if rw.extractStrings {
		rw.strings[len(rw.strings)-1] += s
	} else {
		_, err = rw.write(s)
		if err != nil {
			return
		}
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
