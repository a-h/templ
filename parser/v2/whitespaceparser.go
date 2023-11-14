package parser

import "github.com/a-h/parse"

// Eat any whitespace.
var whitespaceExpression = parse.Func(func(pi *parse.Input) (n Node, ok bool, err error) {
	var r Whitespace
	if r.Value, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}
	return r, len(r.Value) > 0, nil
})
