package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// IfExpression.
var ifExpression ifExpressionParser

type ifExpressionParser struct {
}

var ifExpressionStartParser = parse.String("if ")

func (p ifExpressionParser) asChildren(parts []interface{}) (result interface{}, ok bool) {
	if len(parts) == 0 {
		return []Node{}, true
	}
	return parts[0].([]Node), true
}

func (p ifExpressionParser) Parse(pi parse.Input) parse.Result {
	var r IfExpression

	// Check the prefix first.
	prefixResult := ifExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we've got a prefix, read until {\n.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.All(parse.WithStringConcatCombiner, openBrace, newLine))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's no {\n, which is an error.
	if !pr.Success {
		return parse.Failure("ifExpressionParser", newParseError("if: unterminated (missing closing '{\n')", from, NewPositionFromInput(pi)))
	}
	r.Expression = NewExpression("if "+pr.Item.(string)+"{", from, NewPositionFromInput(pi))

	// Eat " {".
	from = NewPositionFromInput(pi)
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("ifExpressionParser", newParseError("if: unterminated (missing closing '{')", from, NewPositionFromInput(pi)))
	}

	// Once we've had the start of an if block, we must conclude the block.

	// Eat optional newline.
	if lb := newLine(pi); lb.Error != nil {
		return lb
	}

	// Read the 'Then' nodes.
	from = NewPositionFromInput(pi)
	pr = newTemplateNodeParser(parse.Or(elseExpression.Parse, closeBraceWithOptionalPadding)).Parse(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's a problem in the template nodes.
	if !pr.Success {
		return parse.Failure("ifExpressionParser", newParseError("if: expected nodes, but none were found", from, NewPositionFromInput(pi)))
	}
	r.Then = pr.Item.([]Node)

	// Read the optional 'Else' Nodes.
	from = NewPositionFromInput(pi)
	pr = parse.Optional(p.asChildren, elseExpression.Parse)(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	r.Else = pr.Item.([]Node)

	// Read the required closing brace.
	if ie := closeBraceWithOptionalPadding(pi); !ie.Success {
		return parse.Failure("ifExpressionParser", newParseError("if: missing end (expected '}')", from, NewPositionFromInput(pi)))
	}

	return parse.Success("if", r, nil)
}

var elseExpression elseExpressionParser

type elseExpressionParser struct {
}

func (p elseExpressionParser) asElseExpression(parts []interface{}) (result interface{}, ok bool) {
	return parts[1].([]Node), true // the array of nodes from templateNodeParser
}

func (p elseExpressionParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asElseExpression,
		endElseParser,
		newTemplateNodeParser(closeBraceWithOptionalPadding).Parse, // else contents
	)(pi)
}

var endElseParser = parse.All(parse.WithStringConcatCombiner,
	parse.Rune('{'),
	optionalWhitespaceParser,
	parse.String("else"),
	optionalWhitespaceParser,
	parse.Rune('}'),
	optionalWhitespaceParser)
