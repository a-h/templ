package parser

import "github.com/a-h/parse"

// Eat any whitespace.
var whitespaceExpression = parse.Func(func(pi *parse.Input) (n Node, ok bool, err error) {
	r := &Whitespace{}
	start := pi.Position()
	if r.Value, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}
	r.Range = NewRange(start, pi.Position())
	return r, len(r.Value) > 0, nil
})
