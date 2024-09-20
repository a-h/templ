package parser

import "github.com/a-h/parse"

func parseTrailingSpace(pi *parse.Input, allowMulti bool, forceVertical bool) (space TrailingSpace, err error) {
	// Add trailing space.
	ws, _, err := parse.Whitespace.Parse(pi)
	if err != nil {
		return "", err
	}

	trailing, err := NewTrailingSpace(ws, allowMulti)
	if err != nil {
		return "", err
	}

	// If the trailing space is not vertical, set it to vertical.
	if forceVertical && trailing != SpaceVertical && trailing != SpaceVerticalDouble {
		return SpaceVertical, nil
	}

	return trailing, nil
}

// Eat any whitespace.
var whitespaceExpression = parse.Func(func(pi *parse.Input) (n Node, ok bool, err error) {
	var r Whitespace
	if r.Value, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}
	return r, len(r.Value) > 0, nil
})
