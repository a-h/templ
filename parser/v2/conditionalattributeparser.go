package parser

import (
	"github.com/a-h/parse"
)

var conditionalAttributeParser = parse.Func(func(pi *parse.Input) (r ConditionalAttribute, ok bool, err error) {
	start := pi.Index()

	// Strip leading whitespace and look for `if `.
	if _, ok, err = parse.All(parse.OptionalWhitespace, parse.String("if ")).Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Once we've got a prefix, read until {\n.
	if r.Expression, ok, err = Must(ExpressionOf(parse.StringUntil(parse.All(openBraceWithOptionalPadding, parse.NewLine))), "attribute if: unterminated (missing closing '{\n')").Parse(pi); err != nil || !ok {
		return
	}

	// Eat " {\n".
	if _, ok, err = Must(parse.All(openBraceWithOptionalPadding, parse.NewLine), "attribute if: unterminated (missing closing '{')").Parse(pi); err != nil || !ok {
		return
	}

	// Read the 'Then' attributes.
	// If there's no match, there's a problem reading the attributes.
	if r.Then, ok, err = Must[[]Attribute](attributesParser{}, "attribute if: expected attributes in block, but none were found").Parse(pi); err != nil || !ok {
		return
	}

	// Read the optional 'Else' Nodes.
	if r.Else, _, err = attributeElseExpression.Parse(pi); err != nil {
		return
	}

	// Clear any optional whitespace.
	_, _, _ = parse.OptionalWhitespace.Parse(pi)

	// Read the required closing brace.
	if _, ok, err = Must(closeBraceWithOptionalPadding, "attribute if: missing end (expected '}')").Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
})

var attributeElseExpression parse.Parser[[]Attribute] = attributeElseExpressionParser{}

type attributeElseExpressionParser struct{}

func (attributeElseExpressionParser) Parse(in *parse.Input) (r []Attribute, ok bool, err error) {
	start := in.Index()

	// Strip any initial whitespace.
	_, _, _ = parse.OptionalWhitespace.Parse(in)

	// } else {
	var endElseParser = parse.All(
		parse.Rune('}'),
		parse.OptionalWhitespace,
		parse.String("else"),
		parse.OptionalWhitespace,
		parse.Rune('{'))
	if _, ok, err = endElseParser.Parse(in); err != nil || !ok {
		in.Seek(start)
		return
	}

	// Else contents
	if r, ok, err = Must[[]Attribute](attributesParser{}, "attribute if: expected attributes in else block, but none were found").Parse(in); err != nil || !ok {
		in.Seek(start)
		return
	}

	return r, true, nil
}
