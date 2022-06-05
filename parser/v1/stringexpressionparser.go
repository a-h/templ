package parser

import (
	"fmt"
	"io"

	"github.com/a-h/lexical/parse"
)

// StringExpression.
func newStringExpressionParser() stringExpressionParser {
	return stringExpressionParser{}
}

type stringExpressionParser struct {
}

var stringExpressionStartParser = parse.Or(parse.String("{%= "), parse.String("{%="))

func (p stringExpressionParser) Parse(pi parse.Input) parse.Result {
	// Check the prefix first.
	prefixResult := stringExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we've seen a string expression prefix, read until the tag end.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(expressionEnd)(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return parse.Failure("stringExpressionParser", fmt.Errorf("stringExpressionParser: failed to read until tag end: %w", pr.Error))
	}
	// If there's no tag end, the string expression parser wasn't terminated.
	if !pr.Success {
		return parse.Failure("stringExpressionParser", newParseError("string expression not terminated", from, NewPositionFromInput(pi)))
	}

	// Success! Create the expression.
	to := NewPositionFromInput(pi)
	r := StringExpression{
		Expression: NewExpression(pr.Item.(string), from, to),
	}

	// Eat the tag end.
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("stringExpressionParser", newParseError("could not terminate string expression", from, NewPositionFromInput(pi)))
	}

	return parse.Success("stringExpressionParser", r, nil)
}
