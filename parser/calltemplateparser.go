package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// newCallTemplateExpressionParser creates a new callTemplateExpressionParser.
func newCallTemplateExpressionParser() callTemplateExpressionParser {
	return callTemplateExpressionParser{}
}

var callTemplateExpressionStartParser = parse.Or(parse.String("{%! "), parse.String("{%!"))

type callTemplateExpressionParser struct{}

func (p callTemplateExpressionParser) Parse(pi parse.Input) parse.Result {
	var r CallTemplateExpression

	// Check the prefix first.
	prefixResult := callTemplateExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have a prefix, we must have an expression that returns a template, followed by a tagEnd.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.Or(expressionEnd, newLine))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's no tagEnd or newLine, which is an error.
	if !pr.Success {
		return parse.Failure("callTemplateExpressionParser", newParseError("call: unterminated (missing closing ' %}')", from, NewPositionFromInput(pi)))
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// Eat " %}".
	from = NewPositionFromInput(pi)
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("callTemplateExpressionParser", newParseError("call: unterminated (missing closing ' %}')", from, NewPositionFromInput(pi)))
	}

	return parse.Success("callTemplate", r, nil)
}
