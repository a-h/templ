package templ

import (
	"io"

	"github.com/a-h/lexical/parse"
)

// Element.

// Element open tag.
type elementOpenTag struct {
	Name       string
	Attributes []Attribute
}

func newElementOpenTagParser(sril *SourceRangeToItemLookup) *elementOpenTagParser {
	return &elementOpenTagParser{
		SourceRangeToItemLookup: sril,
	}
}

type elementOpenTagParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *elementOpenTagParser) asElementOpenTag(parts []interface{}) (result interface{}, ok bool) {
	return elementOpenTag{
		Name:       parts[1].(string),
		Attributes: parts[2].([]Attribute),
	}, true
}

func (p *elementOpenTagParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asElementOpenTag,
		parse.Rune('<'),
		elementNameParser,
		newAttributesParser(p.SourceRangeToItemLookup).Parse,
		parse.Optional(parse.WithStringConcatCombiner, whitespaceParser),
		parse.Rune('>'),
	)(pi)
}

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

func newConstantAttributeParser(sril *SourceRangeToItemLookup) *constantAttributeParser {
	return &constantAttributeParser{
		SourceRangeToItemLookup: sril,
	}
}

type constantAttributeParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *constantAttributeParser) asConstantAttribute(parts []interface{}) (result interface{}, ok bool) {
	return ConstantAttribute{
		Name:  parts[1].(string),
		Value: parts[4].(string),
	}, true
}

func (p *constantAttributeParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asConstantAttribute,
		whitespaceParser,
		attributeNameParser,
		parse.Rune('='),
		parse.Rune('"'),
		attributeConstantValueParser,
		parse.Rune('"'),
	)(pi)
}

// ExpressionAttribute.
func newExpressionAttributeParser(sril *SourceRangeToItemLookup) *expressionAttributeParser {
	return &expressionAttributeParser{
		SourceRangeToItemLookup: sril,
	}
}

type expressionAttributeParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *expressionAttributeParser) asExpressionAttribute(parts []interface{}) (result interface{}, ok bool) {
	return ExpressionAttribute{
		Name:  parts[1].(string),
		Value: parts[3].(StringExpression),
	}, true
}

func (p *expressionAttributeParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asExpressionAttribute,
		whitespaceParser,
		attributeNameParser,
		parse.Rune('='),
		newStringExpressionParser(p.SourceRangeToItemLookup).Parse,
	)(pi)
}

// Attributes.
func newAttributesParser(sril *SourceRangeToItemLookup) *attributesParser {
	return &attributesParser{
		SourceRangeToItemLookup: sril,
	}
}

type attributesParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *attributesParser) asAttributeArray(parts []interface{}) (result interface{}, ok bool) {
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

func (p *attributesParser) Parse(pi parse.Input) parse.Result {
	return parse.Many(p.asAttributeArray, 0, 255,
		parse.Or(
			newExpressionAttributeParser(p.SourceRangeToItemLookup).Parse,
			newConstantAttributeParser(p.SourceRangeToItemLookup).Parse,
		),
	)(pi)
}

// Element name.
var lowerAZ = "abcdefghijklmnopqrstuvwxyz"
var upperAZ = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
var numbers = "0123456789"
var elementNameParser = parse.Then(parse.WithStringConcatCombiner,
	parse.RuneIn(lowerAZ),
	parse.Many(parse.WithStringConcatCombiner, 0, 15, parse.RuneIn(lowerAZ+upperAZ+numbers)),
)

// Element.
func newElementOpenCloseParser(sril *SourceRangeToItemLookup) *elementOpenCloseParser {
	return &elementOpenCloseParser{
		SourceRangeToItemLookup: sril,
	}
}

type elementOpenCloseParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *elementOpenCloseParser) asElement(parts []interface{}) (result interface{}, ok bool) {
	e := Element{
		Name:       parts[0].(elementOpenTag).Name,
		Attributes: parts[0].(elementOpenTag).Attributes,
	}
	if arr, isArray := parts[1].([]Node); isArray {
		e.Children = append(e.Children, arr...)
	}
	return e, true
}

func (p *elementOpenCloseParser) asChildren(parts []interface{}) (result interface{}, ok bool) {
	return parts, true
}

func (p *elementOpenCloseParser) Parse(pi parse.Input) parse.Result {
	//TODO: Don't use parse.All, check that the close tag name matches the start name!
	return parse.All(p.asElement,
		newElementOpenTagParser(p.SourceRangeToItemLookup).Parse,
		newTemplateNodeParser(p.SourceRangeToItemLookup).Parse,
		elementCloseTagParser,
	)(pi)
}

// Element self-closing tag.
func newElementSelfClosingParser(sril *SourceRangeToItemLookup) elementSelfClosingParser {
	return elementSelfClosingParser{
		SourceRangeToItemLookup: sril,
	}
}

type elementSelfClosingParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p elementSelfClosingParser) asElement(parts []interface{}) (result interface{}, ok bool) {
	return Element{
		Name:       parts[1].(string),
		Attributes: parts[2].([]Attribute),
	}, true
}

func (p elementSelfClosingParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asElement,
		parse.Rune('<'),
		elementNameParser,
		newAttributesParser(p.SourceRangeToItemLookup).Parse,
		parse.Optional(parse.WithStringConcatCombiner, whitespaceParser),
		parse.String("/>"),
	)(pi)
}

// Element
func newElementParser(sril *SourceRangeToItemLookup) *elementParser {
	return &elementParser{
		SourceRangeToItemLookup: sril,
	}
}

type elementParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *elementParser) Parse(pi parse.Input) parse.Result {
	return parse.Or(
		newElementSelfClosingParser(p.SourceRangeToItemLookup).Parse,
		newElementOpenCloseParser(p.SourceRangeToItemLookup).Parse,
	)(pi)
}

// StringExpression.
func newStringExpressionParser(sril *SourceRangeToItemLookup) *stringExpressionParser {
	return &stringExpressionParser{
		SourceRangeToItemLookup: sril,
	}
}

type stringExpressionParser struct {
	SourceRangeToItemLookup *SourceRangeToItemLookup
}

func (p *stringExpressionParser) asStringExpression(parts []interface{}) (result interface{}, ok bool) {
	return StringExpression{
		Expression: parts[0].(string),
	}, true
}

func (p *stringExpressionParser) Parse(pi parse.Input) parse.Result {
	from := NewPositionFromInput(pi)

	// Check the prefix first.
	prefixResult := parse.String("{%= ")(pi)
	if !prefixResult.Success {
		return prefixResult
	}

	// Once we've seen a string expression prefix, we must have a tag end.
	pr := parse.All(p.asStringExpression,
		parse.StringUntil(tagEnd),
		tagEnd)(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return pr
	}
	if !pr.Success {
		return parse.Failure("packageParser", newParseError("package literal not terminated", from, NewPositionFromInput(pi)))
	}
	r := pr.Item.(StringExpression)
	from = p.SourceRangeToItemLookup.Add(r, from, NewPositionFromInput(pi))
	return pr
}
