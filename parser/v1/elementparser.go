package parser

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/a-h/lexical/input"
	"github.com/a-h/lexical/parse"
)

// Element.

// Element open tag.
type elementOpenTag struct {
	Name       string
	Attributes []Attribute
}

func newElementOpenTagParser() elementOpenTagParser {
	return elementOpenTagParser{}
}

type elementOpenTagParser struct {
}

func (p elementOpenTagParser) asElementOpenTag(parts []interface{}) (result interface{}, ok bool) {
	return elementOpenTag{
		Name:       parts[1].(string),
		Attributes: parts[2].([]Attribute),
	}, true
}

func (p elementOpenTagParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asElementOpenTag,
		parse.Rune('<'),
		elementNameParser,
		newAttributesParser().Parse,
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
var attributeNameFirst = "abcdefghijklmnopqrstuvwxyz"
var attributeNameSubsequent = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
var attributeNameParser = parse.Then(parse.WithStringConcatCombiner,
	parse.RuneIn(attributeNameFirst),
	parse.Many(parse.WithStringConcatCombiner, 0, 128, parse.RuneIn(attributeNameSubsequent)),
)

// Constant attribute.
var attributeConstantValueParser = parse.StringUntil(parse.Rune('"'))

func newConstantAttributeParser() constantAttributeParser {
	return constantAttributeParser{}
}

type constantAttributeParser struct {
}

func (p constantAttributeParser) asConstantAttribute(parts []interface{}) (result interface{}, ok bool) {
	return ConstantAttribute{
		Name:  parts[1].(string),
		Value: html.UnescapeString(parts[4].(string)),
	}, true
}

func (p constantAttributeParser) Parse(pi parse.Input) parse.Result {
	return parse.All(p.asConstantAttribute,
		whitespaceParser,
		attributeNameParser,
		parse.Rune('='),
		parse.Rune('"'),
		attributeConstantValueParser,
		parse.Rune('"'),
	)(pi)
}

// BoolConstantAttribute.
func newBoolConstantAttributeParser() boolConstantAttributeParser {
	return boolConstantAttributeParser{}
}

type boolConstantAttributeParser struct {
}

func (p boolConstantAttributeParser) Parse(pi parse.Input) parse.Result {
	var r BoolConstantAttribute

	start := pi.Index()
	pr := whitespaceParser(pi)
	if !pr.Success {
		return pr
	}

	pr = attributeNameParser(pi)
	if !pr.Success {
		err := rewind(pi, start)
		if err != nil {
			return parse.Failure("failed to rewind reader", err)
		}
		return pr
	}
	r.Name = pr.Item.(string)

	// We have a name, but if we have an equals sign, it's not a constant boolean attribute.
	next, err := pi.Peek()
	if err != nil {
		return parse.Failure("boolConstantAttributeParser", fmt.Errorf("boolConstantAttributeParser: unexpected error reading after attribute name: %w", pr.Error))
	}
	if next == '=' || next == '?' {
		// It's one of the other attribute types.
		err := rewind(pi, start)
		if err != nil && err != input.ErrStartOfFile {
			return parse.Failure("failed to rewind reader", err)
		}
		return parse.Failure("boolConstantAttributeParser", nil)
	}
	if !(next == ' ' || next == '\n' || next == '/') {
		return parse.Failure("boolConstantAttributeParser", fmt.Errorf("boolConstantAttributeParser: expected attribute name to end with space, newline or '/>', but got %q", string(next)))
	}

	return parse.Success("boolConstantAttributeParser", r, nil)
}

// BoolExpressionAttribute.
func newBoolExpressionAttributeParser() boolExpressionAttributeParser {
	return boolExpressionAttributeParser{}
}

var boolExpressionStart = parse.Any(parse.String("?={%= "), parse.String("?={%="))

type boolExpressionAttributeParser struct {
}

func (p boolExpressionAttributeParser) Parse(pi parse.Input) parse.Result {
	var r BoolExpressionAttribute

	start := pi.Index()
	pr := whitespaceParser(pi)
	if !pr.Success {
		return pr
	}

	pr = attributeNameParser(pi)
	if !pr.Success {
		err := rewind(pi, start)
		if err != nil {
			return parse.Failure("failed to rewind reader", err)
		}
		return pr
	}
	r.Name = pr.Item.(string)

	// Check whether this is a boolean expression attribute.
	if pr = boolExpressionStart(pi); !pr.Success {
		err := rewind(pi, start)
		if err != nil {
			return parse.Failure("failed to rewind reader", err)
		}
		return pr
	}

	// Once we've seen a expression prefix, read until the tag end.
	from := NewPositionFromInput(pi)
	pr = parse.StringUntil(expressionEnd)(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return parse.Failure("boolExpressionAttributeParser", fmt.Errorf("boolExpressionAttributeParser: failed to read until tag end: %w", pr.Error))
	}
	// If there's no tag end, the string expression parser wasn't terminated.
	if !pr.Success {
		return parse.Failure("boolExpressionAttributeParser", newParseError("bool expression attribute not terminated", from, NewPositionFromInput(pi)))
	}

	// Success! Create the expression.
	to := NewPositionFromInput(pi)
	r.Expression = NewExpression(pr.Item.(string), from, to)

	// Eat the tag end.
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("boolExpressionAttributeParser", newParseError("could not terminate boolean expression", from, NewPositionFromInput(pi)))
	}

	return parse.Success("boolExpressionAttributeParser", r, nil)
}

// ExpressionAttribute.
func newExpressionAttributeParser() expressionAttributeParser {
	return expressionAttributeParser{}
}

type expressionAttributeParser struct {
}

func (p expressionAttributeParser) Parse(pi parse.Input) parse.Result {
	var r ExpressionAttribute

	start := pi.Index()
	pr := whitespaceParser(pi)
	if !pr.Success {
		return pr
	}

	pr = attributeNameParser(pi)
	if !pr.Success {
		err := rewind(pi, start)
		if err != nil {
			return parse.Failure("failed to rewind reader", err)
		}
		return pr
	}
	r.Name = pr.Item.(string)

	if pr = parse.String("={%= ")(pi); !pr.Success {
		err := rewind(pi, start)
		if err != nil {
			return parse.Failure("failed to rewind reader", err)
		}
		return pr
	}

	// Once we've seen a expression prefix, read until the tag end.
	from := NewPositionFromInput(pi)
	pr = parse.StringUntil(expressionEnd)(pi)
	if pr.Error != nil && pr.Error != io.EOF {
		return parse.Failure("expressionAttributeParser", fmt.Errorf("expressionAttributeParser: failed to read until tag end: %w", pr.Error))
	}
	// If there's no tag end, the string expression parser wasn't terminated.
	if !pr.Success {
		return parse.Failure("expressionAttributeParser", newParseError("expression attribute not terminated", from, NewPositionFromInput(pi)))
	}

	// Success! Create the expression.
	to := NewPositionFromInput(pi)
	r.Expression = NewExpression(pr.Item.(string), from, to)

	// Eat the tag end.
	if te := expressionEnd(pi); !te.Success {
		return parse.Failure("expressionAttributeParser", newParseError("could not terminate string expression", from, NewPositionFromInput(pi)))
	}

	return parse.Success("expressionAttributeParser", r, nil)
}

func rewind(pi parse.Input, to int64) error {
	for i := pi.Index(); i > to; i-- {
		if _, err := pi.Retreat(); err != nil {
			return err
		}
	}
	return nil
}

// Attributes.
func newAttributesParser() attributesParser {
	return attributesParser{}
}

type attributesParser struct {
}

func (p attributesParser) asAttributeArray(parts []interface{}) (result interface{}, ok bool) {
	op := make([]Attribute, len(parts))
	for i := 0; i < len(parts); i++ {
		switch v := parts[i].(type) {
		case BoolConstantAttribute:
			op[i] = v
		case ConstantAttribute:
			op[i] = v
		case BoolExpressionAttribute:
			op[i] = v
		case ExpressionAttribute:
			op[i] = v
		}
	}
	return op, true
}

var attributeParser = parse.Any(
	newBoolConstantAttributeParser().Parse,
	newConstantAttributeParser().Parse,
	newBoolExpressionAttributeParser().Parse,
	newExpressionAttributeParser().Parse,
)

func (p attributesParser) Parse(pi parse.Input) parse.Result {
	return parse.Many(p.asAttributeArray, 0, 255, attributeParser)(pi)
}

// Element name.
var elementNameFirst = "abcdefghijklmnopqrstuvwxyz"
var elementNameSubsequent = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
var elementNameParser = parse.Then(parse.WithStringConcatCombiner,
	parse.RuneIn(elementNameFirst),
	parse.Many(parse.WithStringConcatCombiner, 0, 15, parse.RuneIn(elementNameSubsequent)),
)

// Element.
func newElementOpenCloseParser() elementOpenCloseParser {
	return elementOpenCloseParser{}
}

type elementOpenCloseParser struct {
	SourceRangeToItemLookup SourceMap
}

func (p elementOpenCloseParser) Parse(pi parse.Input) parse.Result {
	var r Element

	// Check the open tag.
	otr := newElementOpenTagParser().Parse(pi)
	if otr.Error != nil || !otr.Success {
		return otr
	}
	ot := otr.Item.(elementOpenTag)
	r.Name = ot.Name
	r.Attributes = ot.Attributes

	// Once we've got an open tag, the rest must be present.
	from := NewPositionFromInput(pi)
	tnpr := newTemplateNodeParser(nil).Parse(pi)
	if !tnpr.Success {
		if _, isParseError := tnpr.Error.(ParseError); isParseError {
			return tnpr
		}
		return parse.Failure("elementOpenCloseParser", newParseError(fmt.Sprintf("<%s>: %v", r.Name, tnpr.Error), from, NewPositionFromInput(pi)))
	}
	if arr, isArray := tnpr.Item.([]Node); isArray {
		r.Children = append(r.Children, arr...)
	}

	// Close tag.
	ectpr := elementCloseTagParser(pi)
	if !ectpr.Success {
		return parse.Failure("elementOpenCloseParser", newParseError(fmt.Sprintf("<%s>: expected end tag not present or invalid tag contents", r.Name), from, NewPositionFromInput(pi)))
	}
	if ct := ectpr.Item.(elementCloseTag); ct.Name != r.Name {
		return parse.Failure("elementOpenCloseParser", newParseError(fmt.Sprintf("<%s>: mismatched end tag, expected '</%s>', got '</%s>'", r.Name, r.Name, ct.Name), from, NewPositionFromInput(pi)))
	}

	return parse.Success("elementOpenCloseParser", r, nil)
}

// Element self-closing tag.
func newElementSelfClosingParser() elementSelfClosingParser {
	return elementSelfClosingParser{}
}

type elementSelfClosingParser struct {
	SourceRangeToItemLookup SourceMap
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
		newAttributesParser().Parse,
		optionalWhitespaceParser,
		parse.String("/>"),
	)(pi)
}

// Element
func newElementParser() elementParser {
	return elementParser{}
}

type elementParser struct {
}

func (p elementParser) Parse(pi parse.Input) parse.Result {
	var r Element

	// Self closing.
	from := NewPositionFromInput(pi)
	scr := newElementSelfClosingParser().Parse(pi)
	if scr.Error != nil && scr.Error != io.EOF {
		return scr
	}
	if scr.Success {
		r = scr.Item.(Element)
		if msgs, ok := r.Validate(); !ok {
			return parse.Failure("elementParser", newParseError(fmt.Sprintf("<%s>: %s", r.Name, strings.Join(msgs, ", ")), from, NewPositionFromInput(pi)))
		}
		return parse.Success("elementParser", r, nil)
	}

	// Open/close pair.
	ocr := newElementOpenCloseParser().Parse(pi)
	if ocr.Error != nil && ocr.Error != io.EOF {
		return ocr
	}
	if ocr.Success {
		r = ocr.Item.(Element)
		if msgs, ok := r.Validate(); !ok {
			return parse.Failure("elementParser", newParseError(fmt.Sprintf("<%s>: %s", r.Name, strings.Join(msgs, ", ")), from, NewPositionFromInput(pi)))
		}
		return parse.Success("elementParser", r, nil)
	}

	return parse.Failure("elementParser", nil)
}
