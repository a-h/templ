package parser

import (
	"io"
	"strings"

	"github.com/a-h/lexical/parse"
)

func newScriptTemplateParser() scriptTemplateParser {
	return scriptTemplateParser{}
}

type scriptTemplateParser struct {
}

func (p scriptTemplateParser) Parse(pi parse.Input) parse.Result {
	var r ScriptTemplate

	// Parse the name.
	pr := newScriptExpressionParser().Parse(pi)
	if !pr.Success {
		return pr
	}
	r.Name = pr.Item.(scriptExpression).Name
	r.Parameters = pr.Item.(scriptExpression).Parameters

	// Read code expression.
	sr := exp.Parse(pi)
	if sr.Error != nil {
		return sr
	}
	if sr.Success {
		r.Value = sr.Item.(string)
	}

	// Try for }
	pr, ok := chompBrace(pi)
	if !ok {
		return pr
	}
	return parse.Success("script", r, nil)
}

// script Func() {
type scriptExpression struct {
	Name       Expression
	Parameters Expression
}

func newScriptExpressionParser() scriptExpressionParser {
	return scriptExpressionParser{}
}

type scriptExpressionParser struct {
}

var scriptExpressionNameParser = parse.All(parse.WithStringConcatCombiner,
	parse.Letter,
	parse.Many(parse.WithStringConcatCombiner, 0, 1000, parse.Any(parse.Letter, parse.ZeroToNine)),
)

var scriptExpressionStartParser = parse.String("script ")

func (p scriptExpressionParser) Parse(pi parse.Input) parse.Result {
	var r scriptExpression

	// Check the prefix first.
	prefixResult := scriptExpressionStartParser(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we have the prefix, we must have a name and parameters.
	// Read the name of the function.
	from := NewPositionFromInput(pi)
	pr := scriptExpressionNameParser(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, the name wasn't correctly terminated.
	if !pr.Success {
		return parse.Failure("scriptExpressionParser", newParseError("script expression: invalid name", from, NewPositionFromInput(pi)))
	}
	to := NewPositionFromInput(pi)
	r.Name = NewExpression(pr.Item.(string), from, to)
	from = to

	// Eat the open bracket.
	if lb := parse.Rune('(')(pi); !lb.Success {
		return parse.Failure("scriptExpressionParser", newParseError("script expression: parameters missing open bracket", from, NewPositionFromInput(pi)))
	}

	// Read the parameters.
	from = NewPositionFromInput(pi)
	pr = parse.StringUntil(parse.Rune(')'))(pi) // p Person, other Other, t thing.Thing)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	// If there's no match, the name wasn't correctly terminated.
	if !pr.Success {
		return parse.Failure("scriptExpressionParser", newParseError("script expression: parameters missing close bracket", from, NewPositionFromInput(pi)))
	}
	r.Parameters = NewExpression(strings.TrimSuffix(pr.Item.(string), ")"), from, NewPositionFromInput(pi))

	// Eat ") {".
	from = NewPositionFromInput(pi)
	if lb := expressionFuncEnd(pi); !lb.Success {
		return parse.Failure("scriptExpressionParser", newParseError("script expression: unterminated (missing ') {')", from, NewPositionFromInput(pi)))
	}

	// Expect a newline.
	from = NewPositionFromInput(pi)
	if lb := newLine(pi); !lb.Success {
		return parse.Failure("scriptExpressionParser", newParseError("script expression: missing terminating newline", from, NewPositionFromInput(pi)))
	}

	return parse.Success("scriptExpressionParser", r, nil)
}
