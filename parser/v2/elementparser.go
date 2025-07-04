package parser

import (
	"fmt"
	"strings"

	"github.com/a-h/parse"
	"github.com/a-h/templ/parser/v2/goexpression"
)

// Element.

// Element open tag.
type elementOpenTag struct {
	Name        string
	Attributes  []Attribute
	IndentAttrs bool
	NameRange   Range
	Void        bool
}

var elementOpenTagParser = parse.Func(func(pi *parse.Input) (e elementOpenTag, matched bool, err error) {
	if next, _ := pi.Peek(2); len(next) < 2 || next[0] != '<' || next == "<!" || next == "</" {
		// This is not a tag, or it's a comment, doctype, or closing tag.
		return e, false, nil
	}

	// <
	if _, matched, err = lt.Parse(pi); err != nil || !matched {
		return
	}

	// Element name.
	l := pi.Position().Line
	if e.Name, matched, err = elementNameParser.Parse(pi); err != nil || !matched {
		return e, true, err
	}
	e.NameRange = NewRange(pi.PositionAt(pi.Index()-len(e.Name)), pi.Position())

	if e.Attributes, matched, err = (attributesParser{}).Parse(pi); err != nil || !matched {
		return e, true, err
	}

	// If any attribute is not on the same line as the element name, indent them.
	if pi.Position().Line != l {
		e.IndentAttrs = true
	}

	// Optional whitespace.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		return e, true, err
	}

	// />
	if _, matched, err = parse.String("/>").Parse(pi); err != nil {
		return e, true, err
	}
	if matched {
		e.Void = true
		return e, true, nil
	}

	// >
	if _, matched, err = gt.Parse(pi); err != nil {
		return e, true, err
	}

	// If it's not a self-closing or complete open element, we have an error.
	if !matched {
		err = parse.Error(fmt.Sprintf("<%s>: malformed open element", e.Name), pi.Position())
		return
	}

	return e, true, nil
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

type attributeValueParser struct {
	EqualsAndQuote parse.Parser[string]
	Suffix         parse.Parser[string]
	UseSingleQuote bool
}

func (avp attributeValueParser) Parse(pi *parse.Input) (value string, ok bool, err error) {
	start := pi.Index()
	if _, ok, err = avp.EqualsAndQuote.Parse(pi); err != nil || !ok {
		return
	}
	if value, ok, err = parse.StringUntil(avp.Suffix).Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}
	if _, ok, err = avp.Suffix.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}
	return value, true, nil
}

// Constant attribute.
var (
	attributeValueParsers = []attributeValueParser{
		// Double quoted.
		{EqualsAndQuote: parse.StringFrom(parse.OptionalWhitespace, parse.String(`="`)), Suffix: parse.String(`"`), UseSingleQuote: false},
		// Single quoted.
		{EqualsAndQuote: parse.StringFrom(parse.OptionalWhitespace, parse.String(`='`)), Suffix: parse.String(`'`), UseSingleQuote: true},
		// Unquoted.
		// A valid unquoted attribute value in HTML is any string of text that is not an empty string,
		// and that doesn’t contain spaces, tabs, line feeds, form feeds, carriage returns, ", ', `, =, <, or >.
		{EqualsAndQuote: parse.StringFrom(parse.OptionalWhitespace, parse.String("=")), Suffix: parse.Any(parse.RuneIn(" \t\n\r\"'`=<>/"), parse.EOF[string]()), UseSingleQuote: false},
	}
	constantAttributeParser = parse.Func(func(pi *parse.Input) (attr *ConstantAttribute, ok bool, err error) {
		start := pi.Index()

		// Optional whitespace leader.
		if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
			return
		}

		attr = &ConstantAttribute{}

		// Attribute name.
		if attr.Key, ok, err = attributeKeyParser.Parse(pi); err != nil || !ok {
			pi.Seek(start)
			return
		}

		for _, p := range attributeValueParsers {
			attr.Value, ok, err = p.Parse(pi)
			if err != nil {
				pos := pi.Position()
				if pErr, isParseError := err.(parse.ParseError); isParseError {
					pos = pErr.Pos
				}
				return attr, false, parse.Error(fmt.Sprintf("%s: %v", attr.Key, err), pos)
			}
			if ok {
				attr.SingleQuote = p.UseSingleQuote
				break
			}
		}

		if !ok {
			pi.Seek(start)
			return attr, false, nil
		}

		// Only use single quotes if actually required, due to double quote in the value (prefer double quotes).
		attr.SingleQuote = attr.SingleQuote && strings.Contains(attr.Value, "\"")

		return attr, true, nil
	})
)

var expressionAttributeKeyParser = parse.Func(func(pi *parse.Input) (attr AttributeKey, ok bool, err error) {
	start := pi.Index()
	// Eat the first brace.
	if _, ok, err = openBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}
	var ek ExpressionAttributeKey

	// Expression.
	if ek.Expression, err = parseGoSliceArgs(pi); err != nil {
		return attr, false, err
	}

	// Eat the Final brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("attribute key: missing closing brace", pi.Position())
		pi.Seek(start)
		return
	}
	return ek, true, nil
})

var constantAttributeKeyParser = parse.Func(func(pi *parse.Input) (k AttributeKey, ok bool, err error) {
	start := pi.Index()
	n, ok, err := attributeNameParser.Parse(pi)
	if err != nil || !ok {
		pi.Seek(start)
		return
	}
	r := NewRange(pi.PositionAt(start), pi.Position())
	k = ConstantAttributeKey{Name: n, NameRange: r}
	return k, true, nil
})

var attributeKeyParser = parse.Any(constantAttributeKeyParser, expressionAttributeKeyParser)

// BoolConstantAttribute.
var boolConstantAttributeParser = parse.Func(func(pi *parse.Input) (attr *BoolConstantAttribute, ok bool, err error) {
	start := pi.Index()

	// Optional whitespace leader.
	if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}

	attr = &BoolConstantAttribute{}

	// Attribute name.
	if attr.Key, ok, err = attributeKeyParser.Parse(pi); err != nil || !ok {
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
	if next != " " && next != "\t" && next != "\r" && next != "\n" && next != "/" && next != ">" {
		err = parse.Error(fmt.Sprintf("boolConstantAttributeParser: expected attribute name to end with space, newline, '/>' or '>', but got %q", next), pi.Position())
		return attr, false, err
	}

	return attr, true, nil
})

// BoolExpressionAttribute.
var boolExpressionStart = parse.Or(parse.String("?={ "), parse.String("?={"))

var boolExpressionAttributeParser = parse.Func(func(pi *parse.Input) (r *BoolExpressionAttribute, ok bool, err error) {
	start := pi.Index()

	// Optional whitespace leader.
	if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	r = &BoolExpressionAttribute{}

	// Attribute name.
	if r.Key, ok, err = attributeKeyParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Check whether this is a boolean expression attribute.
	if _, ok, err = boolExpressionStart.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Once we have a prefix, we must have an expression that returns a boolean.
	if r.Expression, err = parseGo("boolean attribute", pi, goexpression.Expression); err != nil {
		return r, false, err
	}

	// Eat the Final brace.
	if _, ok, err = closeBraceWithOptionalPadding.Parse(pi); err != nil || !ok {
		err = parse.Error("boolean expression: missing closing brace", pi.Position())
		pi.Seek(start)
		return
	}

	return r, true, nil
})

var expressionAttributeStartParser = parse.StringFrom(parse.OptionalWhitespace, parse.String("="), parse.OptionalWhitespace, parse.String("{"), parse.OptionalWhitespace)

var expressionAttributeParser = parse.Func(func(pi *parse.Input) (attr *ExpressionAttribute, ok bool, err error) {
	start := pi.Index()

	// Optional whitespace leader.
	if _, ok, err = parse.OptionalWhitespace.Parse(pi); err != nil || !ok {
		return
	}

	attr = &ExpressionAttribute{}

	// Attribute name.
	if attr.Key, ok, err = attributeKeyParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// ={
	if _, ok, err = expressionAttributeStartParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Expression.
	if attr.Expression, err = parseGoSliceArgs(pi); err != nil {
		return attr, false, err
	}

	// Eat whitespace, plus the final brace.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		return attr, false, err
	}
	if _, ok, err = closeBrace.Parse(pi); err != nil || !ok {
		err = parse.Error("string expression attribute: missing closing brace", pi.Position())
		return
	}

	return attr, true, nil
})

var spreadAttributesParser = parse.Func(func(pi *parse.Input) (attr *SpreadAttributes, ok bool, err error) {
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
	attr = &SpreadAttributes{}
	if attr.Expression, err = parseGo("spread attributes", pi, goexpression.Expression); err != nil {
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
	if out, ok, err = spreadAttributesParser.Parse(in); err != nil || ok {
		return
	}
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
	elementNameSubsequent = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-:"
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

// Void element closer.
var voidElementCloser voidElementCloserParser

type voidElementCloserParser struct{}

var voidElementCloseTags = []string{"</area>", "</base>", "</br>", "</col>", "</command>", "</embed>", "</hr>", "</img>", "</input>", "</keygen>", "</link>", "</meta>", "</param>", "</source>", "</track>", "</wbr>"}

func (voidElementCloserParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	var ve string
	for _, ve = range voidElementCloseTags {
		s, canPeekLen := pi.Peek(len(ve))
		if !canPeekLen {
			continue
		}
		if !strings.EqualFold(s, ve) {
			continue
		}
		// Found a match.
		ok = true
		break
	}
	if !ok {
		return nil, false, nil
	}
	pi.Take(len(ve))
	return nil, true, nil
}

// Element.
var element elementParser

type elementParser struct{}

func (elementParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	start := pi.Position()

	if prefix, _ := pi.Peek(len("<script")); prefix == "<script" {
		// This is a script element, which has special handling and should
		// be parsed by a different parser to this one.
		return nil, false, nil
	}

	// Check the open tag.
	var ot elementOpenTag
	if ot, ok, err = elementOpenTagParser.Parse(pi); err != nil || !ok {
		return
	}
	r := &Element{
		Name:        ot.Name,
		Attributes:  ot.Attributes,
		IndentAttrs: ot.IndentAttrs,
		NameRange:   ot.NameRange,
	}

	// Once we've got an open tag, the rest must be present.
	l := pi.Position().Line

	// If the element is self-closing, even if it's not really a void element (br, hr etc.), we can return early.
	if ot.Void || r.IsVoidElement() {
		// Escape early, no need to try to parse children for self-closing elements.
		return addTrailingSpaceAndValidate(start, r, pi)
	}

	// Parse children.
	closer := StripType(parse.All(parse.String("</"), parse.String(ot.Name), parse.Rune('>')))
	tnp := newTemplateNodeParser(closer, fmt.Sprintf("<%s>: close tag", ot.Name))
	nodes, _, err := tnp.Parse(pi)
	if err != nil {
		notFoundErr, isNotFoundError := err.(UntilNotFoundError)
		if isNotFoundError {
			err = notFoundErr.ParseError
		}
		// If we got any nodes, take them, because the LSP might want to use them.
		r.Children = nodes.Nodes
		return r, true, err
	}
	r.Children = nodes.Nodes
	// If the children are not all on the same line, indent them.
	if l != pi.Position().Line {
		r.IndentChildren = true
	}

	// Close tag.
	_, ok, err = closer.Parse(pi)
	if err != nil {
		return r, true, err
	}
	if !ok {
		err = parse.Error(fmt.Sprintf("<%s>: expected end tag not present or invalid tag contents", r.Name), pi.Position())
		return r, true, err
	}

	return addTrailingSpaceAndValidate(start, r, pi)
}

func addTrailingSpaceAndValidate(start parse.Position, e *Element, pi *parse.Input) (n Node, ok bool, err error) {
	// Elide any void close tags.
	if _, _, err = voidElementCloser.Parse(pi); err != nil {
		return e, false, err
	}
	// Add trailing space.
	ws, _, err := parse.Whitespace.Parse(pi)
	if err != nil {
		return e, false, err
	}
	e.TrailingSpace, err = NewTrailingSpace(ws)
	if err != nil {
		return e, false, err
	}

	// Validate.
	var msgs []string
	if msgs, ok = e.Validate(); !ok {
		err = parse.Error(fmt.Sprintf("<%s>: %s", e.Name, strings.Join(msgs, ", ")), start)
		return e, false, err
	}

	return e, true, nil
}
