package parser

import (
	"github.com/a-h/parse"
)

var fallthroughString = parse.String("fallthrough")

var fallthroughExpression = parse.Func(func(pi *parse.Input) (n Node, ok bool, err error) {
	start := pi.Position()
	if _, ok, err = fallthroughString.Parse(pi); err != nil || !ok {
		return
	}

	// Eat terminating newline.
	_, _, _ = parse.ZeroOrMore(parse.RuneIn(" \t")).Parse(pi)
	_, ok, err = parse.NewLine.Parse(pi)
	if err != nil || !ok {
		err = parse.Error("expected newline after fallthrough", pi.Position())
		return nil, true, err
	}

	return &Fallthrough{
		Range: NewRange(start, pi.Position()),
	}, true, nil
})
