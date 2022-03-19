package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// Package.
func newPackageParser() packageParser {
	return packageParser{}
}

type packageParser struct {
}

var packageExpressionStartParser = createStartParser("package")

func (p packageParser) Parse(pi parse.Input) parse.Result {
	// Check the prefix first.
	prefixResult := packageExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have the prefix, we must have an expression and tag end on the same line.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.Or(newLine, expressionEnd))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no tag end, the package literal wasn't terminated.
	if !pr.Success {
		return parse.Failure("packageParser", newParseError("package literal not terminated", from, NewPositionFromInput(pi)))
	}

	// Success!
	to := NewPositionFromInput(pi)
	r := Package{
		Expression: NewExpression(pr.Item.(string), from, to),
	}

	// Eat the tag end.
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("packageParser", newParseError("package literal not terminated", from, NewPositionFromInput(pi)))
	}

	return parse.Success("packageParser", r, nil)
}
