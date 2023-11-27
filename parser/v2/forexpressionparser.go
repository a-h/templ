package parser

import (
	"github.com/a-h/parse"
)

var forExpression parse.Parser[Node] = forExpressionParser{}

type forExpressionParser struct{}

func (_ forExpressionParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.String("for ").Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got a prefix, read until {\n.
	// If there's no match, there's no {\n, which is an error.
	from := pi.Position()
	var r ForExpression
	until := parse.All(openBraceWithOptionalPadding, parse.NewLine)
	var fexp string
	if fexp, ok, err = parse.StringUntil(until).Parse(pi); err != nil || !ok {
		err = parse.Error("for: "+unterminatedMissingCurly, pi.Position())
		return
	}
	r.Expression = NewExpression(fexp, from, pi.Position())

	// Eat " {".
	if _, ok, err = until.Parse(pi); err != nil || !ok {
		err = parse.Error("for: "+unterminatedMissingCurly, pi.Position())
		return
	}

	// Node contents.
	tnp := newTemplateNodeParser(closeBraceWithOptionalPadding, "for expression closing brace")
	var nodes Nodes
	if nodes, ok, err = tnp.Parse(pi); err != nil || !ok {
		err = parse.Error("for: expected nodes, but none were found", pi.Position())
		return
	}
	r.Children = nodes.Nodes
	r.Diagnostics = nodes.Diagnostics

	// Read the required closing brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("for: "+unterminatedMissingEnd, pi.Position())
		return
	}

	return r, true, nil
}
