package parser

import (
	"fmt"
	"unicode"

	"github.com/a-h/parse"
)

// JSX-like component parser for templ
// Converts: <Component attr1="value1" attr2="value2" /> to @Component("value1", "value2")
// Converts: <Component attr1="value1">children</Component> to @Component("value1") { children }

// Component name validation - must be Component or package.Component format
var (
	componentNameParser = parse.Func(func(in *parse.Input) (name string, ok bool, err error) {
		start := in.Index()

		// Try to parse identifier (could be package name or component name)
		identifierFirst := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
		identifierSubsequent := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

		var prefix, suffix string
		if prefix, ok, err = parse.RuneIn(identifierFirst).Parse(in); err != nil || !ok {
			return
		}
		if suffix, ok, err = parse.StringUntil(parse.RuneNotIn(identifierSubsequent)).Parse(in); err != nil || !ok {
			in.Seek(start)
			return
		}

		firstIdentifier := prefix + suffix

		// Check if there's a dot (package.Component format)
		dotStart := in.Index()
		if _, dotOk, dotErr := parse.Rune('.').Parse(in); dotErr != nil || !dotOk {
			// No dot, this should be a Component name starting with uppercase
			if !unicode.IsUpper(rune(firstIdentifier[0])) {
				in.Seek(start)
				return "", false, nil
			}
			return firstIdentifier, true, nil
		}

		// Found a dot, parse the component name after it
		var componentPrefix, componentSuffix string
		if componentPrefix, ok, err = parse.RuneIn("ABCDEFGHIJKLMNOPQRSTUVWXYZ").Parse(in); err != nil || !ok {
			// Component name after dot must start with uppercase
			in.Seek(start)
			return "", false, nil
		}
		if componentSuffix, ok, err = parse.StringUntil(parse.RuneNotIn(identifierSubsequent)).Parse(in); err != nil || !ok {
			in.Seek(dotStart)
			return "", false, nil
		}

		fullName := firstIdentifier + "." + componentPrefix + componentSuffix
		if len(fullName) > 128 {
			ok = false
			err = parse.Error("component names must be < 128 characters long", in.Position())
			return
		}

		return fullName, true, nil
	})
)

// JSX component open tag
type jsxComponentOpenTag struct {
	Name        string
	Attributes  []Attribute
	SelfClosing bool
}

var jsxComponentOpenTagParser = parse.Func(func(pi *parse.Input) (e jsxComponentOpenTag, matched bool, err error) {
	start := pi.Index()

	if next, _ := pi.Peek(2); len(next) < 2 || next[0] != '<' || next == "<!" || next == "</" {
		// This is not a tag, or it's a comment, doctype, or closing tag.
		return e, false, nil
	}

	// <
	if _, matched, err = lt.Parse(pi); err != nil || !matched {
		return
	}

	// Component name - must start with uppercase letter
	if e.Name, matched, err = componentNameParser.Parse(pi); err != nil || !matched {
		pi.Seek(start)       // Restore parser state
		return e, false, nil // Not a component, let regular element parser handle it
	}

	// Parse attributes
	if e.Attributes, matched, err = (attributesParser{}).Parse(pi); err != nil || !matched {
		return e, true, err
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
		e.SelfClosing = true
		return e, true, nil
	}

	// >
	if _, matched, err = gt.Parse(pi); err != nil {
		return e, true, err
	}

	// If it's not a self-closing or complete open component, we have an error.
	if !matched {
		err = parse.Error(fmt.Sprintf("<%s>: malformed open component", e.Name), pi.Position())
		return
	}

	return e, true, nil
})

// JSX Component parser
var jsxComponent jsxComponentParser

type jsxComponentParser struct{}

func (jsxComponentParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	start := pi.Position()

	// Check the open tag.
	var ot jsxComponentOpenTag
	if ot, ok, err = jsxComponentOpenTagParser.Parse(pi); err != nil || !ok {
		return
	}

	// Create JSXComponentElement to preserve JSX syntax information
	l := pi.Position().Line
	r := &JSXComponentElement{
		Name:        ot.Name,
		NameRange:   NewRange(start, pi.Position()),
		Attributes:  ot.Attributes,
		SelfClosing: ot.SelfClosing,
	}

	// If any attribute is not on the same line as the component name, indent them.
	if pi.Position().Line != l {
		r.IndentAttrs = true
	}

	// If the component is self-closing, add trailing space and we're done
	if ot.SelfClosing {
		// Add trailing space.
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

	// Parse children for non-self-closing components
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
		err = parse.Error(fmt.Sprintf("<%s>: expected end tag not present or invalid tag contents", ot.Name), pi.Position())
		return r, true, err
	}

	// Add trailing space.
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
