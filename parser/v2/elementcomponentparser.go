package parser

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/a-h/parse"
)

var componentNameParser = parse.Func(func(in *parse.Input) (name string, ok bool, err error) {
	start := in.Index()

	// Try to parse identifier (could be package name, variable name, or component name)
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

	fullName := prefix + suffix

	// Parse chained identifiers separated by dots
	for {
		dotStart := in.Index()
		if _, dotOk, dotErr := parse.Rune('.').Parse(in); dotErr != nil || !dotOk {
			break // No more dots
		}

		// Parse the next identifier after the dot
		var nextPrefix, nextSuffix string
		if nextPrefix, ok, err = parse.RuneIn(identifierFirst).Parse(in); err != nil || !ok {
			// If we can't parse an identifier after a dot, revert to before the dot
			in.Seek(dotStart)
			break
		}
		if nextSuffix, ok, err = parse.StringUntil(parse.RuneNotIn(identifierSubsequent)).Parse(in); err != nil || !ok {
			in.Seek(dotStart)
			break
		}

		fullName = fullName + "." + nextPrefix + nextSuffix
	}

	// Validate the component name format
	parts := strings.Split(fullName, ".")
	if len(parts) == 0 {
		in.Seek(start)
		return "", false, nil
	}

	// For a valid component name, the last part must start with uppercase
	lastPart := parts[len(parts)-1]
	if len(lastPart) == 0 || !unicode.IsUpper(rune(lastPart[0])) {
		in.Seek(start)
		return "", false, nil
	}

	// If there's only one part, it must be a direct Component (starts with uppercase)
	if len(parts) == 1 {
		if !unicode.IsUpper(rune(fullName[0])) {
			in.Seek(start)
			return "", false, nil
		}
	}

	if len(fullName) > 128 {
		ok = false
		err = parse.Error("component names must be < 128 characters long", in.Position())
		return
	}

	return fullName, true, nil
})

type elementComponentOpenTag struct {
	Name        string
	NameRange   Range
	Attributes  []Attribute
	IndentAttrs bool
	SelfClosing bool
}

var elementComponentOpenTagParser = parse.Func(func(pi *parse.Input) (e elementComponentOpenTag, matched bool, err error) {
	start := pi.Index()
	l := pi.Position().Line

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
	e.NameRange = NewRange(pi.PositionAt(pi.Index()-len(e.Name)), pi.Position())

	// Parse attributes
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

// Element Component parser
var elementComponent elementComponentParser

type elementComponentParser struct{}

func (elementComponentParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	start := pi.Position()

	// Check the open tag.
	var ot elementComponentOpenTag
	if ot, ok, err = elementComponentOpenTagParser.Parse(pi); err != nil || !ok {
		return
	}

	l := pi.Position().Line
	r := &ElementComponent{
		Name:        ot.Name,
		NameRange:   ot.NameRange,
		Attributes:  ot.Attributes,
		IndentAttrs: ot.IndentAttrs,
		SelfClosing: ot.SelfClosing,
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
		r.Range = NewRange(start, pi.Position())
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
		r.Range = NewRange(start, pi.Position())
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
		r.Range = NewRange(start, pi.Position())
		return r, true, err
	}
	if !ok {
		err = parse.Error(fmt.Sprintf("<%s>: expected end tag not present or invalid tag contents", ot.Name), pi.Position())
		r.Range = NewRange(start, pi.Position())
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
	r.Range = NewRange(start, pi.Position())

	return r, true, nil
}
