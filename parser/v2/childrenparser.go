package parser

import (
	"github.com/a-h/parse"
)

var childrenExpressionParser = parse.StringFrom(
	openBraceWithOptionalPadding,
	parse.OptionalWhitespace,
	parse.String("children..."),
	parse.OptionalWhitespace,
	closeBraceWithOptionalPadding,
)

var childrenExpression = parse.Func(func(in *parse.Input) (n Node, ok bool, err error) {
	start := in.Position()
	_, ok, err = childrenExpressionParser.Parse(in)
	if err != nil || !ok {
		return
	}
	r := &ChildrenExpression{
		Range: NewRange(start, in.Position()),
	}
	return r, true, nil
})
