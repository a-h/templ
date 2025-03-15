package parser

import (
	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

// goCode is the parser used to parse Raw Go code within templates.
//
// goCodeInJavaScript is the same, but handles the case where Go expressions
// are embedded within JavaScript.
//
// The only difference is that goCode normalises whitespace after the
// closing brace pair, whereas goCodeInJavaScript retains all whitespace.
var goCode = getGoCodeParser(true)
var goCodeInJavaScript = getGoCodeParser(false)

func getGoCodeParser(normalizeWhitespace bool) parse.Parser[Node] {
	return parse.Func(func(pi *parse.Input) (n Node, ok bool, err error) {
		// Check the prefix first.
		if _, ok, err = dblOpenBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
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
		if normalizeWhitespace {
			if r.TrailingSpace, err = NewTrailingSpace(ws); err != nil {
				return r, false, err
			}
		} else {
			r.TrailingSpace = TrailingSpace(ws)
		}

		return r, true, nil
	})
}
