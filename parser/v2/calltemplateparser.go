package parser

import (
	"github.com/a-h/parse"
)

var callTemplateExpression callTemplateExpressionParser

var callTemplateExpressionStart = parse.Or(parse.String("{! "), parse.String("{!"))

type callTemplateExpressionParser struct{}

func (p callTemplateExpressionParser) Parse(pi *parse.Input) (r CallTemplateExpression, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = callTemplateExpressionStart.Parse(pi); err != nil || !ok {
		return
	}

	// Once we have a prefix, we must have an expression that returns a template.
	if r.Expression, ok, err = exp.Parse(pi); err != nil || !ok {
		return
	}

	// Eat the final brace.
	if _, ok, err = Must(closeBraceWithOptionalPadding, "call template expression: missing closing brace").Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
}
