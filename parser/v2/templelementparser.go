package parser

import (
	"fmt"
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

func (p templElementExpressionParser) Parse(pi *parse.Input) (r TemplElementExpression, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.Rune('@').Parse(pi); err != nil || !ok {
		return
	}

	// Parse the identifier.
	if r.Expression, ok, err = Must(templElementStartExpression, "templ element: found start '@' but expression was not closed").Parse(pi); err != nil || !ok {
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
	if r.Children, ok, err = Must[[]Node](np, fmt.Sprintf("@%s: expected nodes, but none were found", r.Expression.Value)).Parse(pi); err != nil || !ok {
		return
	}

	// Read the required closing brace.
	if _, ok, err = Must(closeBraceWithOptionalPadding, fmt.Sprintf("@%s: missing end (expected '}')", r.Expression.Value)).Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
}

var templElementExpression templElementExpressionParser
