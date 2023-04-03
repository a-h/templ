package parser

import (
	"github.com/a-h/parse"
)

var childrenExpression = parse.Func(func(in *parse.Input) (out ChildrenExpression, ok bool, err error) {
	_, ok, err = parse.StringFrom(
		openBraceWithOptionalPadding,
		parse.OptionalWhitespace,
		parse.String("children..."),
		parse.OptionalWhitespace,
		closeBraceWithOptionalPadding,
	).Parse(in)
	return out, ok, err
})
