package parse

import (
	"unicode"
)

// Whitespace parses whitespace.
var Whitespace Parser[string] = StringFrom(OneOrMore(RuneInRanges(unicode.White_Space)))

// OptionalWhitespace parses optional whitespace.
var OptionalWhitespace = Func(func(in *Input) (output string, ok bool, err error) {
	output, ok, err = Whitespace.Parse(in)
	if err != nil {
		return
	}
	return output, true, nil
})

// CR is a carriage return.
var CR = Rune('\r')

// CR parses a line feed, used by Unix systems as the newline.
var LF = Rune('\n')

// CRLF parses a carriage returned, followed by a line feed, used by Windows systems as the newline.
var CRLF = String("\r\n")

// NewLine matches either a Windows or Unix line break character.
var NewLine = Any(CRLF, LF)
