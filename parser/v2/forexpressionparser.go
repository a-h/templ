package parser

import (
	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

var forExpression parse.Parser[Node] = forExpressionParser{}

type forExpressionParser struct{}

func (forExpressionParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	var r ForExpression
	start := pi.Index()

	// Strip leading whitespace and look for `for `.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		return r, false, err
	}
	if !peekPrefix(pi, "for ") {
		pi.Seek(start)
		return r, false, nil
	}

	// Parse the Go for expression.
	if r.Expression, err = parseGo("for", pi, goexpression.For); err != nil {
		return r, false, err
	}

	// Eat " {\n".
	if _, ok, err = parse.All(openBraceWithOptionalPadding, parse.NewLine).Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return r, false, err
	}

	// Node contents.
	tnp := newTemplateNodeParser(closeBraceWithOptionalPadding, "for expression closing brace")
	var nodes Nodes
	if nodes, ok, err = tnp.Parse(pi); err != nil || !ok {
		err = parse.Error("for: expected nodes, but none were found", pi.Position())
		return
	}
	r.Children = nodes.Nodes

	// Read the required closing brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("for: "+unterminatedMissingEnd, pi.Position())
		return
	}

	return r, true, nil
}
