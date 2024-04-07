package parser

import (
	"github.com/a-h/parse"
)

// ) {
var expressionFuncEnd = parse.All(parse.Rune(')'), openBraceWithOptionalPadding)

// Template

var template = parse.Func(func(pi *parse.Input) (r HTMLTemplate, ok bool, err error) {
	// templ FuncName(p Person, other Other) {
	var te templateExpression
	if te, ok, err = templateExpressionParser.Parse(pi); err != nil || !ok {
		return
	}
	r.Expression = te.Expression

	childrenStartPos := pi.Position()

	// Once we're in a template, we should expect some template whitespace, if/switch/for,
	// or node string expressions etc.
	var nodes Nodes
	nodes, ok, err = newTemplateNodeParser(closeBraceWithOptionalPadding, "template closing brace").Parse(pi)
	if err != nil {
		return
	}
	if !ok {
		err = parse.Error("templ: expected nodes in templ body, but found none", pi.Position())
		return
	}
	r.Children = nodes.Nodes

	// Eat any whitespace.
	_, _, err = parse.OptionalWhitespace.Parse(pi)
	if err != nil {
		return
	}

	r.ChildrenRange = NewRange(childrenStartPos, pi.Position())

	// Try for }
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("template: missing closing brace", pi.Position())
		return
	}

	return r, true, nil
})
