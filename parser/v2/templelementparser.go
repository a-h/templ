package parser

import (
	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

type templElementExpressionParser struct{}

func (p templElementExpressionParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.Rune('@').Parse(pi); err != nil || !ok {
		return
	}

	r := TemplElementExpression{
		// Default behavior is always a trailing space
		TrailingSpace: SpaceVertical,
	}

	// Parse the Go expression.
	if r.Expression, err = parseGo("templ element", pi, goexpression.TemplExpression); err != nil {
		return r, false, err
	}

	// Once we've got a start expression, check to see if there's an open brace for children. {\n.
	var hasOpenBrace bool
	_, hasOpenBrace, err = openBraceWithOptionalPadding.Parse(pi)
	if err != nil {
		return
	}
	if !hasOpenBrace {
		// Parse trailing whitespace after expression.
		if _, _, err := addTrailingSpace(&r, pi, true); err != nil {
			return r, false, err
		}

		return r, true, nil
	}

	// Once we've had the start of an element's children, we must conclude the block.

	// Node contents.
	np := newTemplateNodeParser(closeBraceWithOptionalPadding, "templ element closing brace")
	var nodes Nodes
	if nodes, ok, err = np.Parse(pi); err != nil || !ok {
		err = parse.Error("@"+r.Expression.Value+": expected nodes, but none were found", pi.Position())
		return
	}
	r.Children = nodes.Nodes

	// Read the required closing brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("@"+r.Expression.Value+": missing end (expected '}')", pi.Position())
		return
	}

	// Parse trailing whitespace after closing brace.
	if _, _, err := addTrailingSpace(&r, pi, true); err != nil {
		return r, false, err
	}

	return r, true, nil
}

var templElementExpression templElementExpressionParser
