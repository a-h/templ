package parser

import (
	"github.com/a-h/parse"
)

var stringExpression = parse.Func(func(pi *parse.Input) (n Node, matched bool, err error) {
	// Check the prefix first.
	if _, matched, err = parse.Or(parse.String("{ "), parse.String("{")).Parse(pi); err != nil || !matched {
		return
	}

	// Once we have a prefix, we must have an expression that returns a string, with optional err.
	r := &StringExpression{}
	if r.Expression, err = parseGoSliceArgs(pi); err != nil {
		// We return true because we should have completed the string expression, but didn't.
		// That means we found a node, but the node is invalid (has an error).
		return r, true, err
	}

	// Clear any optional whitespace.
	_, _, _ = parse.OptionalWhitespace.Parse(pi)

	// }
	if _, matched, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !matched {
		return r, true, parse.Error("string expression: missing close brace", pi.Position())
	}

	// Parse trailing whitespace.
	ws, _, err := parse.Whitespace.Parse(pi)
	if err != nil {
		return r, true, err
	}
	r.TrailingSpace, err = NewTrailingSpace(ws)
	if err != nil {
		return r, true, err
	}

	return r, true, nil
})
