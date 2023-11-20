package parser

import (
	"unicode"

	"github.com/a-h/parse"
)

var templElementStartExpressionParams = parse.StringFrom(
	parse.String("("),
	parse.StringFrom[Expression](functionArgsParser{
		startBracketCount: 1,
	}),
	parse.String(")"),
)

var templElementStartExpression = ExpressionOf(parse.StringFrom(
	parse.AtLeast(1, parse.StringFrom(
		parse.StringFrom(parse.Optional(parse.String("."))),
		parse.StringFrom(parse.Optional(parse.String("_"))),
		parse.RuneInRanges(unicode.Letter),
		parse.StringFrom(parse.AtMost(255, parse.RuneInRanges(unicode.Letter, unicode.Number))),
		parse.StringFrom(parse.Optional(templElementStartExpressionParams)),
	)),
))

type templElementExpressionParser struct{}

func (p templElementExpressionParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.Rune('@').Parse(pi); err != nil || !ok {
		return
	}

	// Parse the identifier.
	var r TemplElementExpression
	if r.Expression, ok, err = templElementStartExpression.Parse(pi); err != nil || !ok {
		err = parse.Error("templ element: found start '@' but expression was not closed", pi.Position())
		return
	}

	// Once we've got a start expression, check to see if there's an open brace for children. {\n.
	var hasOpenBrace bool
	_, hasOpenBrace, err = openBraceWithOptionalPadding.Parse(pi)
	if err != nil {
		return
	}
	if !hasOpenBrace {
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
	r.Diagnostics = nodes.Diagnostics

	// Read the required closing brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("@"+r.Expression.Value+": missing end (expected '}')", pi.Position())
		return
	}

	return r, true, nil
}

var templElementExpression templElementExpressionParser
