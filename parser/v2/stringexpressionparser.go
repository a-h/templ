package parser

import (
	"github.com/a-h/parse"
)

var stringExpression = parse.Func(func(pi *parse.Input) (r StringExpression, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.Or(parse.String("{ "), parse.String("{")).Parse(pi); err != nil || !ok {
		return
	}

	// Once we have a prefix, we must have an expression that returns a string.
	if r.Expression, ok, err = exp.Parse(pi); err != nil || !ok {
		return
	}

	// }
	if _, ok, err = Must(closeBraceWithOptionalPadding, "string expression: missing close brace").Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
})
