package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

func newSwitchExpressionParser() switchExpressionParser {
	return switchExpressionParser{}
}

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

func (p switchExpressionParser) Parse(pi parse.Input) parse.Result {
	var r SwitchExpression

	// Check the prefix first.
	blockPrefix := "switch "
	prefixResult := parse.String("{% " + blockPrefix)(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we've got a prefix, we must have the switch expression, followed by a tagEnd.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.Or(tagEnd, newLine))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's no tagEnd or newLine, which is an error.
	if !pr.Success {
		return parse.Failure("switchExpressionParser", newParseError("switch: unterminated (missing closing ' %}')", from, NewPositionFromInput(pi)))
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// Eat " %}".
	from = NewPositionFromInput(pi)
	if te := tagEnd(pi); !te.Success {
		return parse.Failure("switchExpressionParser", newParseError("switch: unterminated (missing closing ' %}')", from, NewPositionFromInput(pi)))
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

	// Read the optional 'default' node.
	pr = newDefaultCaseExpressionParser().Parse(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	if pr.Success {
		r.Default = pr.Item.([]Node)
	}

	// Read the required "endswitch" statement.
	if ie := parse.String("{% endswitch %}")(pi); !ie.Success {
		return parse.Failure("switchExpressionParser", newParseError("switch: missing end (expected '{% endswitch %}')", from, NewPositionFromInput(pi)))
	}

	return parse.Success("switch", r, nil)
}

func newDefaultCaseExpressionParser() defaultCaseExpressionParser {
	return defaultCaseExpressionParser{}
}

type defaultCaseExpressionParser struct {
}

func (p defaultCaseExpressionParser) Parse(pi parse.Input) parse.Result {
	var r []Node

	// Start parsing if we have the required "default" statement.
	if ie := parse.String("{% default %}")(pi); !ie.Success {
		return parse.Failure("defaultCaseExpressionParser", nil)
	}

	from := NewPositionFromInput(pi)
	pr := newTemplateNodeParser(endDefaultParser).Parse(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's a problem in the template nodes.
	if !pr.Success {
		return parse.Failure("defaultCaseExpressionParser", newParseError("defaultCase: expected nodes, but none were found", from, NewPositionFromInput(pi)))
	}
	r = pr.Item.([]Node)

	// Read the required "enddefault" statement.
	if ie := endDefaultParser(pi); !ie.Success {
		return parse.Failure("defaultCaseExpressionParser", newParseError("defaultCase: missing end (expected '{% enddefault %}')", from, NewPositionFromInput(pi)))
	}

	if lb := optionalWhitespaceParser(pi); lb.Error != nil {
		return lb
	}

	return parse.Success("defaultCase", r, nil)
}

var endDefaultParser = parse.String("{% enddefault %}")

func newCaseExpressionParser() caseExpressionParser {
	return caseExpressionParser{}
}

type caseExpressionParser struct {
}

func (p caseExpressionParser) Parse(pi parse.Input) parse.Result {
	var r CaseExpression

	// Check the prefix first.
	blockPrefix := "case "
	prefixResult := parse.String("{% " + blockPrefix)(pi)
	if !prefixResult.Success {
		return prefixResult
	}
	// Once we've got a prefix, we must have the if expression, followed by a tagEnd.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.Or(tagEnd, newLine))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}

	// If there's no match, there's no tagEnd or newLine, which is an error.
	if !pr.Success {
		return parse.Failure("caseExpressionParser", newParseError("case: unterminated (missing closing ' %}')", from, NewPositionFromInput(pi)))
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// Eat " %}".
	from = NewPositionFromInput(pi)
	if te := tagEnd(pi); !te.Success {
		return parse.Failure("caseExpressionParser", newParseError("case: unterminated (missing closing ' %}')", from, NewPositionFromInput(pi)))
	}

	// Once we've had the start of a case block, we must conclude the block.

	// Eat optional newline.
	if lb := optionalWhitespaceParser(pi); lb.Error != nil {
		return lb
	}

	// Read the 'Then' nodes.
	from = NewPositionFromInput(pi)
	pr = newTemplateNodeParser(endCaseParser).Parse(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}

	// If there's no match, there's a problem in the template nodes.
	if !pr.Success {
		return parse.Failure("caseExpressionParser", newParseError("case: expected nodes, but none were found", from, NewPositionFromInput(pi)))
	}
	r.Children = pr.Item.([]Node)

	// Read the required "endif" statement.
	if ie := endCaseParser(pi); !ie.Success {
		return parse.Failure("caseExpressionParser", newParseError("if: missing end (expected '{% endcase %}')", from, NewPositionFromInput(pi)))
	}

	if lb := optionalWhitespaceParser(pi); lb.Error != nil {
		return lb
	}

	return parse.Success("case", r, nil)
}

var endCaseParser = parse.String("{% endcase %}")
