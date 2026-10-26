package jstokenizer

import (
	"unicode/utf8"

	"github.com/a-h/parse"
)

// utf8RuneParser matches a single UTF-8 encoded rune for which Predicate returns true.
// parse.RuneWhere matches a single byte, so it splits a multi-byte rune such as U+2028
// into bytes, and reads a byte such as 0xE2 as the letter â.
type utf8RuneParser struct {
	Predicate func(r rune) bool
}

func (p utf8RuneParser) Parse(pi *parse.Input) (match string, ok bool, err error) {
	remaining, ok := pi.Peek(-1)
	if !ok || remaining == "" {
		return "", false, nil
	}
	r, size := utf8.DecodeRuneInString(remaining)
	if !p.Predicate(r) {
		return "", false, nil
	}
	match, _ = pi.Take(size)
	return match, true, nil
}
