package parser

import (
	"github.com/a-h/parse"
)

var goSingleLineCommentStart = parse.String("//")
var goSingleLineCommentEnd = parse.Any(parse.NewLine, parse.EOF[string]())

type goSingleLineCommentParser struct {
}

var goSingleLineComment = goSingleLineCommentParser{}

func (p goSingleLineCommentParser) Parse(pi *parse.Input) (c GoComment, ok bool, err error) {
	// Comment start.
	if _, ok, err = goSingleLineCommentStart.Parse(pi); err != nil || !ok {
		return
	}
	// Once we've got the comment start sequence, parse anything until the end
	// sequence as the comment contents.
	if c.Contents, ok, err = Must(parse.StringUntil(goSingleLineCommentEnd), "expected end comment literal '\n' not found").Parse(pi); err != nil || !ok {
		return
	}
	// Move past the end element.
	_, _, _ = goSingleLineCommentEnd.Parse(pi)
	// Return the comment.
	c.Multiline = false
	return c, true, nil
}

var goMultiLineCommentStart = parse.String("/*")
var goMultiLineCommentEnd = parse.String("*/")

type goMultiLineCommentParser struct {
}

var goMultiLineComment = goMultiLineCommentParser{}

func (p goMultiLineCommentParser) Parse(pi *parse.Input) (c GoComment, ok bool, err error) {
	// Comment start.
	if _, ok, err = goMultiLineCommentStart.Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got the comment start sequence, parse anything until the end
	// sequence as the comment contents.
	if c.Contents, ok, err = Must(parse.StringUntil(goMultiLineCommentEnd), "expected end comment literal '*/' not found").Parse(pi); err != nil || !ok {
		return
	}
	// Move past the end element.
	_, _, _ = goMultiLineCommentEnd.Parse(pi)
	// Return the comment.
	c.Multiline = true
	return c, true, nil
}

var goComment = parse.Any(goSingleLineComment, goMultiLineComment)
