package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

func newDocTypeParser() docTypeParser {
	return docTypeParser{}
}

type docTypeParser struct {
}

var doctypeStartParser = parse.StringInsensitive("<!doctype ")

func (p docTypeParser) Parse(pi parse.Input) parse.Result {
	var r DocType

	from := NewPositionFromInput(pi)
	dtr := doctypeStartParser(pi)
	if dtr.Error != nil && dtr.Error != io.EOF {
		return dtr
	}
	if !dtr.Success {
		return parse.Failure("docTypeParser", nil)
	}

	// Once a doctype has started, take everything until the end.
	tagOpen := parse.Rune('<')
	tagClose := parse.Rune('>')
	dtr = parse.StringUntil(parse.Or(tagClose, tagOpen))(pi)
	if dtr.Error != nil && dtr.Error != io.EOF {
		return dtr
	}
	if !dtr.Success {
		return parse.Failure("docTypeParser", newParseError("unclosed DOCTYPE", from, NewPositionFromInput(pi)))
	}
	r.Value = dtr.Item.(string)

	// Clear the final '>'.
	from = NewPositionFromInput(pi)
	dtr = tagClose(pi)
	if dtr.Error != nil && dtr.Error != io.EOF {
		return dtr
	}
	if !dtr.Success {
		return parse.Failure("docTypeParser", newParseError("unclosed DOCTYPE", from, NewPositionFromInput(pi)))
	}

	return parse.Success("docTypeParser", r, nil)
}
