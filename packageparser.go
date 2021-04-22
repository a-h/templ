package templ

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

func (p packageParser) Parse(pi parse.Input) parse.Result {
	// Check the prefix first.
	packageStmtPrefix := "package "
	prefixResult := parse.String("{% " + packageStmtPrefix)(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have the prefix, we must have an expression and tag end on the same line.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.Or(tagEnd, newLine))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no tag end, the package literal wasn't terminated.
	if !pr.Success {
		return parse.Failure("packageParser", newParseError("package literal not terminated", from, NewPositionFromInput(pi)))
	}

	// Success!
	// Include "package " in the Go expression.
	from.Col -= len(packageStmtPrefix)
	to := NewPositionFromInput(pi)
	r := Package{
		Expression: NewExpression(packageStmtPrefix+pr.Item.(string), from, to),
	}

	// Eat the tag end.
	if te := tagEnd(pi); !te.Success {
		return te
	}

	return parse.Success("packageParser", r, nil)
}
