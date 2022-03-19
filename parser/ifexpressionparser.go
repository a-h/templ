package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// IfExpression.
func newIfExpressionParser() ifExpressionParser {
	return ifExpressionParser{}
}

type ifExpressionParser struct {
}

var ifExpressionStartParser = createStartParser("if")

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

	// Once we've got a prefix, we must have the if expression, followed by a tagEnd.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.Or(expressionEnd, newLine))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's no tagEnd or newLine, which is an error.
	if !pr.Success {
		return parse.Failure("ifExpressionParser", newParseError("if: unterminated (missing closing ' %}')", from, NewPositionFromInput(pi)))
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// Eat " %}".
	from = NewPositionFromInput(pi)
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("ifExpressionParser", newParseError("if: unterminated (missing closing ' %}')", from, NewPositionFromInput(pi)))
	}

	// Once we've had the start of an if block, we must conclude the block.

	// Eat optional newline.
	if lb := newLine(pi); lb.Error != nil {
		return lb
	}

	// Read the 'Then' nodes.
	from = NewPositionFromInput(pi)
	eep := newElseExpressionParser()
	pr = newTemplateNodeParser(parse.Or(eep.Parse, endIfParser)).Parse(pi)
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
	pr = parse.Optional(p.asChildren, eep.Parse)(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	r.Else = pr.Item.([]Node)

	// Read the required "endif" statement.
	if ie := endIfParser(pi); !ie.Success {
		return parse.Failure("ifExpressionParser", newParseError("if: missing end (expected '{% endif %}')", from, NewPositionFromInput(pi)))
	}

	return parse.Success("if", r, nil)
}

func newElseExpressionParser() elseExpressionParser {
	return elseExpressionParser{}
}

type elseExpressionParser struct {
}

func (p elseExpressionParser) asElseExpression(parts []interface{}) (result interface{}, ok bool) {
	return parts[1].([]Node), true // the array of nodes from templateNodeParser
}

func (p elseExpressionParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asElseExpression,
		endElseParser,
		newTemplateNodeParser(endIfParser).Parse, // else contents
	)(pi)
}

var endElseParser = createEndParser("else") // {% else %}
var endIfParser = createEndParser("endif")  // {% endif %}
