package parser

import (
	"github.com/a-h/parse"
)

// ) {
var expressionFuncEnd = parse.All(parse.Rune(')'), openBraceWithOptionalPadding)

// Template

var template = parse.Func(func(pi *parse.Input) (r *HTMLTemplate, matched bool, err error) {
	start := pi.Position()

	// templ FuncName(p Person, other Other) {
	var te templateExpression
	if te, matched, err = templateExpressionParser.Parse(pi); err != nil || !matched {
		return r, matched, err
	}
	r = &HTMLTemplate{
		Expression: te.Expression,
	}
	defer func() {
		r.Range = NewRange(start, pi.Position())
	}()

	// Once we're in a template, we should expect some template whitespace, if/switch/for,
	// or node string expressions etc.
	var nodes Nodes
	nodes, matched, err = newTemplateNodeParser(closeBraceWithOptionalPadding, "template closing brace").Parse(pi)
	if err != nil {
		// The LSP wants as many nodes as possible, so even though there was an error,
		// we probably have some valid nodes that the LSP can use.
		r.Children = nodes.Nodes
		return r, true, err
	}
	if !matched {
		return r, true, parse.Error("templ: expected nodes in templ body, but found none", pi.Position())
	}
	r.Children = nodes.Nodes

	// Eat any whitespace.
	_, _, err = parse.OptionalWhitespace.Parse(pi)
	if err != nil {
		return
	}

	// Try for }
	if _, matched, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !matched {
		err = parse.Error("template: missing closing brace", pi.Position())
		return
	}

	return r, true, nil
})
