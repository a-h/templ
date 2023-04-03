package parser

import (
	"fmt"
	"strings"

	"github.com/a-h/parse"
)

var templBlockElementExpression = parse.Func(func(pi *parse.Input) (r TemplElementExpression, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.Rune('@').Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got a prefix, read until {\n.
	endOfStatementExpression := ExpressionOf(parse.StringUntil(parse.All(openBraceWithOptionalPadding, parse.NewLine)))
	if r.Expression, ok, err = endOfStatementExpression.Parse(pi); err != nil || !ok {
		return
	}

	// Eat " {\n".
	if _, ok, err = Must(parse.All(openBraceWithOptionalPadding, parse.NewLine), "templ element: unterminated (missing closing '{\n')").Parse(pi); err != nil || !ok {
		return
	}

	// Once we've had the start of a for block, we must conclude the block.

	// Node contents.
	np := newTemplateNodeParser(closeBraceWithOptionalPadding, "templ element closing brace")
	if r.Children, ok, err = Must[[]Node](np, fmt.Sprintf("@%s: expected nodes, but none were found", r.Expression.Value)).Parse(pi); err != nil || !ok {
		return
	}

	// Read the required closing brace.
	if _, ok, err = Must(closeBraceWithOptionalPadding, fmt.Sprintf("@%s: missing end (expected '}')", r.Expression.Value)).Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
})

var templSelfClosingElementExpression = parse.Func(func(pi *parse.Input) (e TemplElementExpression, ok bool, err error) {
	start := pi.Index()

	// Check the prefix first.
	if _, ok, err = parse.Rune('@').Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got a prefix, read until \n.
	endOfStatementExpression := ExpressionOf(parse.StringUntil(parse.NewLine))
	if e.Expression, ok, err = Must(endOfStatementExpression, "templ element: unterminated (missing closing newline)").Parse(pi); err != nil || !ok {
		return
	}

	// It isn't a self-closing expression if there's an opening brace.
	if strings.HasSuffix(strings.TrimSpace(e.Expression.Value), "{") {
		pi.Seek(start)
		return e, false, nil
	}

	return e, true, nil
})

var templElementExpression = parse.Any(templSelfClosingElementExpression, templBlockElementExpression)
