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
	"slices"
	"strings"
	"unicode"

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
)

type quote string

const (
	quoteNone     quote = ""
	quoteSingle   quote = `'`
	quoteDouble   quote = `"`
	quoteBacktick quote = "`"
)

// Tokenizer reads JavaScript source one token at a time, and tracks the lexical state
// of the source read so far.
type Tokenizer struct {
	// quote is the delimiter of the string or template literal being read, or quoteNone
	// when code is being read.
	quote quote
	// substitutionBraceDepths holds an entry for each open ${ ... } substitution within a
	// template literal, innermost last. The contents of a substitution are code, so quote
	// is quoteNone within one. Each entry counts the unmatched '{' tokens in code since
	// the substitution opened. When the count reaches zero, the substitution ends and the
	// enclosing template literal resumes.
	substitutionBraceDepths []int
	// regexpAllowed is true when a '/' in code starts a regular expression literal, and
	// false when it is a division operator. A regular expression can only appear where an
	// expression starts, which is determined by the previous token: `x = /a/` is a regular
	// expression, while `x / a / b` and `(x) / a` are divisions.
	regexpAllowed bool
}

// New creates a Tokenizer positioned at the start of JavaScript source.
func New() *Tokenizer {
	return &Tokenizer{
		regexpAllowed: true,
	}
}

// GetContext returns the context of the current position in the source.
func (t *Tokenizer) GetContext() Context {
	switch t.quote {
	case quoteBacktick:
		return ContextTemplateLiteral
	case quoteSingle, quoteDouble:
		return ContextString
	}
	return ContextCode
}

// RecordValue records that a value has been inserted at the current position, such as
// a Go expression. In code, a value ends an expression, so a '/' after it is a division.
func (t *Tokenizer) RecordValue() {
	t.regexpAllowed = false
}

// ReadComment reads a comment, if the input is at the start of one in code. Within a
// string or template literal, // and /* are string data, so no comment is read.
func (t *Tokenizer) ReadComment(pi *parse.Input) (comment string, ok bool, err error) {
	if t.quote != quoteNone {
		return "", false, nil
	}
	return jsComment.Parse(pi)
}

// Read reads the next token of JavaScript source, updates the lexical state, and returns
// the source text of the token. It returns false at the end of the input.
func (t *Tokenizer) Read(pi *parse.Input) (token string, ok bool, err error) {
	if t.quote != quoteNone {
		return t.readStringLiteralToken(pi)
	}
	return t.readCodeToken(pi)
}

func (t *Tokenizer) readStringLiteralToken(pi *parse.Input) (token string, ok bool, err error) {
	token, ok, err = jsCharacter.Parse(pi)
	if err != nil || !ok {
		return token, ok, err
	}
	if t.quote == quoteBacktick && token == "$" {
		peeked, peekOK := pi.Peek(1)
		if peekOK && peeked == "{" {
			pi.Take(1)
			t.quote = quoteNone
			t.substitutionBraceDepths = append(t.substitutionBraceDepths, 1)
			t.regexpAllowed = true
			return "${", true, nil
		}
	}
	if token == string(t.quote) {
		t.quote = quoteNone
		t.regexpAllowed = false
	}
	return token, true, nil
}

func (t *Tokenizer) readCodeToken(pi *parse.Input) (token string, ok bool, err error) {
	token, ok, err = jsComment.Parse(pi)
	if err != nil || ok {
		return token, ok, err
	}

	if t.regexpAllowed {
		token, ok, err = regexpLiteral.Parse(pi)
		if err != nil {
			return token, false, err
		}
		if ok {
			t.regexpAllowed = false
			return token, true, nil
		}
	}

	token, ok, err = jsIdentifierName.Parse(pi)
	if err != nil {
		return token, false, err
	}
	if ok {
		t.regexpAllowed = slices.Contains(jsKeywordsPrecedingExpression, token)
		return token, true, nil
	}

	token, ok, err = jsCharacter.Parse(pi)
	if err != nil || !ok {
		return token, ok, err
	}
	switch token {
	case string(quoteSingle), string(quoteDouble), string(quoteBacktick):
		t.quote = quote(token)
	case "{":
		if len(t.substitutionBraceDepths) > 0 {
			t.substitutionBraceDepths[len(t.substitutionBraceDepths)-1]++
		}
		t.regexpAllowed = true
	case "}":
		t.regexpAllowed = true
		if len(t.substitutionBraceDepths) == 0 {
			break
		}
		innermost := len(t.substitutionBraceDepths) - 1
		t.substitutionBraceDepths[innermost]--
		if t.substitutionBraceDepths[innermost] == 0 {
			t.substitutionBraceDepths = t.substitutionBraceDepths[:innermost]
			t.quote = quoteBacktick
		}
	case ")", "]":
		t.regexpAllowed = false
	default:
		if strings.HasPrefix(token, `\`) {
			t.regexpAllowed = false
			break
		}
		if strings.TrimSpace(token) != "" {
			t.regexpAllowed = true
		}
	}
	return token, true, nil
}

// jsKeywordsPrecedingExpression are the keywords that can be followed by an expression,
// so a '/' after one of them starts a regular expression rather than a division.
var jsKeywordsPrecedingExpression = []string{
	"await",
	"case",
	"delete",
	"do",
	"else",
	"in",
	"instanceof",
	"new",
	"of",
	"return",
	"throw",
	"typeof",
	"void",
	"yield",
}

var jsIdentifierName = parse.StringFrom(parse.OneOrMore(parse.RuneWhere(isJSIdentifierRune)))

func isJSIdentifierRune(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

var jsCharacter = parse.Any(jsEscapedCharacter, parse.AnyRune)

// \uXXXX	Unicode code point escape	'\u0061' = 'a'
var (
	hexDigit        = parse.Any(parse.ZeroToNine, parse.RuneIn("abcdef"), parse.RuneIn("ABCDEF"))
	jsUnicodeEscape = parse.StringFrom(parse.String("\\u"), hexDigit, hexDigit, hexDigit, hexDigit)
)

// \u{X...}	ES6+ extended Unicode escape	'\u{1F600}' = '😀'
var jsExtendedUnicodeEscape = parse.StringFrom(parse.String("\\u{"), hexDigit, parse.StringFrom(parse.AtLeast(1, parse.ZeroOrMore(hexDigit))), parse.String("}"))

// \xXX	Hex code (2-digit)	'\x41' = 'A'
var jsHexEscape = parse.StringFrom(parse.String("\\x"), hexDigit, hexDigit)

// \x Backslash escape	'\\' = '\'
var jsBackslashEscape = parse.StringFrom(parse.String("\\"), parse.AnyRune)

// All escapes.
var jsEscapedCharacter = parse.Any(jsBackslashEscape, jsUnicodeEscape, jsHexEscape, jsExtendedUnicodeEscape)

var jsComment = parse.Any(jsSingleLineComment, jsMultiLineComment)

var (
	jsStartSingleLineComment = parse.String("//")
	jsEndOfSingleLineComment = parse.StringFrom(parse.Or(parse.NewLine, parse.EOF[string]()))
	jsSingleLineComment      = parse.StringFrom(jsStartSingleLineComment, parse.StringUntil(jsEndOfSingleLineComment), jsEndOfSingleLineComment)
)

var (
	jsStartMultiLineComment = parse.String("/*")
	jsEndOfMultiLineComment = parse.StringFrom(parse.Or(parse.String("*/"), parse.EOF[string]()))
	jsMultiLineComment      = parse.StringFrom(jsStartMultiLineComment, parse.StringUntil(jsEndOfMultiLineComment), jsEndOfMultiLineComment, parse.OptionalWhitespace)
)

var regexpLiteral = parse.Func(func(in *parse.Input) (regexp string, ok bool, err error) {
	startIndex := in.Index()

	// Take the initial '/'.
	s, ok := in.Take(1)
	if !ok || s != "/" {
		in.Seek(startIndex)
		return "", false, nil
	}
	// Peek the next char. If it's also a '/', then this is not a regex literal, but the start of a comment.
	p, ok := in.Peek(1)
	if !ok || p == "/" {
		in.Seek(startIndex)
		return "", false, nil
	}
	var literal strings.Builder
	literal.WriteString(s)

	var inClass, escaped bool

	for {
		s, ok := in.Take(1)
		if !ok {
			// Restore position if no closing '/'.
			in.Seek(startIndex)
			return "", false, nil
		}

		literal.WriteString(s)

		if escaped {
			escaped = false
			continue
		}

		switch s {
		case "\n", "\r":
			// Newline in a regex is not allowed, so we restore the position and return false.
			in.Seek(startIndex)
			return "", false, nil
		case "\\":
			escaped = true
		case "[":
			inClass = true
		case "]":
			inClass = false
		case "/":
			if !inClass {
				// We've reached the end of the regex, but there may be flags after it.
				// Read flags until we hit a non-flag character.
				flags, ok, err := regexpFlags.Parse(in)
				if err != nil {
					return "", false, err
				}
				if ok {
					literal.WriteString(flags)
				}
				output := literal.String()
				if strings.Contains(output, "{{") && strings.Contains(output, "}}") {
					// If the regex contains a Go expression, don't treat it as a regex literal.
					in.Seek(startIndex)
					return "", false, nil
				}
				return output, true, nil
			}
		}
	}
})

var regexpFlags = parse.StringFrom(parse.Repeat(0, 5, parse.RuneIn("gimuy")))
