package parse

import (
	"regexp"
)

type regexpParser struct {
	Expression *regexp.Regexp
}

func (p regexpParser) Parse(in *Input) (match string, ok bool, err error) {
	remainder, ok := in.Peek(-1)
	if !ok {
		return
	}
	startAndEndIndex := p.Expression.FindStringIndex(remainder)
	ok = startAndEndIndex != nil && startAndEndIndex[0] == 0
	if !ok {
		return
	}
	match = remainder[startAndEndIndex[0]:startAndEndIndex[1]]
	in.Take(len(match))
	return
}

// Regexp creates a parser that parses from the input's current position, or fails.
func Regexp(exp string) (p Parser[string], err error) {
	r, err := regexp.Compile(exp)
	if err != nil {
		return
	}
	p = regexpParser{
		Expression: r,
	}
	return
}

// MustRegexp creates a parse that parses from the input's current position.
// Passing in a regular expression that doesn't compile will result in a panic.
func MustRegexp(exp string) (p Parser[string]) {
	p, err := Regexp(exp)
	if err != nil {
		panic(err)
	}
	return
}
