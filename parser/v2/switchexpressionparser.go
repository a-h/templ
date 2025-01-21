package parser

import (
	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

var switchExpression parse.Parser[Node] = switchExpressionParser{}

type switchExpressionParser struct{}

func (switchExpressionParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	var r SwitchExpression
	start := pi.Index()

	// Check the prefix first.
	if !peekPrefix(pi, "switch ") {
		pi.Seek(start)
		return
	}

	// Parse the Go switch expression.
	if r.Expression, err = parseGo("switch", pi, goexpression.Switch); err != nil {
		return r, false, err
	}

	// Eat " {\n".
	if _, ok, err = parse.All(openBraceWithOptionalPadding, parse.NewLine).Parse(pi); err != nil || !ok {
		err = parse.Error("switch: "+unterminatedMissingCurly, pi.PositionAt(start))
		return
	}

	// Once we've had the start of a switch block, we must conclude the block.

	// Read the optional 'case' nodes.
	for {
		var ce CaseExpression
		ce, ok, err = caseExpressionParser.Parse(pi)
		if err != nil {
			return
		}
		if !ok {
			break
		}
		r.Cases = append(r.Cases, ce)
	}

	// Read the required closing brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("switch: "+unterminatedMissingEnd, pi.Position())
		return
	}

	return r, true, nil
}

var caseExpressionStartParser = parse.Func(func(pi *parse.Input) (r Expression, ok bool, err error) {
	start := pi.Index()

	// Optional whitespace.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		return
	}

	// Strip leading whitespace and look for `case ` or `default`.
	if !peekPrefix(pi, "case ", "default") {
		pi.Seek(start)
		return r, false, nil
	}
	// Parse the Go expression.
	if r, err = parseGo("case", pi, goexpression.Case); err != nil {
		return r, false, err
	}

	// Eat terminating newline.
	_, _, _ = parse.ZeroOrMore(parse.String(" ")).Parse(pi)
	_, _, _ = parse.NewLine.Parse(pi)

	return r, true, nil
})

var caseExpressionParser = parse.Func(func(pi *parse.Input) (r CaseExpression, ok bool, err error) {
	if r.Expression, ok, err = caseExpressionStartParser.Parse(pi); err != nil || !ok {
		return
	}

	// Read until the next case statement, default, or end of the block.
	pr := newTemplateNodeParser(parse.Any(StripType(closeBraceWithOptionalPadding), StripType(caseExpressionStartParser)), "closing brace or case expression")
	var nodes Nodes
	if nodes, ok, err = pr.Parse(pi); err != nil || !ok {
		err = parse.Error("case: expected nodes, but none were found", pi.Position())
		return
	}
	r.Children = nodes.Nodes

	// Optional whitespace.
	if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
})
