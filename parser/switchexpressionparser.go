package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

var switchExpression switchExpressionParser

type switchExpressionParser struct {
}

func (p switchExpressionParser) asMany(parts []interface{}) (result interface{}, ok bool) {
	r := []CaseExpression{}

	for _, p := range parts {
		if ce, ok := p.(CaseExpression); ok {
			r = append(r, ce)
		} else {
			return nil, false
		}
	}

	return r, true
}

var switchExpressionStartParser = parse.String("switch ") // "switch "

func (p switchExpressionParser) Parse(pi parse.Input) parse.Result {
	var r SwitchExpression

	// Check the prefix first.
	prefixResult := switchExpressionStartParser(pi)
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
		return parse.Failure("switchExpressionParser", newParseError("switch: unterminated (missing closing '{\n')", from, NewPositionFromInput(pi)))
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// Eat " {".
	from = NewPositionFromInput(pi)
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("switchExpressionParser", newParseError("switch: unterminated (missing closing '{')", from, NewPositionFromInput(pi)))
	}

	// Once we've had the start of a switch block, we must conclude the block.

	// Eat optional newline.
	if lb := optionalWhitespaceParser(pi); lb.Error != nil {
		return lb
	}

	// Read the optional 'case' nodes.
	from = NewPositionFromInput(pi)
	pr = parse.Many(p.asMany, 0, -1, newCaseExpressionParser().Parse)(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	if pr.Success {
		r.Cases = pr.Item.([]CaseExpression)
	}

	// Read the required closing brace.
	if ie := closeBraceWithOptionalPadding(pi); !ie.Success {
		return parse.Failure("switchExpressionParser", newParseError("switch: missing end (expected '}')", from, NewPositionFromInput(pi)))
	}

	return parse.Success("switch", r, nil)
}

func newCaseExpressionParser() caseExpressionParser {
	return caseExpressionParser{}
}

type caseExpressionParser struct {
}

var caseExpressionStartParser = parse.All(parse.WithStringConcatCombiner,
	optionalWhitespaceAsString,
	parse.Or(parse.String("case "), parse.String("default")),
	parse.StringUntil(parse.String(":\n")),
	parse.String(":\n"),
)

func (p caseExpressionParser) Parse(pi parse.Input) parse.Result {
	var r CaseExpression

	// Check the prefix first.
	from := NewPositionFromInput(pi)
	prefixResult := caseExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}
	r.Expression = NewExpression(prefixResult.Item.(string), from, NewPositionFromInput(pi))

	// Read until the next case statement, default, or end of the block.
	from = NewPositionFromInput(pi)
	pr := newTemplateNodeParser(parse.Or(closeBraceWithOptionalPadding, caseExpressionStartParser)).Parse(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}

	// If there's no match, there's a problem in the template nodes.
	if !pr.Success {
		return parse.Failure("caseExpressionParser", newParseError("case: expected nodes, but none were found", from, NewPositionFromInput(pi)))
	}
	r.Children = pr.Item.([]Node)

	if lb := optionalWhitespaceParser(pi); lb.Error != nil {
		return lb
	}

	return parse.Success("case", r, nil)
}
