package parser

import (
	"github.com/a-h/parse"
)

type anonymousTemplateParser struct{}

func (p anonymousTemplateParser) Parse(pi *parse.Input) (n Node, matched bool, err error) {
	start := pi.Position()

	// Check for "templ(" prefix (no space between templ and paren)
	if !peekPrefix(pi, "templ(") {
		return nil, false, nil
	}

	// Parse the function parameters
	// e.g., "templ(x string)" -> "(x string)"
	var expr Expression
	if expr, err = parseAnonymousTemplFuncParams(pi); err != nil {
		return nil, true, err
	}

	// Eat " {\n" or "{\n"
	if _, matched, err = parse.All(openBraceWithOptionalPadding, parse.StringFrom(parse.Optional(parse.NewLine))).Parse(pi); err != nil || !matched {
		return nil, true, parse.Error("anonymous templ: expected `templ(...) {`", pi.PositionAt(start.Index))
	}

	r := &AnonymousTemplate{
		Expression: expr,
	}

	// Parse children
	np := newTemplateNodeParser(closeBraceWithOptionalPadding, "anonymous template closing brace")
	var nodes Nodes
	if nodes, matched, err = np.Parse(pi); err != nil || !matched {
		r.Children = nodes.Nodes
		return r, true, err
	}
	r.Children = nodes.Nodes

	// Match closing brace
	if _, matched, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !matched {
		return r, true, parse.Error("anonymous templ: missing closing '}'", pi.Position())
	}

	r.Range = NewRange(start, pi.Position())
	return r, true, nil
}

var anonymousTemplate anonymousTemplateParser
