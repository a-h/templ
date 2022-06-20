package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/a-h/lexical/parse"
)

var templElementExpressionStartParser = parse.Or(parse.String("<! "), parse.String("<!"))
var templElementExpressionEndParser = parse.Or(parse.String(" >"), parse.String(">"))
var templElementExpressionCloseTagParser = parse.Or(parse.String(" />"), parse.String("/>"))

var templBlockElementExpression templBlockElementExpressionParser

type templBlockElementExpressionParser struct{}

func (p templBlockElementExpressionParser) Parse(pi parse.Input) parse.Result {
	var r TemplElementExpression

	// Check the prefix first.
	prefixResult := parse.Rune('@')(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have a prefix, we must have an expression that returns a template.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.All(parse.WithStringConcatCombiner, openBraceWithOptionalPadding))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	if !pr.Success {
		return parse.Failure("templElementParser", newParseError("templ element: unterminated (missing closing '{\n')", from, NewPositionFromInput(pi)))
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	from = NewPositionFromInput(pi)
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("templElementParser", newParseError("templ element: unterminated (missing closing '{')", from, NewPositionFromInput(pi)))
	}

	// Once we've had the start of a for block, we must conclude the block.

	// Eat newline.
	if lb := newLine(pi); lb.Error != nil {
		return lb
	}

	// Node contents.
	from = NewPositionFromInput(pi)
	pr = newTemplateNodeParser(closeBraceWithOptionalPadding).Parse(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's a problem in the template nodes.
	if !pr.Success {
		return parse.Failure("templElementParser", newParseError(fmt.Sprintf("@%s: expected nodes, but none were found", r.Expression.Value), from, NewPositionFromInput(pi)))
	}

	r.Children = pr.Item.([]Node)

	// Read the required closing brace.
	if ie := closeBraceWithOptionalPadding(pi); !ie.Success {
		return parse.Failure("templElementParser", newParseError(fmt.Sprintf("@%s: missing end (expected '}')", r.Expression.Value), from, NewPositionFromInput(pi)))
	}
	return parse.Success("templElementParser", r, nil)
}

var templSelfClosingElementExpression templSelfClosingElementExpressionParser

type templSelfClosingElementExpressionParser struct{}

func (p templSelfClosingElementExpressionParser) asTemplElement(parts []interface{}) (result interface{}, ok bool) {
	return TemplElementExpression{
		Expression: parts[1].(Expression),
	}, true
}

func (p templSelfClosingElementExpressionParser) parseExpression(pi parse.Input) parse.Result {
	start := pi.Index()
	var e Expression
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(newLine)(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return parse.Failure("templElementParser", nil)
	}
	if strings.HasSuffix(strings.TrimSpace(pr.Item.(string)), "{") {
		rewind(pi, start)
		return parse.Failure("templElementParser", nil)
	}
	// If there's no match, there's no \n, which is an error.
	if !pr.Success {
		return parse.Failure("templElementParser", newParseError("templ element: unterminated (missing closing newline)", from, NewPositionFromInput(pi)))
	}
	e = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))
	return parse.Success("templElementParser", e, nil)
}
func (p templSelfClosingElementExpressionParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asTemplElement,
		parse.Rune('@'),
		p.parseExpression,
	)(pi)
}

var templElementExpression templElementExpressionParser

type templElementExpressionParser struct{}

func (p templElementExpressionParser) Parse(pi parse.Input) parse.Result {
	var r TemplElementExpression

	// Self closing.
	pr := templSelfClosingElementExpression.Parse(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	if pr.Success {
		r = pr.Item.(TemplElementExpression)
		return parse.Success("templElementParser", r, nil)
	}

	// Block.
	pr = templBlockElementExpression.Parse(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	if pr.Success {
		r = pr.Item.(TemplElementExpression)
		return parse.Success("templElementParser", r, nil)
	}

	return parse.Failure("templElementParser", nil)
}
