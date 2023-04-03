package parser

import (
	"github.com/a-h/parse"
)

var scriptTemplateParser = parse.Func(func(pi *parse.Input) (r ScriptTemplate, ok bool, err error) {
	start := pi.Index()

	// Parse the name.
	var se scriptExpression
	if se, ok, err = scriptExpressionParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}
	r.Name = se.Name
	r.Parameters = se.Parameters

	// Read code expression.
	var e Expression
	if e, ok, err = exp.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}
	r.Value = e.Value

	// Try for }
	if _, ok, err = Must(closeBraceWithOptionalPadding, "script template: missing closing brace").Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

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
	if r.Name, ok, err = Must(scriptExpressionNameParser, "script expression: invalid name").Parse(pi); err != nil || !ok {
		return
	}

	// Eat the open bracket.
	if _, ok, err = Must(parse.Rune('('), "script expression: parameters missing open bracket").Parse(pi); err != nil || !ok {
		return
	}

	// Read the parameters.
	// p Person, other Other, t thing.Thing)
	if r.Parameters, ok, err = Must(ExpressionOf(parse.StringUntil(parse.Rune(')'))), "script expression: parameters missing close bracket").Parse(pi); err != nil || !ok {
		return
	}

	// Eat ") {".
	if _, ok, err = Must(expressionFuncEnd, "script expression: unterminated (missing ') {')").Parse(pi); err != nil || !ok {
		return
	}

	// Expect a newline.
	if _, ok, err = Must(parse.NewLine, "script expression: missing terminating newline").Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
})
