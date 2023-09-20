package parser

import (
	"fmt"

	"github.com/a-h/parse"
)

type commentParser struct {
	start parse.Parser[string]
	end   parse.Parser[string]
}

var htmlComment = commentParser{
	start: parse.String("<!--"),
	end:   parse.String("-->"),
}

func (p commentParser) Parse(pi *parse.Input) (c Comment, ok bool, err error) {
	// Comment start.
	if _, ok, err = p.start.Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got the comment start sequence, parse anything until the end
	// sequence as the comment contents.
	if c.Contents, ok, err = Must(parse.StringUntil(p.end), fmt.Sprintf("expected end comment sequence not found")).Parse(pi); err != nil || !ok {
		return
	}
	// Cut the end element.
	_, _, _ = p.end.Parse(pi)

	return c, true, nil
}
