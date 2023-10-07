package parser

import (
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

	return t, true, nil
})
