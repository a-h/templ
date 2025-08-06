package parser

import (
	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

var forExpression parse.Parser[Node] = forExpressionParser{}

type forExpressionParser struct{}

func (forExpressionParser) Parse(pi *parse.Input) (n Node, matched bool, err error) {
	r := &ForExpression{}
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
		return r, true, err
	}

	// Eat " {\n".
	_, matched, err = parse.All(openBraceWithOptionalPadding, parse.NewLine).Parse(pi)
	if err != nil {
		return r, true, err
	}
	if !matched {
		err = parse.Error("for: "+unterminatedMissingCurly, pi.PositionAt(start))
		return r, true, err
	}

	// Node contents.
	tnp := newTemplateNodeParser(closeBraceWithOptionalPadding, "for expression closing brace")
	var nodes Nodes
	if nodes, matched, err = tnp.Parse(pi); err != nil || !matched {
		// If we got any nodes, take them, because the LSP might want to use them.
		r.Children = nodes.Nodes
		return r, true, parse.Error("for: expected nodes, but none were found", pi.Position())
	}
	r.Children = nodes.Nodes

	// Read the required closing brace.
	if _, matched, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !matched {
		return r, true, parse.Error("for: "+unterminatedMissingEnd, pi.Position())
	}

	return r, true, nil
}
