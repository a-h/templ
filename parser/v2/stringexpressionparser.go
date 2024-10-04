package parser

import (
	"github.com/a-h/parse"
)

var stringExpression = parse.Func(func(pi *parse.Input) (n Node, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.Or(parse.String("{ "), parse.String("{")).Parse(pi); err != nil || !ok {
		return
	}

	// Once we have a prefix, we must have an expression that returns a string, with optional err.
	var r StringExpression
	if r.Expression, err = parseGoSliceArgs(pi); err != nil {
		return r, false, err
	}

	// Clear any optional whitespace.
	_, _, _ = parse.OptionalWhitespace.Parse(pi)

	// }
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("string expression: missing close brace", pi.Position())
		return
	}

	// Parse trailing whitespace.
	r.TrailingSpace, err = parseTrailingSpace(pi, false, false)
	if err != nil {
		return r, false, err
	}

	return r, true, nil
})
