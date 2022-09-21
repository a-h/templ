package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

var forExpression forExpressionParser

type forExpressionParser struct {
}

var forExpressionStartParser = parse.String("for ")

func (p forExpressionParser) Parse(pi parse.Input) parse.Result {
	var r ForExpression

	// Check the prefix first.
	prefixResult := forExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we've got a prefix, read until {\n.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.All(parse.WithStringConcatCombiner, openBraceWithOptionalPadding, newLine))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's no {\n, which is an error.
	if !pr.Success {
		return parse.Failure("forExpressionParser", newParseError("for: unterminated (missing closing '{\n')", from, NewPositionFromInput(pi)))
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// Eat " {".
	from = NewPositionFromInput(pi)
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("forExpressionParser", newParseError("for: unterminated (missing closing '{')", from, NewPositionFromInput(pi)))
	}

	// Once we've had the start of a for block, we must conclude the block.

	// Eat newline.
	if lb := newLine(pi); lb.Error != nil {
		return lb
	}

	// Node contents.
	from = NewPositionFromInput(pi)
	pr = newTemplateNodeParser(parse.Or(elseExpression.Parse, closeBraceWithOptionalPadding), "else expression or closing brace").Parse(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's a problem in the template nodes.
	if !pr.Success {
		return parse.Failure("forExpressionParser", newParseError("for: expected nodes, but none were found", from, NewPositionFromInput(pi)))
	}
	r.Children = pr.Item.([]Node)

	// Read the required closing brace.
	if ie := closeBraceWithOptionalPadding(pi); !ie.Success {
		return parse.Failure("forExpressionParser", newParseError("for: missing end (expected '}')", from, NewPositionFromInput(pi)))
	}

	return parse.Success("for", r, nil)
}


