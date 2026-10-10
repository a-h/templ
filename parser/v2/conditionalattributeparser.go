package parser

import (
	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

var conditionalAttribute parse.Parser[*ConditionalAttribute] = conditionalAttributeParser{}

type conditionalAttributeParser struct{}

func (conditionalAttributeParser) Parse(pi *parse.Input) (r *ConditionalAttribute, ok bool, err error) {
	start := pi.Index()

	// Strip leading whitespace and look for `if `.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		return
	}
	attrStart := pi.Index()
	if !peekPrefix(pi, "if ") {
		pi.Seek(start)
		return
	}

	// Parse the Go if expression.
	r = &ConditionalAttribute{}
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

	// Read the optional 'ElseIf' attributes.
	if r.ElseIfs, _, err = parse.ZeroOrMore(attributeElseIfExpression).Parse(pi); err != nil {
		return
	}

	// Read the optional 'Else' attributes.
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
	r.Range = NewRange(pi.PositionAt(attrStart), pi.Position())

	return r, true, nil
}

var attributeElseIfExpression parse.Parser[ConditionalElseIfAttribute] = attributeElseIfExpressionParser{}

type attributeElseIfExpressionParser struct{}

func (attributeElseIfExpressionParser) Parse(pi *parse.Input) (r ConditionalElseIfAttribute, matched bool, err error) {
	start := pi.Index()

	// Check the prefix first.
	if _, matched, err = parse.All(parse.OptionalWhitespace, closeBrace, parse.OptionalWhitespace, parse.String("else if")).Parse(pi); err != nil || !matched {
		pi.Seek(start)
		return
	}

	// Rewind to the start of the `if` statement.
	pi.Seek(pi.Index() - 2)
	// Parse the Go if expression.
	if r.Expression, err = parseGo("attribute else if", pi, goexpression.If); err != nil {
		return
	}

	// Eat " {".
	if _, matched, err = openBraceWithOptionalPadding.Parse(pi); err != nil || !matched {
		err = parse.Error("attribute else if: unterminated (missing closing '{\n')", pi.PositionAt(start))
		return
	}
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		return
	}

	// Read the 'Then' attributes.
	if r.Then, matched, err = (attributesParser{}).Parse(pi); err != nil || !matched {
		err = parse.Error("attribute if: expected attributes in else if block, but none were found", pi.Position())
		return
	}

	if len(r.Then) == 0 {
		err = parse.Error("attribute if: invalid content or no attributes were found in the else if block", pi.Position())
		return
	}

	r.Range = NewRange(pi.PositionAt(start), pi.Position())
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
