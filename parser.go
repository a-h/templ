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

func asTemplate(parts []interface{}) (result interface{}, ok bool) {
	t := Template{
		Name:                parts[1].(string),
		ParameterExpression: parts[2].(string),
	}
	t.Children = parts[5].([]Node)
	return t, true
}

var templateParser = parse.All(asTemplate,
	parse.String("{% templ "),          // {% templ
	templateExpressionNameParser,       // FuncName(
	parse.StringUntil(parse.Rune(')')), // p Person, other Other, t thing.Thing)
	parse.String(") %}"),               // ) %}
	newLine,                            // \r\n or \n
	templateNodeParser{}.Parse,         // template whitespace, if/switch/for, or node string expression
	parse.String("{% endtmpl %}"),      // {% endtempl %}
	parse.Optional(parse.WithStringConcatCombiner, whitespaceParser),
)

// Element.

// Element name.
var lowerAZ = "abcdefghijklmnopqrstuvwxyz"
var upperAZ = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
var numbers = "0123456789"
var elementNameParser = parse.Then(parse.WithStringConcatCombiner,
	parse.RuneIn(lowerAZ),
	parse.Many(parse.WithStringConcatCombiner, 0, 15, parse.RuneIn(lowerAZ+upperAZ+numbers)),
)

// Element self-closing tag.
type elementSelfClosing struct {
	Name       string
	Attributes []Attribute
}

func asElementSelfClosing(parts []interface{}) (result interface{}, ok bool) {
	return elementSelfClosing{
		Name:       parts[1].(string),
		Attributes: parts[2].([]Attribute),
	}, true
}

var elementSelfClosingParser = parse.All(asElementSelfClosing,
	parse.Rune('<'),
	elementNameParser,
	attributesParser,
	parse.Optional(parse.WithStringConcatCombiner, whitespaceParser),
	parse.String("/>"),
)

// Element open tag.
type elementOpenTag struct {
	Name       string
	Attributes []Attribute
}

func asElementOpenTag(parts []interface{}) (result interface{}, ok bool) {
	return elementOpenTag{
		Name:       parts[1].(string),
		Attributes: parts[2].([]Attribute),
	}, true
}

var elementOpenTagParser = parse.All(asElementOpenTag,
	parse.Rune('<'),
	elementNameParser,
	attributesParser,
	parse.Optional(parse.WithStringConcatCombiner, whitespaceParser),
	parse.Rune('>'),
)

// Element close tag.
type elementCloseTag struct {
	Name string
}

func asElementCloseTag(parts []interface{}) (result interface{}, ok bool) {
	return elementCloseTag{
		Name: parts[1].(string),
	}, true
}

var elementCloseTagParser = parse.All(asElementCloseTag,
	parse.String("</"),
	elementNameParser,
	parse.Rune('>'),
)

// Attribute name.
var attributeNameParser = parse.StringUntil(parse.Rune('='))

// Constant attribute.
var attributeConstantValueParser = parse.StringUntil(parse.Rune('"'))

func asConstantAttribute(parts []interface{}) (result interface{}, ok bool) {
	return ConstantAttribute{
		Name:  parts[1].(string),
		Value: parts[4].(string),
	}, true
}

var constAttrParser = parse.All(asConstantAttribute,
	whitespaceParser,
	attributeNameParser,
	parse.Rune('='),
	parse.Rune('"'),
	attributeConstantValueParser,
	parse.Rune('"'),
)

func asAttributeArray(parts []interface{}) (result interface{}, ok bool) {
	op := make([]Attribute, len(parts))
	for i := 0; i < len(parts); i++ {
		switch v := parts[i].(type) {
		case ConstantAttribute:
			op[i] = v
		case ExpressionAttribute:
			op[i] = v
		}
	}
	return op, true
}

// ExpressionAttribute.
func asExpressionAttribute(parts []interface{}) (result interface{}, ok bool) {
	return ExpressionAttribute{
		Name:  parts[1].(string),
		Value: parts[3].(StringExpression),
	}, true
}

var exprAttrParser = parse.All(asExpressionAttribute,
	whitespaceParser,
	attributeNameParser,
	parse.Rune('='),
	stringExpressionParser,
)

var attributesParser = parse.Many(asAttributeArray, 0, 255, parse.Or(exprAttrParser, constAttrParser))

// Element with children.
type elementParser struct{}

func (p elementParser) asElement(parts []interface{}) (result interface{}, ok bool) {
	e := Element{
		Name:       parts[0].(elementOpenTag).Name,
		Attributes: parts[0].(elementOpenTag).Attributes,
	}
	for i := 1; i < len(parts); i++ {
		if arr, isArray := parts[i].([]interface{}); isArray {
			for j := 0; j < len(arr); j++ {
				// Handle all possible allowed element children.
				if ce, isElement := arr[j].(Element); isElement {
					e.Children = append(e.Children, ce)
				}
				if ce, isSelfClosing := arr[j].(elementSelfClosing); isSelfClosing {
					e.Children = append(e.Children, Element{
						Name:       ce.Name,
						Attributes: ce.Attributes,
					})
				}
				if nse, isStringExpression := arr[j].(StringExpression); isStringExpression {
					e.Children = append(e.Children, nse)
				}
			}
		}
	}
	return e, true
}

func (p elementParser) asChildren(parts []interface{}) (result interface{}, ok bool) {
	return parts, true
}

func (p elementParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asElement,
		elementOpenTagParser,
		parse.Many(p.asChildren, 0, 100,
			// All of the allowed children.
			parse.Any(stringExpressionParser, elementSelfClosingParser, p.Parse, whitespaceParser),
		),
		elementCloseTagParser,
	)(pi)
}

// StringExpression.
func asStringExpression(parts []interface{}) (result interface{}, ok bool) {
	return StringExpression{
		Expression: parts[1].(string),
	}, true
}

var stringExpressionParser = parse.All(asStringExpression,
	parse.String("{%= "),
	parse.StringUntil(tagEnd),
	tagEnd,
)

// IfExpression.
type ifExpressionParser struct{}

func (p ifExpressionParser) asIfExpression(parts []interface{}) (result interface{}, ok bool) {
	return IfExpression{
		Expression: parts[1].(string),
		Then:       parts[4].([]Node),
		Else:       parts[5].([]Node),
	}, true
}

func (p ifExpressionParser) asChildren(parts []interface{}) (result interface{}, ok bool) {
	if len(parts) == 0 {
		return []Node{}, true
	}
	return parts[0].([]Node), true
}

func (p ifExpressionParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asIfExpression,
		parse.String("{% if "),
		parse.StringUntil(tagEnd),
		tagEnd,
		parse.Optional(parse.WithStringConcatCombiner, newLine),
		templateNodeParser{}.Parse,                                 // if contents
		parse.Optional(p.asChildren, elseExpressionParser{}.Parse), // else
		parse.String("{% endif %}"),                                // endif
	)(pi)
}

type elseExpressionParser struct{}

func (p elseExpressionParser) asElseExpression(parts []interface{}) (result interface{}, ok bool) {
	return parts[1].([]Node), true // the array of nodes from templateNodeParser
}

func (p elseExpressionParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asElseExpression,
		parse.String("{% else %}"),
		templateNodeParser{}.Parse, // else contents
	)(pi)
}

// ForExpression.
type forExpressionParser struct{}

func (p forExpressionParser) asForExpression(parts []interface{}) (result interface{}, ok bool) {
	return ForExpression{
		Expression: parts[1].(string),
		Children:   parts[4].([]Node),
	}, true
}

func (p forExpressionParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asForExpression,
		parse.String("{% for "),
		parse.StringUntil(tagEnd),
		tagEnd,
		newLine,
		templateNodeParser{}.Parse,   // for contents
		parse.String("{% endfor %}"), // endfor
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
type templateNodeParser struct{}

func (p templateNodeParser) asTemplateNodeArray(parts []interface{}) (result interface{}, ok bool) {
	op := make([]Node, len(parts))
	for i := 0; i < len(parts); i++ {
		op[i] = parts[i].(Node)
	}
	return op, true
}

func (p templateNodeParser) Parse(pi parse.Input) parse.Result {
	return parse.AtLeast(p.asTemplateNodeArray, 0, parse.Any(elementParser{}.Parse, whitespaceParser, stringExpressionParser, ifExpressionParser{}.Parse, forExpressionParser{}.Parse, callTemplateExpressionParser{}.Parse))(pi)
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
	SourceRangeToItemLookup SourceRangeToItemLookup
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
	tr := parse.Many(p.asTemplateArray, 0, 1, templateParser)(pi)
	if tr.Error != nil {
		return tr
	}
	tf.Templates = tr.Item.([]Template)

	// Success.
	return parse.Success("template file", tf, nil)
}

func ParseString(template string) (TemplateFile, SourceRangeToItemLookup, error) {
	srl := SourceRangeToItemLookup{}
	tfr := TemplateFileParser{
		SourceRangeToItemLookup: srl,
	}.Parse(input.NewFromString(template))
	if tfr.Error != nil {
		return TemplateFile{}, srl, tfr.Error
	}
	return tfr.Item.(TemplateFile), srl, nil
}
