package parser

import "github.com/a-h/parse"

func addTrailingSpace[TNode TrailingSpaceSetter](e TNode, pi *parse.Input, allowMulti bool) (n TNode, ok bool, err error) {
	// Add trailing space.
	ws, _, err := parse.Whitespace.Parse(pi)
	if err != nil {
		return e, false, err
	}
	trailingSpace, err := NewTrailingSpace(ws, allowMulti)
	if err != nil {
		return e, false, err
	}

	e.SetTrailingSpace(trailingSpace)

	return e, true, nil
}

// Eat any whitespace.
var whitespaceExpression = parse.Func(func(pi *parse.Input) (n Node, ok bool, err error) {
	var r Whitespace
	if r.Value, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}
	return r, len(r.Value) > 0, nil
})
