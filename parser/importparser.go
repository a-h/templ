package parser

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// Import.
func newImportParser() importParser {
	return importParser{}
}

type importParser struct {
}

func (p importParser) Parse(pi parse.Input) parse.Result {
	var r Import
	// Check the prefix first.
	importStmtPrefix := "import "
	prefixResult := parse.String("{% " + importStmtPrefix)(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have the prefix, we must have an expression and tag end on the same line.
	from := NewPositionFromInput(pi)
	pr := parse.StringUntil(parse.Or(tagEnd, newLine))(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no tag end, the import literal wasn't terminated.
	if !pr.Success {
		return parse.Failure("importParser", newParseError("import: not terminated", from, NewPositionFromInput(pi)))
	}
	r.Expression = NewExpression(pr.Item.(string), from, NewPositionFromInput(pi))

	// Eat the tag end.
	if te := tagEnd(pi); !te.Success {
		return parse.Failure("importParser", newParseError("import: missing end of block (' %}')", from, NewPositionFromInput(pi)))
	}

	return parse.Success("importParser", r, nil)
}
