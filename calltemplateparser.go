package templ

import (
	"io"
	"strings"

	"github.com/a-h/lexical/parse"
)

// newCallTemplateExpressionParser creates a new callTemplateExpressionParser.
func newCallTemplateExpressionParser() callTemplateExpressionParser {
	return callTemplateExpressionParser{}
}

type callTemplateExpressionParser struct{}

func (p callTemplateExpressionParser) Parse(pi parse.Input) parse.Result {
	var r CallTemplateExpression

	// Check the prefix first.
	templPrefix := "call "
	prefixResult := parse.String("{% " + templPrefix)(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have the prefix, we must have a name and parameters.
	// Read the name of the function.
	from := NewPositionFromInput(pi)
	pr := templateNameParser(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, the name wasn't correctly terminated.
	if !pr.Success {
		return parse.Failure("callTemplateExpressionParser", newParseError("call template: invalid template name", from, NewPositionFromInput(pi)))
	}
	// Remove the final "(" from the position.
	to := NewPositionFromInput(pi)
	to.Col -= 1
	to.Index -= 1
	r.Name = NewExpression(strings.TrimSuffix(pr.Item.(string), "("), from, to)

	// Eat the left bracket.
	if lb := parse.Rune('(')(pi); !lb.Success {
		return parse.Failure("callTemplateExpressionParser", newParseError("call template: parameters missing open bracket", from, NewPositionFromInput(pi)))
	}

	// Read the parameters.
	from = NewPositionFromInput(pi)
	pr = parse.StringUntil(parse.Rune(')'))(pi) // p Person, other Other, t thing.Thing)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, the name wasn't correctly terminated.
	if !pr.Success {
		return parse.Failure("callTemplateExpressionParser", newParseError("call template: parameters missing close bracket", from, NewPositionFromInput(pi)))
	}
	r.Arguments = NewExpression(strings.TrimSuffix(pr.Item.(string), ")"), from, NewPositionFromInput(pi))

	// Eat ") %}".
	from = NewPositionFromInput(pi)
	if lb := parse.String(") %}")(pi); !lb.Success {
		return parse.Failure("callTemplateExpressionParser", newParseError("call template: unterminated (missing ' %}')", from, NewPositionFromInput(pi)))
	}

	return parse.Success("callTemplate", r, nil)
}
