package parser

import (
	"github.com/a-h/parse"
)

var htmlCommentStart = parse.String("<!--")
var htmlCommentEnd = parse.String("--")

type commentParser struct {
}

var htmlComment = commentParser{}

func (p commentParser) Parse(pi *parse.Input) (c Comment, ok bool, err error) {
	// Comment start.
	if _, ok, err = htmlCommentStart.Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got the comment start sequence, parse anything until the end
	// sequence as the comment contents.
	if c.Contents, ok, err = Must(parse.StringUntil(htmlCommentEnd), "expected end comment literal '-->' not found").Parse(pi); err != nil || !ok {
		return
	}
	// Cut the end element.
	_, _, _ = htmlCommentEnd.Parse(pi)

	// Cut the gt.
	if _, ok, err = Must(gt, "comment contains invalid sequence '--'").Parse(pi); err != nil || !ok {
		return
	}

	return c, true, nil
}
