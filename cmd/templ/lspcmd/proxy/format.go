package proxy

import (
	"bytes"
	"fmt"
	"os"

	"github.com/a-h/templ/parser/v2"
	"github.com/natefinch/atomic"
)

func format(fileName string) (err error) {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", fileName, err)
	}
	t, err := parser.ParseString(string(contents))
	if err != nil {
		return fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	w := new(bytes.Buffer)
	err = t.Write(w)
	if err != nil {
		return fmt.Errorf("%s formatting error: %w", fileName, err)
	}
	if string(contents) == w.String() {
		return nil
	}
	err = atomic.WriteFile(fileName, w)
	if err != nil {
		return fmt.Errorf("%s file write error: %w", fileName, err)
	}
	return
}
