package parser

import (
	"github.com/a-h/parse"
)

var ifExpression ifExpressionParser

type ifExpressionParser struct{}

func (ifExpressionParser) Parse(pi *parse.Input) (r IfExpression, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.String("if ").Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got a prefix, read until {\n.
	// If there's no match, there's no {\n, which is an error.
	if r.Expression, ok, err = Must(ExpressionOf(parse.StringUntil(parse.All(openBraceWithOptionalPadding, parse.NewLine))), "if: unterminated (missing closing '{\n')").Parse(pi); err != nil || !ok {
		return
	}

	// Eat " {\n".
	if _, ok, err = Must(parse.All(openBraceWithOptionalPadding, parse.NewLine), "if: unterminated (missing closing '{')").Parse(pi); err != nil || !ok {
		return
	}

	// Once we've had the start of an if block, we must conclude the block.

	// Read the 'Then' nodes.
	// If there's no match, there's a problem in the template nodes.
	np := newTemplateNodeParser(parse.Any(StripType(elseIfExpression), StripType(elseExpression), StripType(closeBraceWithOptionalPadding)), "else expression or closing brace")
	if r.Then, ok, err = Must[[]Node](np, "if: expected nodes, but none were found").Parse(pi); err != nil || !ok {
		return
	}

	// Read the optional 'ElseIf' Nodes.
	if r.ElseIfs, _, err = parse.ZeroOrMore(elseIfExpression).Parse(pi); err != nil {
		return
	}

	// Read the optional 'Else' Nodes.
	if r.Else, _, err = elseExpression.Parse(pi); err != nil {
		return
	}

	// Read the required closing brace.
	if _, ok, err = Must(closeBraceWithOptionalPadding, "if: missing end (expected '}')").Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
}

var elseIfExpression parse.Parser[ElseIfExpression] = elseIfExpressionParser{}

type elseIfExpressionParser struct{}

func (elseIfExpressionParser) Parse(pi *parse.Input) (r ElseIfExpression, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.All(
		parse.OptionalWhitespace,
		parse.Rune('}'),
		parse.OptionalWhitespace,
		parse.String("else if"),
		parse.Whitespace).Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got a prefix, read until {\n.
	// If there's no match, there's no {\n, which is an error.
	if r.Expression, ok, err = Must(ExpressionOf(parse.StringUntil(parse.All(openBraceWithOptionalPadding, parse.NewLine))), "if: unterminated else if (missing closing '{\n')").Parse(pi); err != nil || !ok {
		return
	}

	// Eat " {\n".
	if _, ok, err = Must(parse.All(openBraceWithOptionalPadding, parse.NewLine), "if: unterminated (missing closing '{')").Parse(pi); err != nil || !ok {
		return
	}

	// Once we've had the start of an if block, we must conclude the block.

	// Read the 'Then' nodes.
	// If there's no match, there's a problem in the template nodes.
	np := newTemplateNodeParser(parse.Any(StripType(elseIfExpression), StripType(elseExpression), StripType(closeBraceWithOptionalPadding)), "else expression or closing brace")
	if r.Then, ok, err = Must[[]Node](np, "if: expected nodes, but none were found").Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
}

var endElseParser = parse.All(
	parse.Rune('}'),
	parse.OptionalWhitespace,
	parse.String("else"),
	parse.OptionalWhitespace,
	parse.Rune('{'),
	parse.OptionalWhitespace)

var elseExpression parse.Parser[[]Node] = elseExpressionParser{}

type elseExpressionParser struct{}

func (elseExpressionParser) Parse(in *parse.Input) (r []Node, ok bool, err error) {
	start := in.Index()

	// } else {
	if _, ok, err = endElseParser.Parse(in); err != nil || !ok {
		in.Seek(start)
		return
	}

	// Else contents
	if r, ok, err = newTemplateNodeParser(closeBraceWithOptionalPadding, "else expression closing brace").Parse(in); err != nil || !ok {
		in.Seek(start)
		return
	}

	return r, true, nil
}
