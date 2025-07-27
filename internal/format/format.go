package format

import (
	"bytes"
	"fmt"

	"github.com/a-h/templ/internal/imports"
	parser "github.com/a-h/templ/parser/v2"
)

// Templ formats templ source, returning the formatted output, whether it changed, and an error if any.
func Templ(src []byte, fileName string) (output []byte, changed bool, err error) {
	t, err := parser.ParseString(string(src))
	if err != nil {
		return nil, false, err
	}
	t.Filepath = fileName
	t, err = imports.Process(t)
	if err != nil {
		return nil, false, err
	}
	w := new(bytes.Buffer)
	if err = t.Write(w); err != nil {
		return nil, false, fmt.Errorf("formatting error: %w", err)
	}
	out := w.Bytes()
	changed = !bytes.Equal(src, out)
	return out, changed, nil
}
