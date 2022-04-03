package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// StringExpression.
func newStringExpressionParser() stringExpressionParser {
	return stringExpressionParser{}
}

type stringExpressionParser struct {
}

var stringExpressionStartParser = parse.Or(parse.String("{ "), parse.String("{"))

func (p stringExpressionParser) Parse(pi parse.Input) parse.Result {
	var r StringExpression
	// Check the prefix first.
	prefixResult := stringExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have a prefix, we must have an expression that returns a string.
	from := NewPositionFromInput(pi)
	pr := exp.Parse(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	if !pr.Success {
		return pr
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// }
	if pr, ok := chompBrace(pi); !ok {
		return pr
	}

	return parse.Success("stringExpressionParser", r, nil)
}
