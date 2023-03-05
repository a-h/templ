package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// CSS.

// CSS Parser.

func newCSSParser() cssParser {
	return cssParser{}
}

type cssParser struct {
}

var endCssParser = createEndParser("endcss") // {% endcss %}

func (p cssParser) Parse(pi parse.Input) parse.Result {
	r := CSSTemplate{
		Properties: []CSSProperty{},
	}

	// Parse the name.
	pr := newCSSExpressionParser().Parse(pi)
	if !pr.Success {
		return pr
	}
	r.Name = pr.Item.(cssExpression).Name

	var from Position
	for {
		var pr parse.Result

		// Try for an expression CSS declaration.
		// background-color: {%= constants.BackgroundColor %};
		pr = newExpressionCSSPropertyParser().Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			r.Properties = append(r.Properties, pr.Item.(CSSProperty))
			continue
		}

		// Try for a constant CSS declaration.
		// color: #ffffff;
		pr = newConstantCSSPropertyParser().Parse(pi)
		if pr.Error != nil {
			return pr
		}
		if pr.Success {
			r.Properties = append(r.Properties, pr.Item.(CSSProperty))
			continue
		}

		// Eat any whitespace.
		pr = optionalWhitespaceParser(pi)
		if pr.Error != nil {
			return pr
		}
		// {% endcss %}
		from = NewPositionFromInput(pi)
		if endCssParser(pi).Success {
			return parse.Success("css", r, nil)
		}
		return parse.Failure("css", newParseError("expected {% endcss %} not found", from, NewPositionFromInput(pi)))
	}
}

// {% css Func() %}
type cssExpression struct {
	Name Expression
}

func newCSSExpressionParser() cssExpressionParser {
	return cssExpressionParser{}
}

type cssExpressionParser struct {
}

var cssExpressionStartParser = createStartParser("css")

var cssExpressionNameParser = parse.All(parse.WithStringConcatCombiner,
	parse.Letter,
	parse.Many(parse.WithStringConcatCombiner, 0, 1000, parse.Any(parse.Letter, parse.ZeroToNine)),
)

func (p cssExpressionParser) Parse(pi parse.Input) parse.Result {
	var r cssExpression

	// Check the prefix first.
	prefixResult := cssExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have the prefix, we must have a name and parameters.
	// Read the name of the function.
	from := NewPositionFromInput(pi)
	pr := cssExpressionNameParser(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, the name wasn't correctly terminated.
	if !pr.Success {
		return parse.Failure("cssExpressionParser", newParseError("css expression: invalid name", from, NewPositionFromInput(pi)))
	}
	to := NewPositionFromInput(pi)
	r.Name = NewExpression(pr.Item.(string), from, to)
	from = to

	// Eat the open bracket.
	if lb := parse.Rune('(')(pi); !lb.Success {
		return parse.Failure("cssExpressionParser", newParseError("css expression: parameters missing open bracket", from, NewPositionFromInput(pi)))
	}

	// Check there's no parameters.
	from = NewPositionFromInput(pi)
	pr = parse.StringUntil(parse.Rune(')'))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, the name wasn't correctly terminated.
	if !pr.Success {
		return parse.Failure("cssExpressionParser", newParseError("css expression: parameters missing close bracket", from, NewPositionFromInput(pi)))
	}
	if len(pr.Item.(string)) > 1 {
		return parse.Failure("cssExpressionParser", newParseError("css expression: found unexpected parameters", from, NewPositionFromInput(pi)))
	}

	// Eat ") %}".
	from = NewPositionFromInput(pi)
	if lb := expressionFuncEnd(pi); !lb.Success {
		return parse.Failure("cssExpressionParser", newParseError("css expression: unterminated (missing ') %}')", from, NewPositionFromInput(pi)))
	}

	// Expect a newline.
	from = NewPositionFromInput(pi)
	if lb := newLine(pi); !lb.Success {
		return parse.Failure("cssExpressionParser", newParseError("css expression: missing terminating newline", from, NewPositionFromInput(pi)))
	}

	return parse.Success("cssExpressionParser", r, nil)
}

// CSS property name parser.
var cssPropertyNameFirst = "abcdefghijklmnopqrstuvwxyz"
var cssPropertyNameSubsequent = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
var cssPropertyNameParser = parse.Then(parse.WithStringConcatCombiner,
	parse.RuneIn(cssPropertyNameFirst),
	parse.Many(parse.WithStringConcatCombiner, 0, 128, parse.RuneIn(cssPropertyNameSubsequent)),
)

// background-color: {%= constants.BackgroundColor %};
func newExpressionCSSPropertyParser() expressionCSSPropertyParser {
	return expressionCSSPropertyParser{}
}

type expressionCSSPropertyParser struct {
}

func (p expressionCSSPropertyParser) Parse(pi parse.Input) parse.Result {
	var r ExpressionCSSProperty
	var pr parse.Result
	start := pi.Index()

	// Optional whitespace.
	if pr = optionalWhitespaceParser(pi); pr.Error != nil {
		return pr
	}
	// Property name.
	if pr = cssPropertyNameParser(pi); !pr.Success {
		err := rewind(pi, start)
		if err != nil {
			return parse.Failure("failed to rewind reader", err)
		}
		return pr
	}
	r.Name = pr.Item.(string)
	// <space>:<space>
	pr = parse.All(parse.WithStringConcatCombiner,
		optionalWhitespaceParser,
		parse.Rune(':'),
		optionalWhitespaceParser)(pi)
	if !pr.Success {
		err := rewind(pi, start)
		if err != nil {
			return parse.Failure("failed to rewind reader", err)
		}
		return pr
	}

	// {%= string %}
	pr = newStringExpressionParser().Parse(pi)
	if !pr.Success {
		err := rewind(pi, start)
		if err != nil {
			return parse.Failure("failed to rewind reader", err)
		}
		return pr
	}
	r.Value = pr.Item.(StringExpression)

	// ;
	from := NewPositionFromInput(pi)
	if pr = parse.String(";")(pi); !pr.Success {
		return parse.Failure("expression css declaration", newParseError("missing expected semicolon (;)", from, NewPositionFromInput(pi)))
	}
	// \n
	from = NewPositionFromInput(pi)
	if pr = parse.String("\n")(pi); !pr.Success {
		return parse.Failure("expression css declaration", newParseError("missing expected linebreak", from, NewPositionFromInput(pi)))
	}

	return parse.Success("expression css declaration", r, nil)
}

// background-color: #ffffff;
func newConstantCSSPropertyParser() constantCSSPropertyParser {
	return constantCSSPropertyParser{}
}

type constantCSSPropertyParser struct {
}

func (p constantCSSPropertyParser) Parse(pi parse.Input) parse.Result {
	var r ConstantCSSProperty
	var pr parse.Result
	start := pi.Index()

	// Optional whitespace.
	if pr = optionalWhitespaceParser(pi); pr.Error != nil {
		return pr
	}
	// Property name.
	if pr = cssPropertyNameParser(pi); !pr.Success {
		err := rewind(pi, start)
		if err != nil {
			return parse.Failure("failed to rewind reader", err)
		}
		return pr
	}
	r.Name = pr.Item.(string)
	// <space>:<space>
	pr = parse.All(parse.WithStringConcatCombiner,
		optionalWhitespaceParser,
		parse.Rune(':'),
		optionalWhitespaceParser)(pi)
	if !pr.Success {
		err := rewind(pi, start)
		if err != nil {
			return parse.Failure("failed to rewind reader", err)
		}
		return pr
	}

	// Everything until ';\n'
	from := NewPositionFromInput(pi)
	untilEnd := parse.All(parse.WithStringConcatCombiner,
		optionalWhitespaceParser,
		parse.String(";\n"),
	)
	pr = parse.StringUntil(untilEnd)(pi)
	if !pr.Success {
		return parse.Failure("constant css declaration", newParseError("missing expected semicolon and linebreak (;\\n)", from, NewPositionFromInput(pi)))
	}
	r.Value = pr.Item.(string)
	// Chomp the ;\n
	pr = untilEnd(pi)
	if !pr.Success {
		return parse.Failure("constant css declaration", newParseError("failed to chomp semicolon and linebreak (;\\n)", from, NewPositionFromInput(pi)))
	}

	return parse.Success("constant css declaration", r, nil)
}
