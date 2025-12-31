package parser

import (
	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

var switchExpression parse.Parser[Node] = switchExpressionParser{}

type switchExpressionParser struct{}

func (switchExpressionParser) Parse(pi *parse.Input) (n Node, matched bool, err error) {
	start := pi.Index()

	// Check the prefix first.
	if !peekPrefix(pi, "switch ") {
		pi.Seek(start)
		return n, false, nil
	}

	// Parse the Go switch expression.
	r := &SwitchExpression{}
	if r.Expression, err = parseGo("switch", pi, goexpression.Switch); err != nil {
		return r, true, err
	}

	// Eat " {\n".
	if _, matched, err = parse.All(openBraceWithOptionalPadding, parse.NewLine).Parse(pi); err != nil || !matched {
		err = parse.Error("switch: "+unterminatedMissingCurly, pi.PositionAt(start))
		return r, true, err
	}

	// Once we've had the start of a switch block, we must conclude the block.

	// Read the optional 'case' nodes.
	for {
		var ce CaseExpression
		ce, matched, err = caseExpressionParser.Parse(pi)
		if err != nil {
			// Capture the case for the LSP.
			r.Cases = append(r.Cases, ce)
			return r, true, err
		}
		if !matched {
			break
		}
		r.Cases = append(r.Cases, ce)
	}

	if err = assertLastCaseDoesNotContainFallthrough(r.Cases); err != nil {
		return r, true, err
	}

	// Optional whitespace.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		return r, false, err
	}

	// Read the required closing brace.
	if _, matched, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !matched {
		err = parse.Error("switch: "+unterminatedMissingEnd, pi.Position())
		return r, true, err
	}

	r.Range = NewRange(pi.PositionAt(start), pi.Position())
	return r, true, nil
}

const fallthroughInLastCaseError = "switch: fallthrough cannot be used in the last case of a switch statement"

func assertLastCaseDoesNotContainFallthrough(cases []CaseExpression) error {
	if len(cases) == 0 {
		return nil
	}
	lastCase := cases[len(cases)-1]
	if len(lastCase.Children) == 0 {
		return nil
	}
	lastChild := lastCase.Children[len(lastCase.Children)-1]
	ff, isFallthrough := lastChild.(*Fallthrough)
	if !isFallthrough {
		return nil
	}
	pos := parse.Position{
		Index: int(ff.Range.From.Index),
		Line:  int(ff.Range.From.Line),
		Col:   int(ff.Range.From.Col),
	}
	return parse.Error(fallthroughInLastCaseError, pos)
}

var caseExpressionStartParser = parse.Func(func(pi *parse.Input) (r Expression, matched bool, err error) {
	start := pi.Index()

	// Optional whitespace.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		return r, false, err
	}

	// Strip leading whitespace and look for `case ` or `default`.
	if !peekPrefix(pi, "case ", "default") {
		pi.Seek(start)
		return r, false, nil
	}
	// Parse the Go expression.
	if r, err = parseGo("case", pi, goexpression.Case); err != nil {
		return r, true, err
	}

	// Eat terminating newline.
	_, _, _ = parse.ZeroOrMore(parse.String(" ")).Parse(pi)
	_, _, _ = parse.NewLine.Parse(pi)

	return r, true, nil
})

var untilNextCaseOrEnd = parse.Any(StripType(caseExpressionStartParser), StripType(closeBraceWithOptionalPadding))

var caseExpressionParser = parse.Func(func(pi *parse.Input) (r CaseExpression, matched bool, err error) {
	if r.Expression, matched, err = caseExpressionStartParser.Parse(pi); err != nil || !matched {
		return r, matched, err
	}

	// Read until the next case statement, default, or end of the block.
	pr := newTemplateNodeParser(untilNextCaseOrEnd, "closing brace or case expression")
	var nodes Nodes
	if nodes, matched, err = pr.Parse(pi); err != nil || !matched {
		// Populate the nodes anyway, so that the LSP can use them.
		r.Children = nodes.Nodes
		if err != nil {
			return r, true, err
		}
		err = parse.Error("case: expected nodes, but none were found", pi.Position())
		return r, true, err
	}
	r.Children = nodes.Nodes

	// If we have children, validate that no statement in the middle is a fallthrough.
	if err = assertFallthroughMustBeLastInCase(r.Children); err != nil {
		return r, true, err
	}

	// Optional whitespace.
	if _, matched, err = parse.OptionalWhitespace.Parse(pi); err != nil || !matched {
		return r, true, err
	}

	return r, true, nil
})

const fallthroughMustBeLastInCaseError = "case: fallthrough can only be used as the last statement in a case block"

func assertFallthroughMustBeLastInCase(nodes []Node) error {
	if len(nodes) == 0 {
		return nil
	}
	for i, n := range nodes {
		if i == len(nodes)-1 {
			break
		}
		if ff, isFallthrough := n.(*Fallthrough); isFallthrough {
			pos := parse.Position{
				Index: int(ff.Range.From.Index),
				Line:  int(ff.Range.From.Line),
				Col:   int(ff.Range.From.Col),
			}
			return parse.Error(fallthroughMustBeLastInCaseError, pos)
		}
	}
	return nil
}
