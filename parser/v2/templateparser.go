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

var templateExpressionParser = parse.Func(func(pi *parse.Input) (r templateExpression, matched bool, err error) {
	start := pi.Index()

	if !peekPrefix(pi, "templ ") {
		return r, false, nil
	}

	// Once we have the prefix, everything to the brace is Go.
	// e.g.
	// templ (x []string) Test() {
	// becomes:
	// func (x []string) Test() templ.Component {
	if _, r.Expression, err = parseTemplFuncDecl(pi); err != nil {
		return r, true, err
	}

	// Eat " {\n".
	if _, matched, err = parse.All(openBraceWithOptionalPadding, parse.StringFrom(parse.Optional(parse.NewLine))).Parse(pi); err != nil || !matched {
		return r, true, parse.Error("templ: malformed templ expression, expected `templ functionName() {`", pi.PositionAt(start))
	}

	return r, true, nil
})

const (
	unterminatedMissingCurly = `unterminated (missing closing '{\n') - to escape "for", "if", "switch" etc. with braces, e.g. '{ "for" }' - https://templ.guide/syntax-and-usage/statements#ifswitchfor-within-text`
	unterminatedMissingEnd   = `missing end (expected '}') - https://templ.guide/syntax-and-usage/statements#ifswitchfor-within-text`
)

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

var rawElements = parse.Any(styleElement, scriptElement)

var templateNodeSkipParsers = []parse.Parser[Node]{
	voidElementCloser, // </br>, </img> etc. - should be ignored.
}

var templateNodeParsers = []parse.Parser[Node]{
	docType,                // <!DOCTYPE html>
	htmlComment,            // <!--
	goComment,              // // or /*
	rawElements,            // <text>, <>, or <style> element (special behaviour - contents are not parsed).
	element,                // <a>, <br/> etc.
	ifExpression,           // if {}
	forExpression,          // for {}
	switchExpression,       // switch {}
	callTemplateExpression, // {! TemplateName(a, b, c) }
	templElementExpression, // @TemplateName(a, b, c) { <div>Children</div> }
	childrenExpression,     // { children... }
	goCode,                 // {{ myval := x.myval }}
	stringExpression,       // { "abc" }
	whitespaceExpression,   // { " " }
	textParser,             // anything &amp; everything accepted...
}

func (p templateNodeParser[T]) Parse(pi *parse.Input) (op Nodes, matched bool, err error) {
outer:
	for {
		// Check if we've reached the end.
		if p.until != nil {
			start := pi.Index()
			_, didMatchUntilParser, err := p.until.Parse(pi)
			if err != nil {
				return op, false, err
			}
			if didMatchUntilParser {
				pi.Seek(start)
				return op, true, nil
			}
		}

		// Skip any nodes that we don't care about.
		for _, p := range templateNodeSkipParsers {
			_, didMatchSkipParser, err := p.Parse(pi)
			if err != nil {
				return op, false, err
			}
			if didMatchSkipParser {
				continue outer
			}
		}

		// Attempt to parse a node.
		// Loop through the parsers and try to parse a node.
		var didMatchTemplateNode bool
		for _, p := range templateNodeParsers {
			var node Node
			node, didMatchTemplateNode, err = p.Parse(pi)
			if err != nil {
				// Even if there's an error, we might have a node that the LSP can use,
				// so return it.
				if node != nil {
					op.Nodes = append(op.Nodes, node)
				}
				return op, true, err
			}
			if didMatchTemplateNode {
				op.Nodes = append(op.Nodes, node)
				break
			}
		}
		if didMatchTemplateNode {
			continue
		}

		if p.until == nil {
			// In this case, we're just reading as many nodes as we can until we can't read any more.
			// If we've reached here, we couldn't find a node.
			// The element parser checks the final node returned to make sure it's the expected close tag.
			break
		}

		err = UntilNotFoundError{
			ParseError: parse.Error(fmt.Sprintf("%v not found", p.untilName), pi.Position()),
		}
		return op, true, err
	}

	return op, true, nil
}

type UntilNotFoundError struct {
	parse.ParseError
}
