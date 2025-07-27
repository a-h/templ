package format

import (
	"bytes"
	"fmt"

	"github.com/a-h/templ/internal/imports"
	parser "github.com/a-h/templ/parser/v2"
	"github.com/a-h/templ/parser/v2/visitor"
)

func findScriptElements(e *parser.Element, scriptElementToDepth map[*parser.ScriptElement]int, depth int) {
loop:
	for _, child := range e.Children {
		switch child := child.(type) {
		case *parser.ScriptElement:
			scriptElementToDepth[child] = depth
			continue loop
		case *parser.Element:
			findScriptElements(child, scriptElementToDepth, depth+1)
		}
	}
}

// Templ formats templ source, returning the formatted output, whether it changed, and an error if any.
// The fileName is used for Go import processing, use an empty name if the source is not from a file.
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

	nodeFormatter := visitor.New()
	// Calculate the depth of each ScriptElement in the tree so that the formatting is properly indented.
	scriptElementToDepth := make(map[*parser.ScriptElement]int)
	nodeFormatter.Element = func(e *parser.Element) error {
		findScriptElements(e, scriptElementToDepth, 0)
		return nil
	}
	nodeFormatter.ScriptElement = func(se *parser.ScriptElement) error {
		depth := scriptElementToDepth[se]
		// There's _always_ a templ node prior to any HTML elements.
		depth++
		return ScriptElement(se, depth)
	}
	if err = nodeFormatter.VisitTemplateFile(t); err != nil {
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
