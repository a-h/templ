package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// ForExpression.
func newForExpressionParser() forExpressionParser {
	return forExpressionParser{}
}

type forExpressionParser struct {
}

var forExpressionStartParser = createStartParser("for")

func (p forExpressionParser) Parse(pi parse.Input) parse.Result {
	var r ForExpression

	// Check the prefix first.
	prefixResult := forExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we've had "{% for ", we're expecting a loop Go expression, followed by a tagEnd.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.Or(expressionEnd, newLine))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's no tagEnd or newLine, which is an error.
	if !pr.Success {
		return parse.Failure("forExpressionParser", newParseError("for: unterminated (missing closing ' %}')", from, NewPositionFromInput(pi)))
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// Eat " %}".
	from = NewPositionFromInput(pi)
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("forExpressionParser", newParseError("for: unterminated (missing closing ' %}')", from, NewPositionFromInput(pi)))
	}

	// Eat required newline.
	from = NewPositionFromInput(pi)
	lb := newLine(pi)
	if lb.Error != nil {
		return lb
	}
	if !lb.Success {
		return parse.Failure("forExpressionParser", newParseError("for: missing newline after closing ' %}'", from, NewPositionFromInput(pi)))
	}

	// Once we've had the start of a for block, we must conclude the block.

	// Node contents.
	from = NewPositionFromInput(pi)
	pr = newTemplateNodeParser(endForParser).Parse(pi)
	if pr.Error != nil {
		return pr
	}
	if !lb.Success {
		return parse.Failure("forExpressionParser", newParseError("for: contents not found", from, NewPositionFromInput(pi)))
	}
	r.Children = pr.Item.([]Node)

	// Eat the required "endfor".
	if pr = endForParser(pi); !pr.Success {
		return parse.Failure("forExpressionParser", newParseError("for: missing end (expected '{% endfor %}')", from, NewPositionFromInput(pi)))
	}

	return parse.Success("for", r, nil)
}

var endForParser = createEndParser("endfor") // {% endfor %}
