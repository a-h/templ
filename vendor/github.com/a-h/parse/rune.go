package parse

import (
	"strings"
	"unicode"
)

// Rune matches a single rune.
func Rune(r rune) Parser[string] {
	return stringParser{
		Match: string(r),
	}
}

type runeWhereParser struct {
	F func(r rune) bool
}

func (p runeWhereParser) Parse(in *Input) (match string, ok bool, err error) {
	match, ok = in.Peek(1)
	if !ok {
		return
	}
	ok = p.F(rune(match[0]))
	if !ok {
		return
	}
	in.Take(1)
	return
}

// RuneWhere matches a single rune using the given predicate function.
func RuneWhere(predicate func(r rune) bool) Parser[string] {
	return runeWhereParser{
		F: predicate,
	}
}

// RuneIn matches a single rune when the rune is in the string s.
func RuneIn(s string) Parser[string] {
	return RuneWhere(func(r rune) bool { return strings.Contains(s, string(r)) })
}

// RuneNotIn matches a single rune when the rune is not in the string s.
func RuneNotIn(s string) Parser[string] {
	return RuneWhere(func(r rune) bool { return !strings.Contains(s, string(r)) })
}

// RuneInRanges matches a single rune when the rune is withig one of the given Unicode ranges.
func RuneInRanges(ranges ...*unicode.RangeTable) Parser[string] {
	return RuneWhere(func(r rune) bool { return unicode.IsOneOf(ranges, r) })
}

// AnyRune matches any single rune.
var AnyRune = RuneWhere(func(r rune) bool { return true })

// Letter returns a parser which accepts a rune within the Letter Unicode range.
var Letter = RuneInRanges(unicode.Letter)
