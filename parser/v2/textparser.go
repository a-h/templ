package parser

import (
	"strings"

	"github.com/a-h/parse"
)

var tagTemplOrNewLine = parse.Any(parse.Rune('<'), parse.Rune('{'), parse.Rune('}'), parse.Rune('\n'))

var textParser = parse.Func(func(pi *parse.Input) (t Text, ok bool, err error) {
	from := pi.Position()

	// Read until a tag or templ expression opens.
	if t.Value, ok, err = parse.StringUntil(tagTemplOrNewLine).Parse(pi); err != nil || !ok {
		return
	}
	if _, ok = pi.Peek(1); !ok {
		err = parse.Error("textParser: unterminated text, expected tag open, templ expression open, or newline", from)
		return
	}

	t.LineBreak = true

	v, ok := pi.Peek(maxIgnoreNewLineTagLength + 3)

	if ok {
		for _, i := range ignoreNewLinesTags {
			if strings.Contains(v, "<"+i+">") || strings.Contains(v, "</"+i+">") {
				t.LineBreak = false
				break
			}
		}
	}

	if line, _ := pi.Peek(1); line == "\n" {
		t.LineBreak = true
	}

	return t, true, nil
})
