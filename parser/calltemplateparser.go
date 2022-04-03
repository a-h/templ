package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// newCallTemplateExpressionParser creates a new callTemplateExpressionParser.
func newCallTemplateExpressionParser() callTemplateExpressionParser {
	return callTemplateExpressionParser{}
}

var callTemplateExpressionStartParser = parse.Or(parse.String("{! "), parse.String("{!"))

type callTemplateExpressionParser struct{}

func (p callTemplateExpressionParser) Parse(pi parse.Input) parse.Result {
	var r CallTemplateExpression

	// Check the prefix first.
	prefixResult := callTemplateExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have a prefix, we must have an expression that returns a template.
	from := NewPositionFromInput(pi)
	pr := exp.Parse(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	if !pr.Success {
		return pr
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// Eat the final brace.
	if pr, ok := chompBrace(pi); !ok {
		return pr
	}

	return parse.Success("callTemplate", r, nil)
}
