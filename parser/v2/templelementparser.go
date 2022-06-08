package parser

import (
	"fmt"
	"io"

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
	prefixResult := templElementExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have a prefix, we must have an expression that returns a template.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.All(parse.WithStringConcatCombiner, optionalWhitespaceParser, templElementExpressionEndParser))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	if !pr.Success {
		return pr
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// Eat closer.
	endResult := templElementExpressionEndParser(pi)
	if !endResult.Success {
		return endResult
	}
	// Once we've got an open tag, the rest must be present.
	from = NewPositionFromInput(pi)
	tnpr := newTemplateNodeParser(nil).Parse(pi)
	if !tnpr.Success {
		if _, isParseError := tnpr.Error.(ParseError); isParseError {
			return tnpr
		}
		return parse.Failure("templElementParser", newParseError(fmt.Sprintf("<%s>: %v", r.Expression.Value, tnpr.Error), from, NewPositionFromInput(pi)))
	}
	if arr, isArray := tnpr.Item.([]Node); isArray {
		r.Children = append(r.Children, arr...)
	}

	// Close tag.
	ectpr := parse.String("</>")(pi)
	if !ectpr.Success {
		return parse.Failure("templElementParser", newParseError(fmt.Sprintf("<%s>: expected end tag not present or invalid tag contents", r.Expression.Value), from, NewPositionFromInput(pi)))
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
	var e Expression
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.All(parse.WithStringConcatCombiner, optionalWhitespaceParser, parse.Or(parse.String(">"), parse.String("/>"))))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's no {\n, which is an error.
	if !pr.Success {
		return parse.Failure("templElementParser", newParseError("templ element: unterminated (missing closing '/>')", from, NewPositionFromInput(pi)))
	}
	e = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))
	return parse.Success("templElementParser", e, nil)
}
func (p templSelfClosingElementExpressionParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asTemplElement,
		templElementExpressionStartParser,
		p.parseExpression,
		optionalWhitespaceParser,
		templElementExpressionCloseTagParser,
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
