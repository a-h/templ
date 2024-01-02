package parser

import (
	"fmt"
	"html"
	"strings"

	"github.com/a-h/parse"
)

// Element.

// Element open tag.
type elementOpenTag struct {
	Name        string
	Attributes  []Attribute
	IndentAttrs bool
}

var elementOpenTagParser = parse.Func(func(pi *parse.Input) (e elementOpenTag, ok bool, err error) {
	start := pi.Index()

	// <
	if _, ok, err = lt.Parse(pi); err != nil || !ok {
		return
	}

	// Element name.
	l := pi.Position().Line
	if e.Name, ok, err = elementNameParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	if e.Attributes, ok, err = (attributesParser{}).Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// If any attribute is not on the same line as the element name, indent them.
	if pi.Position().Line != l {
		e.IndentAttrs = true
	}

	// Optional whitespace.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		pi.Seek(start)
		return
	}

	// >
	if _, ok, err = gt.Parse(pi); err != nil {
		return
	}
	if !ok {
		err = parse.Error(fmt.Sprintf("<%s>: malformed open element", e.Name), pi.Position())
		return e, false, err
	}

	return e, true, nil
})

// Element close tag.
type elementCloseTag struct {
	Name string
}

var elementCloseTagParser = parse.Func(func(in *parse.Input) (ct elementCloseTag, ok bool, err error) {
	var parts []string
	parts, ok, err = parse.All(
		parse.String("</"),
		elementNameParser,
		parse.Rune('>')).Parse(in)
	if err != nil || !ok {
		return
	}
	ct.Name = parts[1]
	return ct, true, nil
})

// Attribute name.
var (
	attributeNameFirst      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ:_@"
	attributeNameSubsequent = attributeNameFirst + "-.0123456789*"
	attributeNameParser     = parse.Func(func(in *parse.Input) (name string, ok bool, err error) {
		start := in.Index()
		var prefix, suffix string
		if prefix, ok, err = parse.RuneIn(attributeNameFirst).Parse(in); err != nil || !ok {
			return
		}
		if suffix, ok, err = parse.StringUntil(parse.RuneNotIn(attributeNameSubsequent)).Parse(in); err != nil {
			in.Seek(start)
			return
		}
		if len(suffix)+1 > 128 {
			ok = false
			err = parse.Error("attribute names must be < 128 characters long", in.Position())
			return
		}
		return prefix + suffix, true, nil
	})
)

// Constant attribute.
var (
	attributeConstantValueParser            = parse.StringUntil(parse.Rune('"'))
	attributeConstantValueSingleQuoteParser = parse.StringUntil(parse.Rune('\''))
	constantAttributeParser                 = parse.Func(func(pi *parse.Input) (attr ConstantAttribute, ok bool, err error) {
		start := pi.Index()

		// Optional whitespace leader.
		if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
			return
		}

		// Attribute name.
		if attr.Name, ok, err = attributeNameParser.Parse(pi); err != nil || !ok {
			pi.Seek(start)
			return
		}

		// ="
		result, ok, err := parse.Or(parse.String(`="`), parse.String(`='`)).Parse(pi)
		if err != nil || !ok {
			pi.Seek(start)
			return
		}

		valueParser := attributeConstantValueParser
		closeParser := parse.String(`"`)
		if result.B.OK {
			valueParser = attributeConstantValueSingleQuoteParser
			closeParser = parse.String(`'`)
			attr.SingleQuote = true
		}

		// Attribute value.
		if attr.Value, ok, err = valueParser.Parse(pi); err != nil || !ok {
			pi.Seek(start)
			return
		}

		attr.Value = html.UnescapeString(attr.Value)
		// Only use single quotes if actually required, due to double quote in the value (prefer double quotes).
		if attr.SingleQuote && !strings.Contains(attr.Value, "\"") {
			attr.SingleQuote = false
		}

		// " - closing quote.
		if _, ok, err = closeParser.Parse(pi); err != nil || !ok {
			err = parse.Error(fmt.Sprintf("missing closing quote on attribute %q", attr.Name), pi.Position())
			return
		}

		return attr, true, nil
	})
)

// BoolConstantAttribute.
var boolConstantAttributeParser = parse.Func(func(pi *parse.Input) (attr BoolConstantAttribute, ok bool, err error) {
	start := pi.Index()

	// Optional whitespace leader.
	if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}

	// Attribute name.
	if attr.Name, ok, err = attributeNameParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// We have a name, but if we have an equals sign, it's not a constant boolean attribute.
	next, ok := pi.Peek(1)
	if !ok {
		err = parse.Error("boolConstantAttributeParser: unexpected EOF after attribute name", pi.Position())
		return
	}
	if next == "=" || next == "?" {
		// It's one of the other attribute types.
		pi.Seek(start)
		return attr, false, nil
	}
	if !(next == " " || next == "\t" || next == "\r" || next == "\n" || next == "/" || next == ">") {
		err = parse.Error(fmt.Sprintf("boolConstantAttributeParser: expected attribute name to end with space, newline, '/>' or '>', but got %q", next), pi.Position())
		return attr, false, err
	}

	return attr, true, nil
})

// BoolExpressionAttribute.
var boolExpressionStart = parse.Or(parse.String("?={ "), parse.String("?={"))

var boolExpressionAttributeParser = parse.Func(func(pi *parse.Input) (r BoolExpressionAttribute, ok bool, err error) {
	start := pi.Index()

	// Optional whitespace leader.
	if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Attribute name.
	if r.Name, ok, err = attributeNameParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Check whether this is a boolean expression attribute.
	if _, ok, err = boolExpressionStart.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Once we have a prefix, we must have an expression that returns a template.
	if r.Expression, ok, err = exp.Parse(pi); err != nil || !ok {
		err = parse.Error("boolean expression: expected Go expression not found", pi.Position())
		return
	}

	// Eat the Final brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("boolean expression: missing closing brace", pi.Position())
		pi.Seek(start)
		return
	}

	return r, true, nil
})

var expressionAttributeParser = parse.Func(func(pi *parse.Input) (attr ExpressionAttribute, ok bool, err error) {
	start := pi.Index()

	// Optional whitespace leader.
	if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}

	// Attribute name.
	if attr.Name, ok, err = attributeNameParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// ={
	if _, ok, err = parse.Or(parse.String("={ "), parse.String("={")).Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Expression.
	if attr.Expression, ok, err = exp.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Eat the final brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("boolean expression: missing closing brace", pi.Position())
		return
	}

	return attr, true, nil
})

var spreadAttributesParser = parse.Func(func(pi *parse.Input) (attr SpreadAttributes, ok bool, err error) {
	start := pi.Index()

	// Optional whitespace leader.
	if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}

	// Eat the first brace.
	if _, ok, err = openBraceWithOptionalPadding.Parse(pi); err != nil ||
		!ok {
		pi.Seek(start)
		return
	}

	// Expression.
	if attr.Expression, ok, err = exp.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Check if end of expression has "..." for spread.
	if !strings.HasSuffix(attr.Expression.Value, "...") {
		pi.Seek(start)
		ok = false
		return
	}

	// Remove extra spread characters from expression.
	attr.Expression.Value = strings.TrimSuffix(attr.Expression.Value, "...")
	attr.Expression.Range.To.Col -= 3
	attr.Expression.Range.To.Index -= 3

	// Eat the final brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("attribute spread expression: missing closing brace", pi.Position())
		return
	}

	return attr, true, nil
})

// Attributes.
type attributeParser struct{}

func (attributeParser) Parse(in *parse.Input) (out Attribute, ok bool, err error) {
	if out, ok, err = boolExpressionAttributeParser.Parse(in); err != nil || ok {
		return
	}
	if out, ok, err = expressionAttributeParser.Parse(in); err != nil || ok {
		return
	}
	if out, ok, err = conditionalAttribute.Parse(in); err != nil || ok {
		return
	}
	if out, ok, err = boolConstantAttributeParser.Parse(in); err != nil || ok {
		return
	}
	if out, ok, err = spreadAttributesParser.Parse(in); err != nil || ok {
		return
	}
	if out, ok, err = constantAttributeParser.Parse(in); err != nil || ok {
		return
	}
	return
}

var attribute attributeParser

type attributesParser struct{}

func (attributesParser) Parse(in *parse.Input) (attributes []Attribute, ok bool, err error) {
	for {
		var attr Attribute
		attr, ok, err = attribute.Parse(in)
		if err != nil {
			return
		}
		if !ok {
			break
		}
		attributes = append(attributes, attr)
	}
	return attributes, true, nil
}

// Element name.
var (
	elementNameFirst      = "abcdefghijklmnopqrstuvwxyz"
	elementNameSubsequent = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
	elementNameParser     = parse.Func(func(in *parse.Input) (name string, ok bool, err error) {
		start := in.Index()
		var prefix, suffix string
		if prefix, ok, err = parse.RuneIn(elementNameFirst).Parse(in); err != nil || !ok {
			return
		}
		if suffix, ok, err = parse.StringUntil(parse.RuneNotIn(elementNameSubsequent)).Parse(in); err != nil || !ok {
			in.Seek(start)
			return
		}
		if len(suffix)+1 > 128 {
			ok = false
			err = parse.Error("element names must be < 128 characters long", in.Position())
			return
		}
		return prefix + suffix, true, nil
	})
)

// Element.
var elementOpenClose elementOpenCloseParser

type elementOpenCloseParser struct{}

func (elementOpenCloseParser) Parse(pi *parse.Input) (r Element, ok bool, err error) {
	// Check the open tag.
	var ot elementOpenTag
	if ot, ok, err = elementOpenTagParser.Parse(pi); err != nil || !ok {
		return
	}
	r.Name = ot.Name
	r.Attributes = ot.Attributes
	r.IndentAttrs = ot.IndentAttrs

	// Once we've got an open tag, the rest must be present.
	l := pi.Position().Line
	var nodes Nodes
	if nodes, ok, err = newTemplateNodeParser[any](nil, "").Parse(pi); err != nil || !ok {
		return
	}
	r.Children = nodes.Nodes
	r.Diagnostics = nodes.Diagnostics
	// If the children are not all on the same line, indent them
	if l != pi.Position().Line {
		r.IndentChildren = true
	}

	// Close tag.
	pos := pi.Position()
	var ct elementCloseTag
	ct, ok, err = elementCloseTagParser.Parse(pi)
	if err != nil {
		return
	}
	if !ok {
		err = parse.Error(fmt.Sprintf("<%s>: expected end tag not present or invalid tag contents", r.Name), pi.Position())
		return
	}
	if ct.Name != r.Name {
		err = parse.Error(fmt.Sprintf("<%s>: mismatched end tag, expected '</%s>', got '</%s>'", r.Name, r.Name, ct.Name), pos)
		return
	}

	// Parse trailing whitespace.
	ws, _, err := parse.Whitespace.Parse(pi)
	if err != nil {
		return r, false, err
	}
	r.TrailingSpace, err = NewTrailingSpace(ws)
	if err != nil {
		return r, false, err
	}

	return r, true, nil
}

// Element self-closing tag.
var selfClosingElement = parse.Func(func(pi *parse.Input) (e Element, ok bool, err error) {
	start := pi.Index()

	// lt
	if _, ok, err = lt.Parse(pi); err != nil || !ok {
		return
	}

	// Element name.
	l := pi.Position().Line
	if e.Name, ok, err = elementNameParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	if e.Attributes, ok, err = (attributesParser{}).Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Optional whitespace.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		pi.Seek(start)
		return
	}

	// If any attribute is not on the same line as the element name, indent them.
	if pi.Position().Line != l {
		e.IndentAttrs = true
	}

	if _, ok, err = parse.String("/>").Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Parse trailing whitespace.
	ws, _, err := parse.Whitespace.Parse(pi)
	if err != nil {
		return e, false, err
	}
	e.TrailingSpace, err = NewTrailingSpace(ws)
	if err != nil {
		return e, false, err
	}

	return e, true, nil
})

// Element
var element elementParser

type elementParser struct{}

func (elementParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	start := pi.Position()

	var r Element
	if r, ok, err = parse.Any[Element](selfClosingElement, elementOpenClose).Parse(pi); err != nil || !ok {
		return
	}
	var msgs []string
	if msgs, ok = r.Validate(); !ok {
		err = parse.Error(fmt.Sprintf("<%s>: %s", r.Name, strings.Join(msgs, ", ")), start)
	}

	return r, ok, err
}
