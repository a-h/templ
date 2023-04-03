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

	// Once we're in a template, we should expect some template whitespace, if/switch/for,
	// or node string expressions etc.
	r.Children, ok, err = Must[[]Node](newTemplateNodeParser(closeBraceWithOptionalPadding, "template closing brace"), "templ: expected nodes in templ body, but found none").Parse(pi)
	if err != nil || !ok {
		return
	}

	// Eat any whitespace.
	_, _, err = parse.OptionalWhitespace.Parse(pi)
	if err != nil {
		return
	}

	// Try for }
	_, _, err = Must(closeBraceWithOptionalPadding, "template: missing closing brace").Parse(pi)
	if err != nil {
		return
	}

	return r, true, nil
})
