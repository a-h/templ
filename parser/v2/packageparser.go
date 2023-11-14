package parser

import (
	"github.com/a-h/parse"
)

// Package.
var pkg = parse.Func(func(pi *parse.Input) (pkg Package, ok bool, err error) {
	start := pi.Position()

	// Package prefix.
	if _, ok, err = parse.String("package ").Parse(pi); err != nil || !ok {
		return
	}

	// Once we have the prefix, it's an expression until the end of the line.
	var exp string
	if exp, ok, err = Must(stringUntilNewLine, "package literal not terminated").Parse(pi); err != nil || !ok {
		return
	}
	if len(exp) == 0 {
		ok = false
		err = parse.Error("package literal not terminated", start)
		return
	}

	// Success!
	pkg.Expression = NewExpression("package "+exp, start, pi.Position())

	return pkg, true, nil
})
