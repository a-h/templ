package templ

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// Package.
func newPackageParser(sril SourceRangeToItemLookup) *packageParser {
	return &packageParser{
		SourceRangeToItemLookup: sril,
	}
}

type packageParser struct {
	SourceRangeToItemLookup SourceRangeToItemLookup
}

func (pp *packageParser) asPackage(parts []interface{}) (result interface{}, ok bool) {
	result = Package{
		Expression: parts[1].(string),
	}
	return result, true
}

func (pp *packageParser) Parse(pi parse.Input) parse.Result {
	from := NewPositionFromInput(pi)

	// Check the prefix first.
	prefixResult := parse.String("{% package")(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have the prefix, we must have an expression and tag end.
	pr := parse.All(pp.asPackage,
		parse.Rune(' '),
		parse.StringUntil(parse.Or(tagEnd, newLine)),
		tagEnd)(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	if !pr.Success {
		return parse.Failure("packageParser", newParseError("package literal not terminated", from, NewPositionFromInput(pi)))
	}
	p := pr.Item.(Package)
	from = pp.SourceRangeToItemLookup.Add(p, from, NewPositionFromInput(pi))
	return pr
}
