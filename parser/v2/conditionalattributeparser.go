package parser

import (
	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

var conditionalAttribute parse.Parser[ConditionalAttribute] = conditionalAttributeParser{}

type conditionalAttributeParser struct{}

func (conditionalAttributeParser) Parse(pi *parse.Input) (r ConditionalAttribute, ok bool, err error) {
	start := pi.Index()

	// Strip leading whitespace and look for `if `.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		return
	}
	if !peekPrefix(pi, "if ") {
		pi.Seek(start)
		return
	}

	// Parse the Go if expression.
	if r.Expression, err = parseGo("if attribute", pi, goexpression.If); err != nil {
		return
	}

	// Eat " {\n".
	if _, ok, err = openBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("attribute if: unterminated (missing closing '{\n')", pi.PositionAt(start))
		return
	}
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		return
	}

	// Read the 'Then' attributes.
	// If there's no match, there's a problem reading the attributes.
	if r.Then, ok, err = (attributesParser{}).Parse(pi); err != nil || !ok {
		err = parse.Error("attribute if: expected attributes in block, but none were found", pi.Position())
		return
	}

	if len(r.Then) == 0 {
		err = parse.Error("attribute if: invalid content or no attributes were found in the if block", pi.Position())
		return
	}

	// Read the optional 'Else' Nodes.
	if r.Else, ok, err = attributeElseExpression.Parse(pi); err != nil {
		return
	}
	if ok && len(r.Else) == 0 {
		err = parse.Error("attribute if: invalid content or no attributes were found in the else block", pi.Position())
		return
	}

	// Clear any optional whitespace.
	_, _, _ = parse.OptionalWhitespace.Parse(pi)

	// Read the required closing brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("attribute if: missing end (expected '}')", pi.Position())
		return
	}

	return r, true, nil
}

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
	if r, ok, err = (attributesParser{}).Parse(in); err != nil || !ok {
		err = parse.Error("attribute if: expected attributes in else block, but none were found", in.Position())
		return
	}

	return r, true, nil
}
