// Package jstokenizer reads JavaScript source one token at a time, and tracks the
// lexical state needed to decide how a value inserted between tokens must be escaped.
//
// The templ parser uses it to classify the position of each {{ }} Go expression in a
// <script> element. The source is not valid JavaScript while it contains {{ }}
// expressions, so a JavaScript parser cannot be used. A tokenizer is enough, because
// the escaping only depends on whether a value is in code, a string, or a template
// literal.
//
// See GHSA-jq3r-rjqp-mwgj, where Go expressions in template literals were not escaped
// for template literal context, so values such as ${alert(1)} could execute.
package jstokenizer

import (
	"errors"
	"slices"

	"github.com/a-h/parse"
)

// Context describes the position that a value inserted into JavaScript source is
// written into, which determines how the value must be escaped.
//
// name = "${evil()}", v = "alert(1)". Both inputs look like they could run code.
// ContextTemplateLiteral keeps them inert by escaping $, { and }; ContextCode keeps
// them inert by JSON-quoting them instead:
//
// Template                Context                  Rendered                        Evaluates to        Consequence of alert(message)
// `Hi {{ name }}`         ContextTemplateLiteral   `Hi \u0024\u007bevil()\u007d`   "Hi ${evil()}"      shows "Hi ${evil()}" - evil() never runs
// `Total: ${ {{ v }} }`   ContextCode              `Total: ${ "alert(1)" }`        "Total: alert(1)"   shows "Total: alert(1)" - alert(1) never runs
type Context int

const (
	// ContextCode values are written into JavaScript code, including the code within a
	// template literal's ${ } substitution.
	ContextCode Context = iota
	// ContextString values are written into a single or double quoted string literal.
	ContextString
	// ContextTemplateLiteral values are written into the text of a backtick quoted
	// template literal.
	ContextTemplateLiteral
	// ContextRegexp values are written into the pattern of a regular expression
	// literal.
	ContextRegexp
	// ContextRegexpStart values are written at the start of the pattern of a regular
	// expression literal. An empty value there would join the opening '/' to the next
	// character, so that `/{{ x }}/` becomes the comment `//`.
	ContextRegexpStart
)

// ErrAmbiguousContext is returned for a value whose context depends on whether a '/'
// before it is a division or the start of a regular expression, which the tokenizer
// can't determine.
var ErrAmbiguousContext = errors.New("the context of the value depends on whether a '/' before it is a division or the start of a regular expression, which can't be determined")

// maxAlternatives limits the number of interpretations of the source that are read at
// the same time. Each ambiguous '/' doubles the number of interpretations until they
// converge.
const maxAlternatives = 16

// Tokenizer reads JavaScript source one token at a time, and tracks the lexical state
// of the source read so far.
//
// Whether a '/' in code is a division or the start of a regular expression depends on
// the grammar, which the tokenizer doesn't parse. When it can't be determined, the
// tokenizer reads the source both ways, until both interpretations reach the same
// state. A value inserted while the interpretations differ about its context is an
// error, see ADR 0009.
type Tokenizer struct {
	// lexer reads the source at the position of the input passed to the Tokenizer.
	lexer *lexer
	// alternatives are the other interpretations of the source. Each reads the source
	// in its own tokens, so it can be at a different position to lexer.
	alternatives []alternative
	// hasTooManyAlternatives is true once the number of interpretations exceeded
	// maxAlternatives, after which the context of every value is ambiguous.
	hasTooManyAlternatives bool
}

// alternative is an interpretation of the source, and the index of the input that it
// has read up to.
type alternative struct {
	lexer *lexer
	index int
}

// IsEqual returns true if the alternatives have read the source up to the same index, and
// have the same state.
func (a alternative) IsEqual(other alternative) bool {
	return a.index == other.index && a.lexer.IsEqual(other.lexer)
}

// New creates a Tokenizer positioned at the start of JavaScript source.
func New() *Tokenizer {
	return &Tokenizer{
		lexer: newLexer(),
	}
}

// GetContext returns the context of the current position in the source.
func (t *Tokenizer) GetContext() Context {
	return t.lexer.GetContext()
}

// RecordValue records that a value, such as a Go expression, has been inserted into the
// source between start and the current position of pi, and returns the context of the
// value. It returns ErrAmbiguousContext if the interpretations of the source differ
// about the context of the value.
func (t *Tokenizer) RecordValue(pi *parse.Input, start int) (context Context, err error) {
	// The lexer has read the source up to start, so the alternatives are advanced to start
	// before the value is recorded.
	err = t.advanceAlternatives(pi, start)
	if err != nil {
		return t.lexer.GetContext(), err
	}
	context = t.lexer.GetContext()
	t.lexer.RecordValue()
	end := pi.Index()

	isAmbiguous := t.hasTooManyAlternatives
	for i := range t.alternatives {
		if t.alternatives[i].index != start || t.alternatives[i].lexer.GetContext() != context {
			// The interpretation read the position of the value as part of a longer token,
			// e.g. a comment, or has the value in a different context.
			isAmbiguous = true
			continue
		}
		t.alternatives[i].lexer.RecordValue()
		t.alternatives[i].index = end
	}
	if isAmbiguous {
		return context, ErrAmbiguousContext
	}
	t.mergeAlternatives(end)
	return context, nil
}

// ReadComment reads a comment, if the input is at the start of one in code. Within a
// string, template literal, or regular expression, // and /* are data, so no comment is
// read.
func (t *Tokenizer) ReadComment(pi *parse.Input) (comment string, ok bool, err error) {
	comment, ok, err = t.lexer.ReadComment(pi)
	if err != nil || !ok {
		return comment, ok, err
	}
	return comment, true, t.advanceAlternatives(pi, pi.Index())
}

// Read reads the next token of JavaScript source, updates the lexical state, and returns
// the source text of the token. It returns false at the end of the input.
func (t *Tokenizer) Read(pi *parse.Input) (token string, ok bool, err error) {
	token, ok, err = t.lexer.Read(pi)
	if err != nil || !ok {
		return token, ok, err
	}
	t.addAlternative(t.lexer, pi.Index())
	return token, true, t.advanceAlternatives(pi, pi.Index())
}

// addAlternative adds the Alternative of l, if it has one, as an interpretation of the
// source read up to index.
func (t *Tokenizer) addAlternative(l *lexer, index int) {
	if l.Alternative == nil {
		return
	}
	t.alternatives = append(t.alternatives, alternative{lexer: l.Alternative, index: index})
	l.Alternative = nil
}

// advanceAlternatives reads the source with each alternative interpretation, until it
// reaches target, and then merges the interpretations that have the same state. The
// position of pi is restored afterwards.
func (t *Tokenizer) advanceAlternatives(pi *parse.Input, target int) (err error) {
	index := pi.Index()
	defer pi.Seek(index)
	// An alternative can add further alternatives, which are advanced in turn.
	for i := 0; i < len(t.alternatives); i++ {
		pi.Seek(t.alternatives[i].index)
		for pi.Index() < target {
			_, ok, err := t.alternatives[i].lexer.Read(pi)
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			t.addAlternative(t.alternatives[i].lexer, pi.Index())
		}
		t.alternatives[i].index = pi.Index()
	}
	t.mergeAlternatives(target)
	return nil
}

// mergeAlternatives removes each alternative at index that has the same state as the
// lexer, or as another alternative.
func (t *Tokenizer) mergeAlternatives(index int) {
	var merged []alternative
	for _, a := range t.alternatives {
		if a.index == index && a.lexer.IsEqual(t.lexer) {
			continue
		}
		if slices.ContainsFunc(merged, a.IsEqual) {
			continue
		}
		merged = append(merged, a)
	}
	t.alternatives = merged
	if len(t.alternatives) > maxAlternatives {
		t.hasTooManyAlternatives = true
		t.alternatives = nil
	}
}
