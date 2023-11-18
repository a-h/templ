package parser

import (
	"github.com/a-h/parse"
)

var ifExpression ifExpressionParser

var untilElseIfElseOrEnd = parse.Any(StripType(elseIfExpression), StripType(elseExpression), StripType(closeBraceWithOptionalPadding))

type ifExpressionParser struct{}

func (ifExpressionParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.String("if ").Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got a prefix, read until {\n.
	// If there's no match, there's no {\n, which is an error.
	var r IfExpression
	until := parse.All(openBraceWithOptionalPadding, parse.NewLine)
	if r.Expression, ok, err = ExpressionOf(parse.StringUntil(until)).Parse(pi); err != nil || !ok {
		err = parse.Error("if: "+unterminatedMissingCurly, pi.Position())
		return
	}

	// Eat " {\n".
	if _, ok, err = until.Parse(pi); err != nil || !ok {
		err = parse.Error("if: "+unterminatedMissingCurly, pi.Position())
		return
	}

	// Once we've had the start of an if block, we must conclude the block.

	// Read the 'Then' nodes.
	// If there's no match, there's a problem in the template nodes.
	np := newTemplateNodeParser(untilElseIfElseOrEnd, "else expression or closing brace")
	var thenNodes Nodes
	if thenNodes, ok, err = np.Parse(pi); err != nil || !ok {
		err = parse.Error("if: expected nodes, but none were found", pi.Position())
		return
	}
	r.Then = thenNodes.Nodes
	r.Diagnostics = append(r.Diagnostics, thenNodes.Diagnostics...)

	// Read the optional 'ElseIf' Nodes.
	if r.ElseIfs, _, err = parse.ZeroOrMore(elseIfExpression).Parse(pi); err != nil {
		return
	}

	// Read the optional 'Else' Nodes.
	var elseNodes Nodes
	if elseNodes, _, err = elseExpression.Parse(pi); err != nil {
		return
	}
	r.Else = elseNodes.Nodes
	r.Diagnostics = append(r.Diagnostics, elseNodes.Diagnostics...)

	// Read the required closing brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("if: "+unterminatedMissingEnd, pi.Position())
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
		closeBrace,
		parse.OptionalWhitespace,
		parse.String("else if"),
		parse.Whitespace).Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got a prefix, read until {\n.
	// If there's no match, there's no {\n, which is an error.
	until := parse.All(openBraceWithOptionalPadding, parse.NewLine)
	if r.Expression, ok, err = ExpressionOf(parse.StringUntil(until)).Parse(pi); err != nil || !ok {
		err = parse.Error("if: unterminated else if (missing closing '{\n')", pi.Position())
		return
	}

	// Eat " {\n".
	if _, ok, err = until.Parse(pi); err != nil || !ok {
		err = parse.Error("if: unterminated (missing closing '{')", pi.Position())
		return
	}

	// Once we've had the start of an if block, we must conclude the block.

	// Read the 'Then' nodes.
	// If there's no match, there's a problem in the template nodes.
	np := newTemplateNodeParser(untilElseIfElseOrEnd, "else expression or closing brace")
	var thenNodes Nodes
	if thenNodes, ok, err = np.Parse(pi); err != nil || !ok {
		err = parse.Error("if: expected nodes, but none were found", pi.Position())
		return
	}
	r.Then = thenNodes.Nodes
	r.Diagnostics = append(r.Diagnostics, thenNodes.Diagnostics...)

	return r, true, nil
}

var endElseParser = parse.All(
	parse.Rune('}'),
	parse.OptionalWhitespace,
	parse.String("else"),
	parse.OptionalWhitespace,
	parse.Rune('{'),
	parse.OptionalWhitespace)

var elseExpression parse.Parser[Nodes] = elseExpressionParser{}

type elseExpressionParser struct{}

func (elseExpressionParser) Parse(in *parse.Input) (r Nodes, ok bool, err error) {
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
