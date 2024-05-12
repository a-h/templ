package parser

import (
	"github.com/a-h/parse"
	"github.com/a-h/templ/cfg"
	"github.com/a-h/templ/parser/v2/goexpression"
)

var goCode = parse.Func(func(pi *parse.Input) (n Node, ok bool, err error) {
	if !cfg.Experiment.RawGo {
		return
	}
	// Check the prefix first.
	if _, ok, err = parse.Or(parse.String("{{ "), parse.String("{{")).Parse(pi); err != nil || !ok {
		return
	}

	// Once we have a prefix, we must have an expression that returns a string, with optional err.
	l := pi.Position().Line
	var r GoCode
	if r.Expression, err = parseGo("go code", pi, goexpression.Expression); err != nil {
		return r, false, err
	}

	if l != pi.Position().Line {
		r.Multiline = true
	}

	// Clear any optional whitespace.
	_, _, _ = parse.OptionalWhitespace.Parse(pi)

	// }}
	if _, ok, err = dblCloseBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("go code: missing close braces", pi.Position())
		return
	}

	// Parse trailing whitespace.
	ws, _, err := parse.Whitespace.Parse(pi)
	if err != nil {
		return r, false, err
	}
	r.TrailingSpace, err = NewTrailingSpace(ws)
	if err != nil {
		return r, false, err
	}

	return r, true, nil
})
