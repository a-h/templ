package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// Package.
var pkg packageParser

type packageParser struct {
}

var packageExpressionStartParser = parse.String("package ")

func (p packageParser) Parse(pi parse.Input) parse.Result {
	// Check the prefix first.
	from := NewPositionFromInput(pi)
	prefixResult := packageExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have the prefix, we must have an expression and tag end on the same line.
	pr := parse.StringUntil(newLine)(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no newline, the package literal wasn't terminated.
	if !pr.Success || len(pr.Item.(string)) == 0 {
		return parse.Failure("packageParser", newParseError("package literal not terminated", from, NewPositionFromInput(pi)))
	}

	// Success!
	to := NewPositionFromInput(pi)
	r := Package{
		Expression: NewExpression(pr.Item.(string), from, to),
	}

	return parse.Success("packageParser", r, nil)
}
