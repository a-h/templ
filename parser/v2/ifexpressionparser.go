package parser

import (
	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

var ifExpression ifExpressionParser

var untilElseIfElseOrEnd = parse.Any(StripType(elseIfExpression), StripType(elseKeyword), StripType(closeBraceWithOptionalPadding))

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
	if _, matched, err = openBraceAndNewLine.Parse(pi); err != nil || !matched {
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
		if err != nil {
			return r, true, err
		}
		return r, true, parse.Error("if: expected nodes, but none were found", pi.Position())
	}
	r.Then = thenNodes.Nodes

	// Read the optional 'ElseIf' Nodes.
	if r.ElseIfs, _, err = elseIfExpressions.Parse(pi); err != nil {
		return r, true, err
	}

	// Read the optional 'Else' Nodes.
	elseStart := pi.Position()
	var elseMatched bool
	if _, elseMatched, err = elseKeyword.Parse(pi); err != nil {
		return r, true, err
	}
	if elseMatched {
		r.ElseRange = NewRange(elseStart, pi.Position())
		if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
			return r, true, err
		}
		np := newTemplateNodeParser(closeBraceWithOptionalPadding, "else expression closing brace")
		var elseNodes Nodes
		elseNodes, _, err = np.Parse(pi)
		// Populate the nodes even on error, so that the LSP can use them.
		r.Else = elseNodes.Nodes
		if err != nil {
			return r, true, err
		}
	}

	// Read the required closing brace.
	if _, matched, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !matched {
		return r, true, parse.Error("if: expected closing brace", pi.Position())
	}

	r.Range = NewRange(pi.PositionAt(start), pi.Position())
	return r, true, nil
}

func (p ifExpressionParser) Match(in *parse.Input) (ok bool, err error) {
	return parse.Match(p, in)
}

var elseIfExpression parse.Parser[ElseIfExpression] = elseIfExpressionParser{}
var elseIfExpressions = parse.ZeroOrMore(elseIfExpression)
var elseIfPrefix = parse.All(parse.OptionalWhitespace, closeBrace, parse.OptionalWhitespace, parse.String("else if"))

type elseIfExpressionParser struct{}

func (elseIfExpressionParser) Parse(pi *parse.Input) (r ElseIfExpression, matched bool, err error) {
	start := pi.Index()

	// Check the prefix first.
	if _, matched, err = elseIfPrefix.Parse(pi); err != nil || !matched {
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
	if _, matched, err = openBraceAndNewLine.Parse(pi); err != nil || !matched {
		err = parse.Error("else if: "+unterminatedMissingCurly, pi.PositionAt(start))
		return
	}

	// Once we've had the start of an if block, we must conclude the block.

	// Read the 'Then' nodes.
	// If there's no match, there's a problem in the template nodes.
	np := newTemplateNodeParser(untilElseIfElseOrEnd, "else expression or closing brace")
	var thenNodes Nodes
	if thenNodes, matched, err = np.Parse(pi); err != nil || !matched {
		if err != nil {
			return r, true, err
		}
		err = parse.Error("if: expected nodes, but none were found", pi.Position())
		return
	}
	r.Then = thenNodes.Nodes

	r.Range = NewRange(pi.PositionAt(start), pi.Position())
	return r, true, nil
}

func (p elseIfExpressionParser) Match(in *parse.Input) (ok bool, err error) {
	return parse.Match(p, in)
}

var elseKeyword = parse.All(
	parse.Rune('}'),
	parse.OptionalWhitespace,
	parse.String("else"),
	parse.OptionalWhitespace,
	parse.Rune('{'))
