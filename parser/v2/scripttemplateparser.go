package parser

import (
	"github.com/a-h/parse"
)

var scriptTemplateParser = parse.Func(func(pi *parse.Input) (r ScriptTemplate, ok bool, err error) {
	start := pi.Position()

	// Parse the name.
	var se scriptExpression
	if se, ok, err = scriptExpressionParser.Parse(pi); err != nil || !ok {
		pi.Seek(start.Index)
		return
	}
	r.Name = se.Name
	r.Parameters = se.Parameters

	// Read code expression.
	var e Expression
	if e, ok, err = exp.Parse(pi); err != nil || !ok {
		pi.Seek(start.Index)
		return
	}
	r.Value = e.Value

	// Try for }
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("script template: missing closing brace", pi.Position())
		return
	}

	r.Range = NewRange(start, pi.Position())

	return r, true, nil
})

// script Func() {
type scriptExpression struct {
	Name       Expression
	Parameters Expression
}

var scriptExpressionNameParser = ExpressionOf(parse.StringFrom(
	parse.Letter,
	parse.StringFrom(parse.AtMost(1000, parse.Any(parse.Letter, parse.ZeroToNine))),
))

var scriptExpressionParser = parse.Func(func(pi *parse.Input) (r scriptExpression, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.String("script ").Parse(pi); err != nil || !ok {
		return
	}

	// Once we have the prefix, we must have a name and parameters.
	// Read the name of the function.
	if r.Name, ok, err = scriptExpressionNameParser.Parse(pi); err != nil || !ok {
		err = parse.Error("script expression: invalid name", pi.Position())
		return
	}

	// Eat the open bracket.
	if _, ok, err = openBracket.Parse(pi); err != nil || !ok {
		err = parse.Error("script expression: parameters missing open bracket", pi.Position())
		return
	}

	// Read the parameters.
	// p Person, other Other, t thing.Thing)
	if r.Parameters, ok, err = ExpressionOf(parse.StringUntil(closeBracket)).Parse(pi); err != nil || !ok {
		err = parse.Error("script expression: parameters missing close bracket", pi.Position())
		return
	}

	// Eat ") {".
	if _, ok, err = expressionFuncEnd.Parse(pi); err != nil || !ok {
		err = parse.Error("script expression: unterminated (missing ') {')", pi.Position())
		return
	}

	// Expect a newline.
	if _, ok, err = parse.NewLine.Parse(pi); err != nil || !ok {
		err = parse.Error("script expression: missing terminating newline", pi.Position())
		return
	}

	return r, true, nil
})
