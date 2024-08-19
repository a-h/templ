package parser

import (
	"unicode"

	"github.com/a-h/parse"
)

var tagTemplOrNewLine = parse.Any(parse.Rune('<'), parse.Rune('{'), parse.Rune('}'), parse.String("\r\n"), parse.Rune('\n'))

var textParser = parse.Func(func(pi *parse.Input) (n Node, ok bool, err error) {
	from := pi.Position()

	// Read until a tag or templ expression opens.
	var t Text
	if t.Value, ok, err = parse.StringUntil(tagTemplOrNewLine).Parse(pi); err != nil || !ok {
		return
	}
	if isWhitespace(t.Value) {
		return t, false, nil
	}
	if _, ok = pi.Peek(1); !ok {
		err = parse.Error("textParser: unterminated text, expected tag open, templ expression open, or newline", from)
		return
	}
	t.Range = NewRange(from, pi.Position())

	// Elide any void element closing tags.
	if _, _, err = voidElementCloser.Parse(pi); err != nil {
		return
	}

	// Parse trailing whitespace.
	ws, _, err := parse.Whitespace.Parse(pi)
	if err != nil {
		return t, false, err
	}
	t.TrailingSpace, err = NewTrailingSpace(ws)
	if err != nil {
		return t, false, err
	}

	return t, true, nil
})

func isWhitespace(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
