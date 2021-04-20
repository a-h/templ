package templ

import (
	"fmt"
	"unicode"

	"github.com/a-h/lexical/input"
	"github.com/a-h/lexical/parse"
)

// Constants.
var tagEnd = parse.String(" %}")
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

// Import.
func asImport(parts []interface{}) (result interface{}, ok bool) {
	result = Import{
		Expression: parts[1].(string),
	}
	return result, true
}

var importParser = parse.All(asImport,
	parse.String("{% import "),
	parse.StringUntil(tagEnd),
	tagEnd,
)

// Template

// a-Z followed by 0-9[a-Z]
var templateNameParser = parse.All(parse.WithStringConcatCombiner,
	parse.Letter,
	parse.Many(parse.WithStringConcatCombiner, 0, 1000, parse.Any(parse.Letter, parse.ZeroToNine)),
)
var templateExpressionInstructionParser = parse.String("templ ")

func asTemplateExpressionName(parts []interface{}) (result interface{}, ok bool) {
	return parts[0].(string), true
}

var templateExpressionNameParser = parse.All(asTemplateExpressionName, templateNameParser, parse.Rune('('))

func newTemplateParser(sril *SourceRangeToItemLookup) *templateParser {
	return &templateParser{
		SourceRangeToItemLookup: sril,
	}
}

type templateParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *templateParser) asTemplate(parts []interface{}) (result interface{}, ok bool) {
	t := Template{
		Name:                parts[1].(string),
		ParameterExpression: parts[2].(string),
	}
	t.Children = parts[5].([]Node)
	return t, true
}

func (p *templateParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asTemplate,
		parse.String("{% templ "),          // {% templ
		templateExpressionNameParser,       // FuncName(
		parse.StringUntil(parse.Rune(')')), // p Person, other Other, t thing.Thing)
		parse.String(") %}"),               // ) %}
		newLine,                            // \r\n or \n
		newTemplateNodeParser(p.SourceRangeToItemLookup).Parse, // template whitespace, if/switch/for, or node string expression
		parse.String("{% endtmpl %}"),                          // {% endtempl %}
		parse.Optional(parse.WithStringConcatCombiner, whitespaceParser),
	)(pi)
}

// IfExpression.
func newIfExpressionParser(sril *SourceRangeToItemLookup) *ifExpressionParser {
	return &ifExpressionParser{
		SourceRangeToItemLookup: sril,
	}
}

type ifExpressionParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *ifExpressionParser) asIfExpression(parts []interface{}) (result interface{}, ok bool) {
	return IfExpression{
		Expression: parts[1].(string),
		Then:       parts[4].([]Node),
		Else:       parts[5].([]Node),
	}, true
}

func (p *ifExpressionParser) asChildren(parts []interface{}) (result interface{}, ok bool) {
	if len(parts) == 0 {
		return []Node{}, true
	}
	return parts[0].([]Node), true
}

func (p *ifExpressionParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asIfExpression,
		parse.String("{% if "),
		parse.StringUntil(tagEnd),
		tagEnd,
		parse.Optional(parse.WithStringConcatCombiner, newLine),
		newTemplateNodeParser(p.SourceRangeToItemLookup).Parse,                                 // if contents
		parse.Optional(p.asChildren, newElseExpressionParser(p.SourceRangeToItemLookup).Parse), // else
		parse.String("{% endif %}"),                                                            // endif
	)(pi)
}

func newElseExpressionParser(sril *SourceRangeToItemLookup) *elseExpressionParser {
	return &elseExpressionParser{
		SourceRangeToItemLookup: sril,
	}
}

type elseExpressionParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *elseExpressionParser) asElseExpression(parts []interface{}) (result interface{}, ok bool) {
	return parts[1].([]Node), true // the array of nodes from templateNodeParser
}

func (p *elseExpressionParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asElseExpression,
		parse.String("{% else %}"),
		newTemplateNodeParser(p.SourceRangeToItemLookup).Parse, // else contents
	)(pi)
}

// ForExpression.
func newForExpressionParser(sril *SourceRangeToItemLookup) *forExpressionParser {
	return &forExpressionParser{
		SourceRangeToItemLookup: sril,
	}
}

type forExpressionParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *forExpressionParser) asForExpression(parts []interface{}) (result interface{}, ok bool) {
	return ForExpression{
		Expression: parts[1].(string),
		Children:   parts[4].([]Node),
	}, true
}

func (p *forExpressionParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asForExpression,
		parse.String("{% for "),
		parse.StringUntil(tagEnd),
		tagEnd,
		newLine,
		newTemplateNodeParser(p.SourceRangeToItemLookup).Parse, // for contents
		parse.String("{% endfor %}"),                           // endfor
	)(pi)
}

// CallTemplateExpressionParser.
type callTemplateExpressionParser struct{}

func (p callTemplateExpressionParser) asCallTemplateExpression(parts []interface{}) (result interface{}, ok bool) {
	return CallTemplateExpression{
		Name:               parts[1].(string),
		ArgumentExpression: parts[3].(string),
	}, true
}

func (p callTemplateExpressionParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asCallTemplateExpression,
		parse.String("{% call "),                // {% call
		parse.StringUntil(parse.String("(")),    // TemplateName
		parse.Rune('('),                         // (
		parse.StringUntil(parse.String(") %}")), // p.Test, p.Other()
		parse.String(") %}"),                    // ) %}
	)(pi)
}

// Template node (element, call, if, switch, for, whitespace etc.)
func newTemplateNodeParser(sril *SourceRangeToItemLookup) *templateNodeParser {
	return &templateNodeParser{
		SourceRangeToItemLookup: sril,
	}
}

type templateNodeParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *templateNodeParser) asTemplateNodeArray(parts []interface{}) (result interface{}, ok bool) {
	op := make([]Node, len(parts))
	for i := 0; i < len(parts); i++ {
		op[i] = parts[i].(Node)
	}
	return op, true
}

func (p *templateNodeParser) Parse(pi parse.Input) parse.Result {
	return parse.AtLeast(p.asTemplateNodeArray, 0, parse.Any(
		newElementParser(p.SourceRangeToItemLookup).Parse, // <a>, <br/> etc.
		whitespaceParser,
		newStringExpressionParser(p.SourceRangeToItemLookup).Parse, // {%= strings.ToUpper("abc") %}
		newIfExpressionParser(p.SourceRangeToItemLookup).Parse,     // if {}
		newForExpressionParser(p.SourceRangeToItemLookup).Parse,    // for {}
		callTemplateExpressionParser{}.Parse,                       // {% call TemplateName(a, b, c) %}
	))(pi)
}

// Parse error.
func newParseError(msg string, from Position, to Position) parseError {
	return parseError{
		Message: msg,
		From:    from,
		To:      to,
	}
}

type parseError struct {
	Message string
	From    Position
	To      Position
}

func (pe parseError) Error() string {
	return fmt.Sprintf("%v from %v to %v", pe.Message, pe.From, pe.To)
}

// TemplateFile.
type TemplateFileParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p TemplateFileParser) asImportArray(parts []interface{}) (result interface{}, ok bool) {
	if len(parts) == 0 {
		return []Import{}, true
	}
	return parts[0].([]Import), true
}

func (p TemplateFileParser) asTemplateArray(parts []interface{}) (result interface{}, ok bool) {
	if len(parts) == 0 {
		return []Template{}, true
	}
	return parts[0].([]Template), true
}

func (p TemplateFileParser) Parse(pi parse.Input) parse.Result {
	var tf TemplateFile
	from := NewPositionFromInput(pi)

	// Required package.
	// {% package name %}
	pr := newPackageParser(p.SourceRangeToItemLookup).Parse(pi)
	if pr.Error != nil {
		return pr
	}
	if !pr.Success {
		return parse.Failure("TemplateFileParser", newParseError("package not found", from, NewPositionFromInput(pi)))
	}
	tf.Package = pr.Item.(Package)
	//TODO: Think about making sure they don't cross over.
	from = p.SourceRangeToItemLookup.Add(tf.Package, from, NewPositionFromInput(pi))

	// Optional whitespace.
	parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi)

	// Optional imports.
	// {% import "strings" %}
	ipr := parse.Many(p.asImportArray, 0, -1, importParser)(pi)
	if ipr.Error != nil {
		return ipr
	}
	tf.Imports = ipr.Item.([]Import)

	// Optional whitespace.
	parse.Optional(parse.WithStringConcatCombiner, whitespaceParser)(pi)

	// Optional templates.
	// {% templ Name(p Parameter) %}
	tr := parse.Optional(p.asTemplateArray,
		newTemplateParser(p.SourceRangeToItemLookup).Parse,
	)(pi)
	if tr.Error != nil {
		return tr
	}
	tf.Templates = tr.Item.([]Template)

	// Success.
	return parse.Success("template file", tf, nil)
}

func ParseString(template string) (TemplateFile, *SourceRangeToItemLookup, error) {
	srl := NewSourceRangeToItemLookup()
	tfr := TemplateFileParser{
		SourceRangeToItemLookup: srl,
	}.Parse(input.NewFromString(template))
	if tfr.Error != nil {
		return TemplateFile{}, srl, tfr.Error
	}
	return tfr.Item.(TemplateFile), srl, nil
}
