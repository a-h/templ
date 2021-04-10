package templ

import (
	"unicode"

	"github.com/a-h/lexical/parse"
)

// Constants.
var tagStart = parse.String("{% ")
var tagEnd = parse.String(" %}")
var newLine = parse.Rune('\n')

// PackageExpression.
func asPackageExpression(parts []interface{}) (result interface{}, ok bool) {
	result = PackageExpression{
		Expression: parts[2].(string),
	}
	return result, true
}

var packageParser = parse.All(asPackageExpression,
	tagStart,
	parse.String("package "),
	parse.StringUntil(tagEnd),
)

// ImportExpression.
func asImportExpression(parts []interface{}) (result interface{}, ok bool) {
	result = ImportExpression{
		Expression: parts[2].(string),
	}
	return result, true
}

var importParser = parse.All(asImportExpression,
	tagStart,
	parse.String("import "),
	parse.StringUntil(tagEnd),
)

// TemplateFileWhitespace.
type templateFileWhitespace struct {
	Text string
}

func asTemplateFileWhitespace(parts []interface{}) (result interface{}, ok bool) {
	result, ok = parse.WithStringConcatCombiner(parts)
	if !ok {
		return
	}
	result = templateFileWhitespace{
		Text: result.(string),
	}
	return
}

var templateFileWhitespaceParser = parse.AtLeast(asTemplateFileWhitespace, 1, parse.RuneInRanges(unicode.White_Space))

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

func asTemplate(parts []interface{}) (result interface{}, ok bool) {
	result = Template{
		Name:                parts[2].(string),
		ParameterExpression: parts[3].(string),
	}
	return result, true
}

var templateWhitespaceParser = parse.AtLeast(parse.WithStringConcatCombiner, 1, parse.RuneInRanges(unicode.White_Space))

var templateEndParser = parse.String("{% endtmpl %}")

var templateParser = parse.All(asTemplate,
	tagStart,                            // {%
	templateExpressionInstructionParser, // templ
	templateExpressionNameParser,        // FuncName(
	templateExpressionParametersParser,  // p Person, other Other, t thing.Thing)
	tagEnd,                              //  %}
	newLine,                             // \n
	templateWhitespaceParser,            // whitespace?
	templateEndParser,
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
	parse.Optional(parse.WithStringConcatCombiner, attributeWhitespaceParser),
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
	parse.Optional(parse.WithStringConcatCombiner, attributeWhitespaceParser),
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
var attributeWhitespaceParser = parse.AtLeast(parse.WithStringConcatCombiner, 1, parse.RuneInRanges(unicode.White_Space))

var attributeConstantValueParser = parse.StringUntil(parse.Rune('"'))

func asConstantAttribute(parts []interface{}) (result interface{}, ok bool) {
	return ConstantAttribute{
		Name:  parts[1].(string),
		Value: parts[3].(string),
	}, true
}

var constAttrParser = parse.All(asConstantAttribute,
	attributeWhitespaceParser,
	attributeNameParser,
	parse.Rune('"'),
	attributeConstantValueParser,
)

//TODO: Add in the expression attribute later.
func asAttributeList(parts []interface{}) (result interface{}, ok bool) {
	op := make([]Attribute, len(parts))
	for i := 0; i < len(parts); i++ {
		op[i] = parts[i].(ConstantAttribute)
	}
	return op, true
}

var attributesParser = parse.Many(asAttributeList, 0, 255, constAttrParser)

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
			parse.Any(nodeStringExpressionParser, elementSelfClosingParser, p.Parse),
		),
		elementCloseTagParser,
	)(pi)
}

// NodeStringExpression.
func asNodeStringExpression(parts []interface{}) (result interface{}, ok bool) {
	return NodeStringExpression{
		Expression: parts[1].(string),
	}, true
}

var nodeStringExpressionParser = parse.All(asNodeStringExpression,
	parse.String("{%= "),
	parse.StringUntil(parse.String(" %}")),
)
