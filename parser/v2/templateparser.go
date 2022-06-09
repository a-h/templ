package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// TemplateExpression.

// TemplateExpression.
// templ Func(p Parameter) {
// templ (data Data) Func(p Parameter) {
// templ (data []string) Func(p Parameter) {
type templateExpression struct {
	Expression Expression
}

func newTemplateExpressionParser() templateExpressionParser {
	return templateExpressionParser{}
}

type templateExpressionParser struct {
}

var templateExpressionStartParser = parse.String("templ ")

func (p templateExpressionParser) Parse(pi parse.Input) parse.Result {
	var r templateExpression

	// Check the prefix first.
	prefixResult := templateExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have the prefix, everything to the brace at the end of the line is Go.
	// e.g.
	// templ (x []string) Test() {
	// becomes:
	// func (x []string) Test() templ.Component {

	// Once we've got a prefix, read until {\n.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.All(parse.WithStringConcatCombiner, openBraceWithOptionalPadding, newLine))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, there's no {\n, which is an error.
	if !pr.Success {
		return parse.Failure("templateExpressionParser", newParseError("templ: unterminated (missing closing '{\n')", from, NewPositionFromInput(pi)))
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// Eat " {".
	from = NewPositionFromInput(pi)
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("templateExpressionParser", newParseError("templ: unterminated (missing closing '{')", from, NewPositionFromInput(pi)))
	}

	// Eat required newline.
	if lb := newLine(pi); lb.Error != nil {
		return lb
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

var rawElements = parse.Any(styleElement.Parse, scriptElement.Parse)

func (p templateNodeParser) Parse(pi parse.Input) parse.Result {
	op := make([]Node, 0)
	for {
		var pr parse.Result

		// Check if we've reached the end.
		if p.until != nil {
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
		}

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

		// Try for a raw <text>, <>, or <style> element (special behaviour - contents are not parsed).
		pr = rawElements(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for an element.
		// <a>, <br/> etc.
		pr = element.Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for an if expression.
		// if {}
		pr = ifExpression.Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for a for expression.
		// for {}
		pr = forExpression.Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for a switch expression.
		// switch {}
		pr = switchExpression.Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for a call template expression.
		// {! TemplateName(a, b, c) }
		pr = callTemplateExpression.Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for a templ element expression.
		// <!TemplateName(a, b, c) />
		pr = templElementExpression.Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for a children element expression.
		// { children... }
		pr = childrenExpression(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			op = append(op, pr.Item.(Node))
			continue
		}

		// Try for a string expression.
		// { "abc" }
		// { strings.ToUpper("abc") }
		pr = stringExpression.Parse(pi)
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

		if p.until == nil {
			// In this case, we're just reading as many nodes as we can.
			// The element parser checks the final node returned to make sure it's the expected close tag.
			break
		}
	}

	return parse.Success("templateNodeParser", op, nil)
}
