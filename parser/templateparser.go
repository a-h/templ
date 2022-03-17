package parser

import (
	"io"
	"strings"

	"github.com/a-h/lexical/parse"
)

// TemplateExpression.

// TemplateExpression.
// {% templ Func(p Parameter) %}
type templateExpression struct {
	Name       Expression
	Parameters Expression
}

func newTemplateExpressionParser() templateExpressionParser {
	return templateExpressionParser{}
}

var templateNameParser = parse.All(parse.WithStringConcatCombiner,
	parse.Letter,
	parse.Many(parse.WithStringConcatCombiner, 0, 1000, parse.Any(parse.Letter, parse.ZeroToNine)),
)

type templateExpressionParser struct {
}

var templateExpressionStartParser = createStartParser("templ")

func (p templateExpressionParser) Parse(pi parse.Input) parse.Result {
	var r templateExpression

	// Check the prefix first.
	prefixResult := templateExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have the prefix, we must have a name and parameters.
	// Read the name of the function.
	from := NewPositionFromInput(pi)
	pr := templateNameParser(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, the name wasn't correctly terminated.
	if !pr.Success {
		return parse.Failure("templateExpressionParser", newParseError("template expression: invalid name", from, NewPositionFromInput(pi)))
	}
	// Remove the final "(" from the position.
	to := NewPositionFromInput(pi)
	to.Col--
	to.Index--
	r.Name = NewExpression(strings.TrimSuffix(pr.Item.(string), "("), from, to)

	// Eat the left bracket.
	if lb := parse.Rune('(')(pi); !lb.Success {
		return parse.Failure("templateExpressionParser", newParseError("template expression: parameters missing open bracket", from, NewPositionFromInput(pi)))
	}

	// Read the parameters.
	from = NewPositionFromInput(pi)
	pr = parse.StringUntil(parse.Rune(')'))(pi) // p Person, other Other, t thing.Thing)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, the name wasn't correctly terminated.
	if !pr.Success {
		return parse.Failure("templateExpressionParser", newParseError("template expression: parameters missing close bracket", from, NewPositionFromInput(pi)))
	}
	r.Parameters = NewExpression(strings.TrimSuffix(pr.Item.(string), ")"), from, NewPositionFromInput(pi))

	// Eat ") %}".
	from = NewPositionFromInput(pi)
	if lb := expressionFuncEnd(pi); !lb.Success {
		return parse.Failure("templateExpressionParser", newParseError("template expression: unterminated (missing ') %}')", from, NewPositionFromInput(pi)))
	}

	// Expect a newline.
	from = NewPositionFromInput(pi)
	if lb := newLine(pi); !lb.Success {
		return parse.Failure("templateExpressionParser", newParseError("template expression: missing terminating newline", from, NewPositionFromInput(pi)))
	}

	return parse.Success("templateExpressionParser", r, nil)
}

// Template node (element, call, if, switch, for, whitespace etc.)
func newTemplateNodeParser(until parse.Function) templateNodeParser {
	return templateNodeParser{
		until: until,
	}
}

type templateNodeParser struct {
	until parse.Function
}

func (p templateNodeParser) Parse(pi parse.Input) parse.Result {
	op := make([]Node, 0)
	for {
		var pr parse.Result

		// Try for a doctype.
		// <!DOCTYPE html>
		pr = newDocTypeParser().Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for an element.
		// <a>, <br/> etc.
		pr = newElementParser().Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for a string expression.
		// {%= "abc" %}
		// {%= strings.ToUpper("abc") %}
		pr = newStringExpressionParser().Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for an if expression.
		// if {}
		pr = newIfExpressionParser().Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for a switch expression.
		// switch {}
		pr = newSwitchExpressionParser().Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for a for expression.
		// for {}
		pr = newForExpressionParser().Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for a call template expression.
		// {%! TemplateName(a, b, c) %}
		pr = newCallTemplateExpressionParser().Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Eat any whitespace.
		pr = optionalWhitespaceParser(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success && len(pr.Item.(Whitespace).Value) > 0 {
			op = append(op, pr.Item.(Whitespace))
			continue
		}

		// Try for text.
		// anything &amp; everything accepted...
		pr = newTextParser().Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success && len(pr.Item.(Text).Value) > 0 {
			op = append(op, pr.Item.(Text))
			continue
		}

		// Check if we've reached the end.
		if p.until == nil {
			// In this case, we're just reading as many nodes as we can.
			// The element parser checks the final node returned to make sure it's the expected close tag.
			break
		} else {
			start := pi.Index()
			pr = p.until(pi)
			if pr.Error != nil {
				return pr
			}
			if pr.Success {
				if err := rewind(pi, start); err != nil {
					return parse.Failure("templateNodeParser", err)
				}
				return parse.Success("templateNodeParser", op, nil)
			}

			pos := NewPositionFromInput(pi)
			return parse.Failure("templateNodeParser", newParseError("template: unexpected token", pos, pos))
		}
	}

	return parse.Success("templateNodeParser", op, nil)
}
