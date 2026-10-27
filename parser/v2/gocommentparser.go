package parser

import (
	"github.com/a-h/parse"
)

var (
	goSingleLineCommentStart = parse.String("//")
	goSingleLineCommentEnd   = parse.Any(parse.NewLine, parse.EOF[string]())
	goSingleLineCommentBody  = parse.StringUntil(goSingleLineCommentEnd)
)

type goSingleLineCommentParser struct{}

var goSingleLineComment = goSingleLineCommentParser{}

func (p goSingleLineCommentParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	// Comment start.
	start := pi.Position()
	if _, ok, err = goSingleLineCommentStart.Parse(pi); err != nil || !ok {
		return
	}
	// Once we've got the comment start sequence, parse anything until the end
	// sequence as the comment contents.
	c := &GoComment{}
	if c.Contents, ok, err = goSingleLineCommentBody.Parse(pi); err != nil || !ok {
		err = parse.Error("expected end comment literal '\n' not found", pi.Position())
		return
	}
	// Return the comment.
	c.Multiline = false
	c.Range = NewRange(start, pi.Position())
	return c, true, nil
}

func (p goSingleLineCommentParser) Match(in *parse.Input) (ok bool, err error) {
	return parse.Match(p, in)
}

var (
	goMultiLineCommentStart = parse.String("/*")
	goMultiLineCommentEnd   = parse.String("*/")
	goMultiLineCommentBody  = parse.StringUntil(goMultiLineCommentEnd)
)

type goMultiLineCommentParser struct{}

var goMultiLineComment = goMultiLineCommentParser{}

func (p goMultiLineCommentParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	// Comment start.
	start := pi.Position()
	if _, ok, err = goMultiLineCommentStart.Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got the comment start sequence, parse anything until the end
	// sequence as the comment contents.
	c := &GoComment{}
	if c.Contents, ok, err = goMultiLineCommentBody.Parse(pi); err != nil || !ok {
		err = parse.Error("expected end comment literal '*/' not found", start)
		return
	}
	// Move past the end element.
	_, _, _ = goMultiLineCommentEnd.Parse(pi)
	// Return the comment.
	c.Multiline = true
	c.Range = NewRange(start, pi.Position())
	return c, true, nil
}

func (p goMultiLineCommentParser) Match(in *parse.Input) (ok bool, err error) {
	return parse.Match(p, in)
}

var goComment = parse.Any(goSingleLineComment, goMultiLineComment)
