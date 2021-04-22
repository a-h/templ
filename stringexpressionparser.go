package templ

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

func (p stringExpressionParser) Parse(pi parse.Input) parse.Result {
	// Check the prefix first.
	prefixResult := parse.String("{%= ")(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we've seen a string expression prefix, read until the tag end.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(tagEnd)(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
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
	if te := tagEnd(pi); !te.Success {
		return te
	}

	return parse.Success("stringExpressionParser", r, nil)
}
