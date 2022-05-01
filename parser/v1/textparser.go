package parser

import (
	"errors"
	"io"

	"github.com/a-h/lexical/parse"
)

func newTextParser() textParser {
	return textParser{}
}

type textParser struct {
}

var tagOrTempl = parse.Or(parse.Rune('<'), parse.String("{%"))

func (p textParser) Parse(pi parse.Input) parse.Result {
	from := NewPositionFromInput(pi)

	// Read until a tag or templ expression opens.
	dtr := parse.StringUntil(tagOrTempl)(pi)
	if dtr.Error != nil {
		if errors.Is(dtr.Error, io.EOF) {
			return parse.Failure("textParser", newParseError("textParser: unterminated text, expected tag open or templ expression open statement", from, NewPositionFromInput(pi)))
		}
		return dtr
	}
	s, ok := dtr.Item.(string)
	if !ok || len(s) == 0 {
		return parse.Failure("textParser", nil)
	}
	return parse.Success("textParser", Text{Value: s}, nil)
}
