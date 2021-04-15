package templ

import (
	"unicode"

	"github.com/a-h/lexical/parse"
)

// Constants.
var tagStart = parse.String("{% ")
var tagEnd = parse.String(" %}")
var newLine = parse.Rune('\n')

// Package.
func asPackage(parts []interface{}) (result interface{}, ok bool) {
	result = Package{
		Expression: parts[2].(string),
	}
	return result, true
}

var packageParser = parse.All(asPackage,
	tagStart,
	parse.String("package "),
	parse.StringUntil(tagEnd),
)

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
		Expression: parts[2].(string),
	}
	return result, true
}

var importParser = parse.All(asImport,
	tagStart,
	parse.String("import "),
	parse.StringUntil(tagEnd),
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
var templateExpressionParametersParser = parse.StringUntil(parse.Rune(')'))

var templateEndParser = parse.String("{% endtmpl %}")

func asTemplate(parts []interface{}) (result interface{}, ok bool) {
	t := Template{
		Name:                parts[2].(string),
		ParameterExpression: parts[3].(string),
	}
	t.Children = parts[7].([]Node)
	return t, true
}

var templateParser = parse.All(asTemplate,
	tagStart,                            // {%
	templateExpressionInstructionParser, // templ
	templateExpressionNameParser,        // FuncName(
	templateExpressionParametersParser,  // p Person, other Other, t thing.Thing)
	parse.String(")"),                   // )
	tagEnd,                              //  %}
	newLine,                             // \n
	templateNodeParser{}.Parse,          // template whitespace, if/switch/for, or node string expression
	templateEndParser,                   // {% endtempl %}
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
	parse.StringUntil(parse.String(" %}")),
	parse.String(" %}"),
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
		parse.StringUntil(parse.String(" %}")),
		parse.String(" %}"),
		newLine,
		templateNodeParser{}.Parse, // if contents
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
		parse.StringUntil(parse.String(" %}")),
		parse.String(" %}"),
		newLine,
		templateNodeParser{}.Parse,   // for contents
		parse.String("{% endfor %}"), // endfor
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
	return parse.AtLeast(p.asTemplateNodeArray, 0, parse.Any(elementParser{}.Parse, whitespaceParser, stringExpressionParser, ifExpressionParser{}.Parse, forExpressionParser{}.Parse))(pi)
}
