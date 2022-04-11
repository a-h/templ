package parser

import (
	"fmt"
	"io"
	"unicode"

	"github.com/a-h/lexical/parse"
)

// Constants.
// {
var expressionEnd = parse.Or(parse.String(" {"), parse.String("{"))

// ) {
var expressionFuncEnd = parse.All(asNil, parse.Rune(')'), expressionEnd)

func asNil(inputs []interface{}) (interface{}, bool) {
	return nil, true
}

// create a parser for `name`
// Deprecated: just use a string "name"
func createStartParser(name string) parse.Function {
	return parse.Or(parse.String("{ "+name+" "), parse.String("{"+name+" "))
}

var newLine = parse.Or(parse.String("\r\n"), parse.Rune('\n'))

// Whitespace.
func asWhitespace(parts []interface{}) (result interface{}, ok bool) {
	var w Whitespace
	for _, ip := range parts {
		w.Value += string(ip.(rune))
	}
	return w, true
}

var whitespaceParser = parse.AtLeast(asWhitespace, 1, parse.RuneInRanges(unicode.White_Space))
var optionalWhitespaceParser = parse.AtLeast(asWhitespace, 0, parse.RuneInRanges(unicode.White_Space))

// Template

var template templateParser

type templateParser struct {
}

func (p templateParser) Parse(pi parse.Input) parse.Result {
	var r HTMLTemplate

	// templ FuncName(p Person, other Other) {
	tepr := newTemplateExpressionParser().Parse(pi)
	if !tepr.Success {
		return tepr
	}
	te := tepr.Item.(templateExpression)
	r.Name = te.Name
	r.Parameters = te.Parameters

	// Once we're in a template, we should expect some template whitespace, if/switch/for,
	// or node string expressions etc.
	from := NewPositionFromInput(pi)
	tnpr := newTemplateNodeParser(closeBraceWithOptionalPadding).Parse(pi)
	if tnpr.Error != nil && tnpr.Error != io.EOF {
		return tnpr
	}
	// If there's no match, there's no template elements.
	if !tnpr.Success {
		return parse.Failure("templateParser", newParseError("templ: expected nodes in templ body, but found none", from, NewPositionFromInput(pi)))
	}
	r.Children = tnpr.Item.([]Node)

	// Eat any whitespace.
	pr := optionalWhitespaceParser(pi)
	if pr.Error != nil {
		return pr
	}

	// Try for }
	pr, ok := chompBrace(pi)
	if !ok {
		return pr
	}
	return parse.Success("templ", r, nil)
}

// Parse error.
func newParseError(msg string, from Position, to Position) ParseError {
	return ParseError{
		Message: msg,
		From:    from,
		To:      to,
	}
}

// ParseError details where the error occurred in the file.
type ParseError struct {
	Message string
	From    Position
	To      Position
}

func (pe ParseError) Error() string {
	return fmt.Sprintf("%v at %v", pe.Message, pe.From)
}
