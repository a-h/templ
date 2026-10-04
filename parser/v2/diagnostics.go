package parser

import (
	"errors"
	"fmt"
	"iter"
	"strings"
	"unicode"
	"unicode/utf8"
)

type diagnoser func(Node) ([]Diagnostic, error)

// Diagnostic for template file.
type Diagnostic struct {
	Message string
	Range   Range
}

func walkTemplate(t *TemplateFile, f func(Node) bool) {
	for _, n := range t.Nodes {
		hn, ok := n.(*HTMLTemplate)
		if !ok {
			continue
		}
		walkNodes(hn.Children, f)
	}
}
func walkNodes(t []Node, f func(Node) bool) {
	for _, n := range t {
		if !f(n) {
			continue
		}
		if h, ok := n.(CompositeNode); ok {
			walkNodes(h.ChildNodes(), f)
		}
	}
}

var diagnosers = []diagnoser{
	useOfLegacyCallSyntaxDiagnoser,
	textCallDiagnoser,
	futureTextCallDiagnoser,
}

func Diagnose(t *TemplateFile) ([]Diagnostic, error) {
	var diags []Diagnostic
	var errs error
	walkTemplate(t, func(n Node) bool {
		for _, d := range diagnosers {
			diag, err := d(n)
			if err != nil {
				errs = errors.Join(errs, err)
				return false
			}
			diags = append(diags, diag...)
		}
		return true
	})
	return diags, errs
}

func useOfLegacyCallSyntaxDiagnoser(n Node) ([]Diagnostic, error) {
	if c, ok := n.(*CallTemplateExpression); ok {
		return []Diagnostic{{
			Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
			Range:   c.Expression.Range,
		}}, nil
	}
	return nil, nil
}

// textCallDiagnoser warns when text contains what looks like a component call,
// such as `(@component())`, that is rendered as text because `@` does not follow
// whitespace.
func textCallDiagnoser(n Node) (diags []Diagnostic, err error) {
	t, ok := n.(*Text)
	if !ok {
		return nil, nil
	}
	for index, name := range findNonWordPrefixedAtNames(t.Value) {
		if !strings.HasPrefix(t.Value[index+len(name):], "(") {
			continue
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("`%s(` is rendered as text, not as a component call, because `@` does not follow whitespace. Place whitespace before `@` to call the component.", name),
			Range:   newTextSubRange(t, index, len(name)),
		})
	}
	return diags, nil
}

// futureTextCallDiagnoser warns about text that a future templ release will
// parse as a component call. That release adopts the Razor rule: `@` followed by
// an identifier starts a call unless it follows a letter, digit, or underscore,
// as in an email address.
func futureTextCallDiagnoser(n Node) (diags []Diagnostic, err error) {
	t, ok := n.(*Text)
	if !ok {
		return nil, nil
	}
	for index, name := range findNonWordPrefixedAtNames(t.Value) {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("A future templ release will parse `%s` as a component call, because `@` that does not follow a letter, digit, or underscore will start a component call. Use `{ \"@\" }` to render a literal `@`.", name),
			Range:   newTextSubRange(t, index, len(name)),
		})
	}
	return diags, nil
}

// findNonWordPrefixedAtNames yields the byte index of each `@` in s that is
// followed by an identifier and does not follow a letter, digit, or underscore,
// together with the `@` and the identifier, e.g. `@pkg.Component`.
func findNonWordPrefixedAtNames(s string) iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for index, r := range s {
			if r != '@' {
				continue
			}
			previous, _ := utf8.DecodeLastRuneInString(s[:index])
			if isWordRune(previous) {
				continue
			}
			name := readQualifiedIdentifier(s[index+1:])
			if name == "" {
				continue
			}
			if !yield(index, "@"+name) {
				return
			}
		}
	}
}

// readQualifiedIdentifier returns the Go identifier at the start of s, including
// any dot-separated qualifiers, e.g. `pkg.Component`.
func readQualifiedIdentifier(s string) (name string) {
	for {
		ident := readIdentifier(s[len(name):])
		if ident == "" {
			return strings.TrimSuffix(name, ".")
		}
		name += ident
		if !strings.HasPrefix(s[len(name):], ".") {
			return name
		}
		name += "."
	}
}

// readIdentifier returns the Go identifier at the start of s.
func readIdentifier(s string) string {
	for index, r := range s {
		if !isWordRune(r) || (index == 0 && unicode.IsDigit(r)) {
			return s[:index]
		}
	}
	return s
}

func isWordRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// newTextSubRange returns the range of length bytes starting at byte index
// within the text. Text nodes do not span lines, so the column advances with
// the index.
func newTextSubRange(t *Text, index, length int) Range {
	from := Position{
		Index: t.Range.From.Index + int64(index),
		Line:  t.Range.From.Line,
		Col:   t.Range.From.Col + uint32(index),
	}
	return Range{
		From: from,
		To: Position{
			Index: from.Index + int64(length),
			Line:  from.Line,
			Col:   from.Col + uint32(length),
		},
	}
}
