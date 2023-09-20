package parser

import (
	"fmt"

	"github.com/a-h/parse"
)

// TemplateExpression.

// TemplateExpression.
// templ Func(p Parameter) {
// templ (data Data) Func(p Parameter) {
// templ (data []string) Func(p Parameter) {
type templateExpression struct {
	Expression Expression
}

var templateExpressionStartParser = parse.String("templ ")

var templateExpressionParser = parse.Func(func(pi *parse.Input) (r templateExpression, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = templateExpressionStartParser.Parse(pi); err != nil || !ok {
		return
	}

	// Once we have the prefix, everything to the brace at the end of the line is Go.
	// e.g.
	// templ (x []string) Test() {
	// becomes:
	// func (x []string) Test() templ.Component {

	// Once we've got a prefix, read until {\n.
	until := parse.All(openBraceWithOptionalPadding, parse.NewLine)
	msg := "templ: malformed templ expression, expected `templ functionName() {`"
	if r.Expression, ok, err = Must(ExpressionOf(parse.StringUntil(until)), msg).Parse(pi); err != nil || !ok {
		return
	}

	// Eat " {\n".
	if _, ok, err = Must(until, msg).Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
})

// Template node (element, call, if, switch, for, whitespace etc.)
func newTemplateNodeParser[TUntil any](until parse.Parser[TUntil], untilName string) templateNodeParser[TUntil] {
	return templateNodeParser[TUntil]{
		until:     until,
		untilName: untilName,
	}
}

type templateNodeParser[TUntil any] struct {
	until     parse.Parser[TUntil]
	untilName string
}

var rawElements = parse.Any[RawElement](styleElement, scriptElement)

var comment = parse.Any[Comment](htmlComment)

func (p templateNodeParser[T]) Parse(pi *parse.Input) (op []Node, ok bool, err error) {
	for {
		// Check if we've reached the end.
		if p.until != nil {
			start := pi.Index()
			_, ok, err = p.until.Parse(pi)
			if err != nil {
				return
			}
			if ok {
				pi.Seek(start)
				return op, true, nil
			}
		}

		// Try for valid nodes.
		// Try for a doctype.
		// <!DOCTYPE html>
		var docTypeNode DocType
		docTypeNode, ok, err = docTypeParser.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			op = append(op, docTypeNode)
			continue
		}

		// Try for a comment.
		// <!--
		var commentNode Comment
		commentNode, ok, err = comment.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			op = append(op, commentNode)
			continue
		}

		// Try for a raw <text>, <>, or <style> element (special behaviour - contents are not parsed).
		var rawElementNode RawElement
		rawElementNode, ok, err = rawElements.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			op = append(op, rawElementNode)
			continue
		}

		// Try for an element.
		// <a>, <br/> etc.
		var elementNode Element
		elementNode, ok, err = element.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			op = append(op, elementNode)
			continue
		}

		// Try for an if expression.
		// if {}
		var ifNode IfExpression
		ifNode, ok, err = ifExpression.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			op = append(op, ifNode)
			continue
		}

		// Try for a for expression.
		// for {}
		var forNode ForExpression
		forNode, ok, err = forExpression.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			op = append(op, forNode)
			continue
		}

		// Try for a switch expression.
		// switch {}
		var switchNode SwitchExpression
		switchNode, ok, err = switchExpression.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			op = append(op, switchNode)
			continue
		}

		// Try for a call template expression.
		// {! TemplateName(a, b, c) }
		var cteNode CallTemplateExpression
		cteNode, ok, err = callTemplateExpression.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			op = append(op, cteNode)
			continue
		}

		// Try for a templ element expression.
		// <!TemplateName(a, b, c) />
		var templElementNode TemplElementExpression
		templElementNode, ok, err = templElementExpression.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			op = append(op, templElementNode)
			continue
		}

		// Try for a children element expression.
		// { children... }
		var childrenExpressionNode ChildrenExpression
		childrenExpressionNode, ok, err = childrenExpression.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			op = append(op, childrenExpressionNode)
			continue
		}

		// Try for a string expression.
		// { "abc" }
		// { strings.ToUpper("abc") }
		var stringExpressionNode StringExpression
		stringExpressionNode, ok, err = stringExpression.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			op = append(op, stringExpressionNode)
			continue
		}

		// Eat any whitespace.
		var ws string
		if ws, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
			return
		}
		if ok && len(ws) > 0 {
			op = append(op, Whitespace{Value: ws})
			continue
		}

		// Try for text.
		// anything &amp; everything accepted...
		var text Text
		if text, ok, err = textParser.Parse(pi); err != nil {
			return
		}
		if ok && len(text.Value) > 0 {
			op = append(op, text)
			continue
		}

		if p.until == nil {
			// In this case, we're just reading as many nodes as we can until we can't read any more.
			// If we've reached here, we couldn't find a node.
			// The element parser checks the final node returned to make sure it's the expected close tag.
			break
		}

		err = fmt.Errorf("%v not found", p.untilName)
		return
	}

	return op, true, nil
}
