package parser

import (
	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

var ifExpression ifExpressionParser

var untilElseIfElseOrEnd = parse.Any(StripType(elseIfExpression), StripType(elseExpression), StripType(closeBraceWithOptionalPadding))

type ifExpressionParser struct{}

// Parsers return:
//  as much of a Node as they can
//  matched=true if the start of a complete, incomplete or invalid node was found, e.g. "if " or "{ "
//  err if an error occurred or a node was started and not completed

func (ifExpressionParser) Parse(pi *parse.Input) (n Node, matched bool, err error) {
	start := pi.Index()

	if !peekPrefix(pi, "if ") {
		return nil, false, nil
	}

	// Parse the Go if expression using the Go parser.
	r := &IfExpression{}
	if r.Expression, err = parseGo("if", pi, goexpression.If); err != nil {
		return r, true, err
	}

	// Eat " {\n".
	if _, matched, err = parse.All(openBraceWithOptionalPadding, parse.NewLine).Parse(pi); err != nil || !matched {
		err = parse.Error("if: "+unterminatedMissingCurly, pi.PositionAt(start))
		return r, true, err
	}

	// Once we've had the start of an if block, we must conclude the block.

	// Read the 'Then' nodes.
	// If there's no match, there's a problem in the template nodes.
	np := newTemplateNodeParser(untilElseIfElseOrEnd, "else expression or closing brace")
	var thenNodes Nodes
	if thenNodes, matched, err = np.Parse(pi); err != nil || !matched {
		// Populate the nodes anyway, so that the LSP can use them.
		r.Then = thenNodes.Nodes
		return r, true, parse.Error("if: expected nodes, but none were found", pi.Position())
	}
	r.Then = thenNodes.Nodes

	// Read the optional 'ElseIf' Nodes.
	if r.ElseIfs, _, err = parse.ZeroOrMore(elseIfExpression).Parse(pi); err != nil {
		return r, true, err
	}

	// Read the optional 'Else' Nodes.
	var elseNodes Nodes
	if elseNodes, _, err = elseExpression.Parse(pi); err != nil {
		// Populate the nodes anyway, so that the LSP can use them.
		r.Else = elseNodes.Nodes
		return r, true, err
	}
	r.Else = elseNodes.Nodes

	// Read the required closing brace.
	if _, matched, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !matched {
		return r, true, parse.Error("if: expected closing brace", pi.Position())
	}

	return r, true, nil
}

var elseIfExpression parse.Parser[ElseIfExpression] = elseIfExpressionParser{}

type elseIfExpressionParser struct{}

func (elseIfExpressionParser) Parse(pi *parse.Input) (r ElseIfExpression, matched bool, err error) {
	start := pi.Index()

	// Check the prefix first.
	if _, matched, err = parse.All(parse.OptionalWhitespace, closeBrace, parse.OptionalWhitespace, parse.String("else if")).Parse(pi); err != nil || !matched {
		pi.Seek(start)
		return
	}

	// Rewind to the start of the `if` statement.
	pi.Seek(pi.Index() - 2)
	// Parse the Go if expression.
	if r.Expression, err = parseGo("else if", pi, goexpression.If); err != nil {
		return r, false, err
	}

	// Eat " {\n".
	if _, matched, err = parse.All(openBraceWithOptionalPadding, parse.NewLine).Parse(pi); err != nil || !matched {
		err = parse.Error("else if: "+unterminatedMissingCurly, pi.PositionAt(start))
		return
	}

	// Once we've had the start of an if block, we must conclude the block.

	// Read the 'Then' nodes.
	// If there's no match, there's a problem in the template nodes.
	np := newTemplateNodeParser(untilElseIfElseOrEnd, "else expression or closing brace")
	var thenNodes Nodes
	if thenNodes, matched, err = np.Parse(pi); err != nil || !matched {
		err = parse.Error("if: expected nodes, but none were found", pi.Position())
		return
	}
	r.Then = thenNodes.Nodes

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

func (elseExpressionParser) Parse(in *parse.Input) (r Nodes, matched bool, err error) {
	start := in.Index()

	// } else {
	if _, matched, err = endElseParser.Parse(in); err != nil || !matched {
		in.Seek(start)
		return
	}

	// Else contents
	if r, matched, err = newTemplateNodeParser(closeBraceWithOptionalPadding, "else expression closing brace").Parse(in); err != nil || !matched {
		in.Seek(start)
		return
	}

	return r, true, nil
}
