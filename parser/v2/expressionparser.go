package parser

import (
	"strings"

	"github.com/a-h/parse"
)

// StripType takes the parser and throws away the return value.
func StripType[T any](p parse.Parser[T]) parse.Parser[any] {
	return parse.Func(func(in *parse.Input) (out any, ok bool, err error) {
		return p.Parse(in)
	})
}

func ExpressionOf(p parse.Parser[string]) parse.Parser[Expression] {
	return parse.Func(func(in *parse.Input) (out Expression, ok bool, err error) {
		from := in.Position()

		var exp string
		if exp, ok, err = p.Parse(in); err != nil || !ok {
			return
		}

		return NewExpression(exp, from, in.Position()), true, nil
	})
}

var lt = parse.Rune('<')
var gt = parse.Rune('>')
var openBrace = parse.String("{")
var optionalSpaces = parse.StringFrom(parse.Optional(
	parse.AtLeast(1, parse.Rune(' '))))
var openBraceWithPadding = parse.StringFrom(optionalSpaces,
	openBrace,
	optionalSpaces)
var openBraceWithOptionalPadding = parse.Any(openBraceWithPadding, openBrace)

var closeBrace = parse.String("}")
var closeBraceWithOptionalPadding = parse.StringFrom(optionalSpaces, closeBrace)

var dblOpenBrace = parse.String("{{")
var dblOpenBraceWithOptionalPadding = parse.StringFrom(dblOpenBrace, optionalSpaces)

var dblCloseBrace = parse.String("}}")
var dblCloseBraceWithOptionalPadding = parse.StringFrom(optionalSpaces, dblCloseBrace)

var openBracket = parse.String("(")
var closeBracket = parse.String(")")

var stringUntilNewLine = parse.StringUntil(parse.NewLine)
var newLineOrEOF = parse.Or(parse.NewLine, parse.EOF[string]())
var stringUntilNewLineOrEOF = parse.StringUntil(newLineOrEOF)

var jsOrGoSingleLineComment = parse.StringFrom(parse.String("//"), parse.StringUntil(parse.Any(parse.NewLine, parse.EOF[string]())))
var jsOrGoMultiLineComment = parse.StringFrom(parse.String("/*"), parse.StringUntil(parse.String("*/")))

var exp = expressionParser{
	startBraceCount: 1,
}

type expressionParser struct {
	startBraceCount int
}

func (p expressionParser) Parse(pi *parse.Input) (s Expression, ok bool, err error) {
	from := pi.Position()

	braceCount := p.startBraceCount

	sb := new(strings.Builder)
loop:
	for {
		var result string

		// Try to parse a single line comment.
		if result, ok, err = jsOrGoSingleLineComment.Parse(pi); err != nil {
			return
		}
		if ok {
			sb.WriteString(result)
			continue
		}

		// Try to parse a multi-line comment.
		if result, ok, err = jsOrGoMultiLineComment.Parse(pi); err != nil {
			return
		}
		if ok {
			sb.WriteString(result)
			continue
		}

		// Try to read a string literal.
		if result, ok, err = string_lit.Parse(pi); err != nil {
			return
		}
		if ok {
			sb.WriteString(result)
			continue
		}
		// Also try for a rune literal.
		if result, ok, err = rune_lit.Parse(pi); err != nil {
			return
		}
		if ok {
			sb.WriteString(result)
			continue
		}
		// Try opener.
		if result, ok, err = openBrace.Parse(pi); err != nil {
			return
		}
		if ok {
			braceCount++
			sb.WriteString(result)
			continue
		}
		// Try closer.
		startOfCloseBrace := pi.Index()
		if result, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil {
			return
		}
		if ok {
			braceCount--
			if braceCount < 0 {
				err = parse.Error("expression: too many closing braces", pi.Position())
				return
			}
			if braceCount == 0 {
				pi.Seek(startOfCloseBrace)
				break loop
			}
			sb.WriteString(result)
			continue
		}

		// Read anything else.
		var c string
		c, ok = pi.Take(1)
		if !ok {
			break loop
		}
		if rune(c[0]) == 65533 { // Invalid Unicode.
			break loop
		}
		sb.WriteString(c)
	}
	if braceCount != 0 {
		err = parse.Error("expression: unexpected brace count", pi.Position())
		return
	}

	return NewExpression(sb.String(), from, pi.Position()), true, nil
}

// Letters and digits

var octal_digit = parse.RuneIn("01234567")
var hex_digit = parse.RuneIn("0123456789ABCDEFabcdef")

// https://go.dev/ref/spec#Rune_literals

var rune_lit = parse.StringFrom(
	parse.Rune('\''),
	parse.StringFrom(parse.Until(
		parse.Any(unicode_value_rune, byte_value),
		parse.Rune('\''),
	)),
	parse.Rune('\''),
)
var unicode_value_rune = parse.Any(little_u_value, big_u_value, escaped_char, parse.RuneNotIn("'"))

// byte_value       = octal_byte_value | hex_byte_value .
var byte_value = parse.Any(octal_byte_value, hex_byte_value)

// octal_byte_value = `\` octal_digit octal_digit octal_digit .
var octal_byte_value = parse.StringFrom(
	parse.String(`\`),
	octal_digit, octal_digit, octal_digit,
)

// hex_byte_value   = `\` "x" hex_digit hex_digit .
var hex_byte_value = parse.StringFrom(
	parse.String(`\x`),
	hex_digit, hex_digit,
)

// little_u_value   = `\` "u" hex_digit hex_digit hex_digit hex_digit .
var little_u_value = parse.StringFrom(
	parse.String(`\u`),
	hex_digit, hex_digit,
	hex_digit, hex_digit,
)

// big_u_value      = `\` "U" hex_digit hex_digit hex_digit hex_digit
var big_u_value = parse.StringFrom(
	parse.String(`\U`),
	hex_digit, hex_digit, hex_digit, hex_digit,
	hex_digit, hex_digit, hex_digit, hex_digit,
)

// escaped_char     = `\` ( "a" | "b" | "f" | "n" | "r" | "t" | "v" | `\` | "'" | `"` ) .
var escaped_char = parse.StringFrom(
	parse.Rune('\\'),
	parse.Any(
		parse.Rune('a'),
		parse.Rune('b'),
		parse.Rune('f'),
		parse.Rune('n'),
		parse.Rune('r'),
		parse.Rune('t'),
		parse.Rune('v'),
		parse.Rune('\\'),
		parse.Rune('\''),
		parse.Rune('"'),
	),
)

// https://go.dev/ref/spec#String_literals

var string_lit = parse.Any(parse.String(`""`), parse.String(`''`), interpreted_string_lit, raw_string_lit)

var interpreted_string_lit = parse.StringFrom(
	parse.Rune('"'),
	parse.StringFrom(parse.Until(
		parse.Any(unicode_value_interpreted, byte_value),
		parse.Rune('"'),
	)),
	parse.Rune('"'),
)
var unicode_value_interpreted = parse.Any(little_u_value, big_u_value, escaped_char, parse.RuneNotIn("\n\""))

var raw_string_lit = parse.StringFrom(
	parse.Rune('`'),
	parse.StringFrom(parse.Until(
		unicode_value_raw,
		parse.Rune('`'),
	)),
	parse.Rune('`'),
)
var unicode_value_raw = parse.Any(little_u_value, big_u_value, escaped_char, parse.RuneNotIn("`"))
