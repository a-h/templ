package parser

import (
	"github.com/a-h/parse"
)

const (
	unterminatedMissingCurly = `switch: unterminated (missing closing '{\n') - INSERT LINK`
	unterminatedMissingEnd   = `switch: missing end (expected '}') - INSERT LINK`
)

var forExpression = parse.Func(func(pi *parse.Input) (r ForExpression, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = parse.String("for ").Parse(pi); err != nil || !ok {
		return
	}

	// Once we've got a prefix, read until {\n.
	// If there's no match, there's no {\n, which is an error.
	from := pi.Position()
	until := parse.All(openBraceWithOptionalPadding, parse.NewLine)
	var fexp string
	if fexp, ok, err = Must(parse.StringUntil(until), "for: "+unterminatedMissingCurly).Parse(pi); err != nil || !ok {
		return
	}
	r.Expression = NewExpression(fexp, from, pi.Position())

	// Eat " {".
	if _, ok, err = Must(until, "for: "+unterminatedMissingCurly).Parse(pi); err != nil || !ok {
		return
	}

	// Node contents.
	tnp := newTemplateNodeParser(closeBraceWithOptionalPadding, "for expression closing brace")
	if r.Children, ok, err = Must[[]Node](tnp, "for: expected nodes, but none were found").Parse(pi); err != nil || !ok {
		return
	}

	// Read the required closing brace.
	if _, ok, err = Must(closeBraceWithOptionalPadding, "for: "+unterminatedMissingEnd).Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
})
