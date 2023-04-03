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
		if _, ok, err = Must(closeBraceWithOptionalPadding, "css property expression: missing closing brace").Parse(pi); err != nil || !ok {
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
	if name, ok, err = Must(cssExpressionNameParser, "css expression: invalid name").Parse(pi); err != nil || !ok {
		return
	}
	r.Name = NewExpression(name, from, pi.Position())

	// Eat the open bracket.
	if _, ok, err = Must(parse.Rune('('), "css expression: parameters missing open bracket").Parse(pi); err != nil || !ok {
		return
	}

	// Check there's no parameters.
	from = pi.Position()
	if _, ok, err = parse.StringUntil(parse.Rune(')')).Parse(pi); err != nil {
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
	if _, ok, err = Must(expressionFuncEnd, "css expression: unterminated (missing ') {')").Parse(pi); err != nil || !ok {
		return
	}

	// Expect a newline.
	if _, ok, err = Must(parse.NewLine, "css expression: missing terminating newline").Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
})

// CSS property name parser.
var cssPropertyNameFirst = "abcdefghijklmnopqrstuvwxyz"
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
	if r.Value, ok, err = stringExpression.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// ;
	if _, ok, err = Must(parse.String(";"), "missing expected semicolon (;)").Parse(pi); err != nil || !ok {
		return
	}
	// \n
	if _, ok, err = Must(parse.NewLine, "missing expected linebreak").Parse(pi); err != nil || !ok {
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
	if r.Value, ok, err = Must(parse.StringUntil(untilEnd), "missing expected semicolon and linebreak (;\\n").Parse(pi); err != nil || !ok {
		return
	}

	// Chomp the ;\n
	if _, ok, err = Must(untilEnd, "failed to chomp semicolon and linebreak (;\\n)").Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
})
