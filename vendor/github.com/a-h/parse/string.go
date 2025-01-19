package parse

import "strings"

type stringParser struct {
	Match       string
	Insensitive bool
}

func (p stringParser) Parse(in *Input) (match string, ok bool, err error) {
	match, ok = in.Peek(len(p.Match))
	if !ok {
		return
	}
	if p.Insensitive {
		ok = strings.EqualFold(p.Match, match)
	} else {
		ok = p.Match == match
	}
	if !ok {
		return "", false, nil
	}
	in.Take(len(p.Match))
	return
}

// String matches a given string constant.
func String(s string) Parser[string] {
	return stringParser{
		Match: s,
	}
}

// StringInsensitive matches a given string constant using Unicode case folding.
func StringInsensitive(s string) Parser[string] {
	return stringParser{
		Match:       s,
		Insensitive: true,
	}
}
