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
	Name       string
	Attributes []Attribute
}

var elementOpenTagParser = parse.Func(func(pi *parse.Input) (e elementOpenTag, ok bool, err error) {
	start := pi.Index()

	// <
	if _, ok, err = lt.Parse(pi); err != nil || !ok {
		return
	}

	// Element name.
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

var elementCloseTagParser = parse.Func(func(in *parse.Input) (ect elementCloseTag, ok bool, err error) {
	var parts []string
	parts, ok, err = parse.All(
		parse.String("</"),
		elementNameParser,
		parse.Rune('>')).Parse(in)
	if err != nil || !ok {
		return
	}
	ect.Name = parts[1]
	return ect, true, nil
})

// Attribute name.
var (
	attributeNameFirst      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ:_@"
	attributeNameSubsequent = attributeNameFirst + "-.0123456789"
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
	attributeConstantValueParser = parse.StringUntil(parse.Rune('"'))
	constantAttributeParser      = parse.Func(func(pi *parse.Input) (attr ConstantAttribute, ok bool, err error) {
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
		if _, ok, err = parse.String(`="`).Parse(pi); err != nil || !ok {
			pi.Seek(start)
			return
		}

		// Attribute value.
		if attr.Value, ok, err = attributeConstantValueParser.Parse(pi); err != nil || !ok {
			pi.Seek(start)
			return
		}

		attr.Value = html.UnescapeString(attr.Value)

		// " - closing quote.
		if _, ok, err = Must(parse.String(`"`), fmt.Sprintf("missing closing quote on attribute %q", attr.Name)).Parse(pi); err != nil || !ok {
			pi.Seek(start)
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
	if !(next == " " || next == "\t" || next == "\n" || next == "/") {
		err = parse.Error(fmt.Sprintf("boolConstantAttributeParser: expected attribute name to end with space, newline or '/>', but got %q", next), pi.Position())
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
	if r.Expression, ok, err = Must[Expression](exp, "boolean expression: expected Go expression not found").Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Eat the Final brace.
	if _, ok, err = Must(closeBraceWithOptionalPadding, "boolean expression: missing closing brace").Parse(pi); err != nil || !ok {
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
	if _, ok, err = Must(closeBraceWithOptionalPadding, "boolean expression: missing closing brace").Parse(pi); err != nil || !ok {
		pi.Seek(start)
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
	if out, ok, err = conditionalAttributeParser.Parse(in); err != nil || ok {
		return
	}
	if out, ok, err = boolConstantAttributeParser.Parse(in); err != nil || ok {
		return
	}
	if out, ok, err = constantAttributeParser.Parse(in); err != nil || ok {
		return
	}
	return
}

// var attributesParser = parse.AtMost[Attribute](255, attributeParser{})
type attributesParser struct{}

func (attributesParser) Parse(in *parse.Input) (attributes []Attribute, ok bool, err error) {
	for {
		var attr Attribute
		attr, ok, err = attributeParser{}.Parse(in)
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
		if len(suffix)+1 > 16 {
			ok = false
			err = parse.Error("element property names must be < 16 characters long", in.Position())
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

	// Once we've got an open tag, the rest must be present.
	if r.Children, ok, err = newTemplateNodeParser[any](nil, "").Parse(pi); err != nil || !ok {
		return
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

	if _, ok, err = parse.String("/>").Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	return e, true, nil
})

// Element
var element elementParser

type elementParser struct{}

func (elementParser) Parse(pi *parse.Input) (r Element, ok bool, err error) {
	start := pi.Position()

	if r, ok, err = parse.Any[Element](selfClosingElement, elementOpenClose).Parse(pi); err != nil || !ok {
		return
	}
	var msgs []string
	if msgs, ok = r.Validate(); !ok {
		err = parse.Error(fmt.Sprintf("<%s>: %s", r.Name, strings.Join(msgs, ", ")), start)
	}

	return r, ok, err
}
