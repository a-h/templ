package parser

import (
	"strings"
	"unicode"

	"github.com/a-h/parse"
)

// CSS.

// CSS Parser.
var cssParser = parse.Func(func(pi *parse.Input) (r CSSTemplate, ok bool, err error) {
	r = CSSTemplate{
		Properties: []CSSProperty{},
	}

	// Parse the name.
	var exp cssExpression
	if exp, ok, err = cssExpressionParser.Parse(pi); err != nil || !ok {
		return
	}
	r.Name = exp.Name

	for {
		var cssProperty CSSProperty

		// Try for an expression CSS declaration.
		// background-color: { constants.BackgroundColor };
		cssProperty, ok, err = expressionCSSPropertyParser.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			r.Properties = append(r.Properties, cssProperty)
			continue
		}

		// Try for a constant CSS declaration.
		// color: #ffffff;
		cssProperty, ok, err = constantCSSPropertyParser.Parse(pi)
		if err != nil {
			return
		}
		if ok {
			r.Properties = append(r.Properties, cssProperty)
			continue
		}

		// Eat any whitespace.
		if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
			return
		}

		// Try for }
		if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
			err = parse.Error("css property expression: missing closing brace", pi.Position())
			return
		}

		return r, true, nil
	}
})

// css Func() {
type cssExpression struct {
	Name Expression
}

var cssExpressionStartParser = parse.String("css ")

var cssExpressionNameParser = parse.Func(func(in *parse.Input) (name string, ok bool, err error) {
	var c string
	if c, ok = in.Peek(1); !ok || !unicode.IsLetter(rune(c[0])) {
		return
	}
	prefix, _, _ := parse.Letter.Parse(in)
	suffix, _, _ := parse.AtMost(1000, parse.Any(parse.Letter, parse.ZeroToNine)).Parse(in)
	return prefix + strings.Join(suffix, ""), true, nil
})

var cssExpressionParser = parse.Func(func(pi *parse.Input) (r cssExpression, ok bool, err error) {
	// Check the prefix first.
	if _, ok, err = cssExpressionStartParser.Parse(pi); err != nil || !ok {
		return
	}

	// Once we have the prefix, we must have a name and parameters.
	// Read the name of the function.
	from := pi.Position()
	// If there's no match, the name wasn't correctly terminated.
	var name string
	if name, ok, err = cssExpressionNameParser.Parse(pi); err != nil || !ok {
		err = parse.Error("css expression: invalid name", pi.Position())
		return
	}
	r.Name = NewExpression(name, from, pi.Position())

	// Eat the open bracket.
	if _, ok, err = openBracket.Parse(pi); err != nil || !ok {
		err = parse.Error("css expression: parameters missing open bracket", pi.Position())
		return
	}

	// Check there's no parameters.
	from = pi.Position()
	if _, ok, err = parse.StringUntil(closeBracket).Parse(pi); err != nil {
		return
	}
	// If there's no match, the name wasn't correctly terminated.
	if !ok {
		return r, ok, parse.Error("css expression: parameters missing close bracket", pi.Position())
	}
	if pi.Index()-int(from.Index) > 0 {
		return r, ok, parse.Error("css expression: found unexpected parameters", pi.Position())
	}

	// Eat ") {".
	if _, ok, err = expressionFuncEnd.Parse(pi); err != nil || !ok {
		err = parse.Error("css expression: unterminated (missing ') {')", pi.Position())
		return
	}

	// Expect a newline.
	if _, ok, err = parse.NewLine.Parse(pi); err != nil || !ok {
		err = parse.Error("css expression: missing terminating newline", pi.Position())
		return
	}

	return r, true, nil
})

// CSS property name parser.
var cssPropertyNameFirst = "abcdefghijklmnopqrstuvwxyz-"
var cssPropertyNameSubsequent = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
var cssPropertyNameParser = parse.Func(func(in *parse.Input) (name string, ok bool, err error) {
	start := in.Position()
	var prefix, suffix string
	if prefix, ok, err = parse.RuneIn(cssPropertyNameFirst).Parse(in); err != nil || !ok {
		return
	}
	if suffix, ok, err = parse.StringUntil(parse.RuneNotIn(cssPropertyNameSubsequent)).Parse(in); err != nil || !ok {
		in.Seek(start.Index)
		return
	}
	if len(suffix)+1 > 128 {
		ok = false
		err = parse.Error("css property names must be < 128 characters long", in.Position())
		return
	}
	return prefix + suffix, true, nil
})

// background-color: {%= constants.BackgroundColor %};
var expressionCSSPropertyParser = parse.Func(func(pi *parse.Input) (r ExpressionCSSProperty, ok bool, err error) {
	start := pi.Index()

	// Optional whitespace.
	if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}
	// Property name.
	if r.Name, ok, err = cssPropertyNameParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}
	// <space>:<space>
	if _, ok, err = parse.All(parse.OptionalWhitespace, parse.Rune(':'), parse.OptionalWhitespace).Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// { string }
	var se Node
	if se, ok, err = stringExpression.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}
	r.Value = se.(StringExpression)

	// ;
	if _, ok, err = parse.String(";").Parse(pi); err != nil || !ok {
		err = parse.Error("missing expected semicolon (;)", pi.Position())
		return
	}
	// \n
	if _, ok, err = parse.NewLine.Parse(pi); err != nil || !ok {
		err = parse.Error("missing expected linebreak", pi.Position())
		return
	}

	return r, true, nil
})

// background-color: #ffffff;
var constantCSSPropertyParser = parse.Func(func(pi *parse.Input) (r ConstantCSSProperty, ok bool, err error) {
	start := pi.Index()

	// Optional whitespace.
	if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}
	// Property name.
	if r.Name, ok, err = cssPropertyNameParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}
	// <space>:<space>
	if _, ok, err = parse.All(parse.OptionalWhitespace, parse.Rune(':'), parse.OptionalWhitespace).Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Everything until ';\n'
	untilEnd := parse.All(
		parse.OptionalWhitespace,
		parse.Rune(';'),
		parse.NewLine,
	)
	if r.Value, ok, err = parse.StringUntil(untilEnd).Parse(pi); err != nil || !ok {
		err = parse.Error("missing expected semicolon and linebreak (;\\n", pi.Position())
		return
	}

	// Chomp the ;\n
	if _, ok, err = untilEnd.Parse(pi); err != nil || !ok {
		err = parse.Error("failed to chomp semicolon and linebreak (;\\n)", pi.Position())
		return
	}

	return r, true, nil
})
