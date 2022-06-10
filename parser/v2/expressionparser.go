package parser

import (
	"errors"
	"io"
	"strings"

	"github.com/a-h/lexical/parse"
)

func parseUntil(combiner parse.MultipleResultCombiner, p parse.Function, delimiter parse.Function) parse.Function {
	return func(pi parse.Input) parse.Result {
		name := "function until delimiter"

		results := make([]interface{}, 0)
		for {
			current := pi.Index()
			ds := delimiter(pi)
			if ds.Success {
				rewind(pi, current)
				break
			}
			pr := p(pi)
			if pr.Error != nil {
				return parse.Failure(name, pr.Error)
			}
			if !pr.Success {
				return parse.Failure(name+": failed to match function", nil)
			}
			results = append(results, pr.Item)
		}
		item, ok := combiner(results)
		if !ok {
			return parse.Failure("until", errors.New("failed to combine results"))
		}
		return parse.Success("until", item, nil)
	}
}

var openBrace = parse.String("{")
var openBraceWithPadding = parse.String(" {")
var openBraceWithOptionalPadding = parse.Or(openBraceWithPadding, openBrace)

var closeBrace = parse.String("}")
var closeBraceWithPadding = parse.String(" }")
var closeBraceWithOptionalPadding = parse.Or(closeBraceWithPadding, closeBrace)

func chompBrace(pi parse.Input) (pr parse.Result, ok bool) {
	from := NewPositionFromInput(pi)
	pr = closeBraceWithOptionalPadding(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return
	}
	if !pr.Success {
		pr = parse.Failure("end brace missing", newParseError("expected closing brace not found", from, NewPositionFromInput(pi)))
		return
	}
	ok = true
	return
}

var exp = expressionParser{
	startBraceCount: 1,
}

type expressionParser struct {
	startBraceCount int
}

func (p expressionParser) Parse(pi parse.Input) parse.Result {
	braceCount := p.startBraceCount

	var sb strings.Builder
	var r rune
	var err error
loop:
	for {
		// Try to read a string literal first.
		result := string_lit(pi)
		if result.Error != nil && result.Error != io.EOF {
			return result
		}
		if result.Success {
			sb.WriteString(string(result.Item.(string)))
			continue
		}
		// Also try for a rune literal.
		result = rune_lit(pi)
		if result.Error != nil && result.Error != io.EOF {
			return result
		}
		if result.Success {
			sb.WriteString(string(result.Item.(string)))
			continue
		}
		// Try opener.
		result = openBrace(pi)
		if result.Error != nil && result.Error != io.EOF {
			return result
		}
		if result.Success {
			braceCount++
			sb.WriteString(result.Item.(string))
			continue
		}
		// Try closer.
		startOfCloseBrace := NewPositionFromInput(pi)
		result = closeBraceWithOptionalPadding(pi)
		if result.Error != nil && result.Error != io.EOF {
			return result
		}
		if result.Success {
			braceCount--
			if braceCount < 0 {
				return parse.Failure("expression: too many closing braces", nil)
			}
			if braceCount == 0 {
				rewind(pi, startOfCloseBrace.Index)
				break loop
			}
			switch r := result.Item.(type) {
			case string:
				sb.WriteString(r)
			case rune:
				sb.WriteRune(r)
			default:
				return parse.Failure("expression: internal error, unexpected result of brace", nil)
			}
			continue
		}

		// Read anything else.
		r, err = pi.Advance()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break loop
			}
			return parse.Failure("expression: failed to read", err)
		}
		if r == 65533 { // Invalid Unicode.
			break loop
		}
		sb.WriteRune(r)
	}
	if braceCount != 0 {
		return parse.Failure("expression: unexpected brace count", nil)
	}
	return parse.Success("expression", sb.String(), nil)
}

// Letters and digits

var octal_digit = parse.RuneIn("01234567")
var hex_digit = parse.RuneIn("0123456789ABCDEFabcdef")

// https://go.dev/ref/spec#Rune_literals

var rune_lit = parse.All(parse.WithStringConcatCombiner,
	parse.Rune('\''),
	parseUntil(parse.WithStringConcatCombiner,
		parse.Or(unicode_value_rune, byte_value),
		parse.Rune('\''),
	),
	parse.Rune('\''),
)
var unicode_value_rune = parse.Any(little_u_value, big_u_value, escaped_char, parse.RuneNotIn("'"))

//byte_value       = octal_byte_value | hex_byte_value .
var byte_value = parse.Any(octal_byte_value, hex_byte_value)

//octal_byte_value = `\` octal_digit octal_digit octal_digit .
var octal_byte_value = parse.All(parse.WithStringConcatCombiner,
	parse.String(`\`),
	octal_digit, octal_digit, octal_digit,
)

//hex_byte_value   = `\` "x" hex_digit hex_digit .
var hex_byte_value = parse.All(parse.WithStringConcatCombiner,
	parse.String(`\x`),
	hex_digit, hex_digit,
)

//little_u_value   = `\` "u" hex_digit hex_digit hex_digit hex_digit .
var little_u_value = parse.All(parse.WithStringConcatCombiner,
	parse.String(`\u`),
	hex_digit, hex_digit,
	hex_digit, hex_digit,
)

//big_u_value      = `\` "U" hex_digit hex_digit hex_digit hex_digit
var big_u_value = parse.All(parse.WithStringConcatCombiner,
	parse.String(`\U`),
	hex_digit, hex_digit, hex_digit, hex_digit,
	hex_digit, hex_digit, hex_digit, hex_digit,
)

//escaped_char     = `\` ( "a" | "b" | "f" | "n" | "r" | "t" | "v" | `\` | "'" | `"` ) .
var escaped_char = parse.All(parse.WithStringConcatCombiner,
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

var string_lit = parse.Or(interpreted_string_lit, raw_string_lit)

var interpreted_string_lit = parse.All(parse.WithStringConcatCombiner,
	parse.Rune('"'),
	parseUntil(parse.WithStringConcatCombiner,
		parse.Or(unicode_value_interpreted, byte_value),
		parse.Rune('"'),
	),
	parse.Rune('"'),
)
var unicode_value_interpreted = parse.Any(little_u_value, big_u_value, escaped_char, parse.RuneNotIn("\n\""))

var raw_string_lit = parse.All(parse.WithStringConcatCombiner,
	parse.Rune('`'),
	parseUntil(parse.WithStringConcatCombiner,
		unicode_value_raw,
		parse.Rune('`'),
	),
	parse.Rune('`'),
)
var unicode_value_raw = parse.Any(little_u_value, big_u_value, escaped_char, parse.RuneNotIn("`"))
