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
	// ContextRegexp values are written into the pattern of a regular expression
	// literal.
	ContextRegexp
)

type quote string

const (
	quoteNone     quote = ""
	quoteSingle   quote = `'`
	quoteDouble   quote = `"`
	quoteBacktick quote = "`"
	quoteRegexp   quote = "/"
)

// Tokenizer reads JavaScript source one token at a time, and tracks the lexical state
// of the source read so far.
type Tokenizer struct {
	// quote is the delimiter of the string, template literal, or regular expression being
	// read, or quoteNone when code is being read.
	quote quote
	// inRegexpClass is true within a [ ] character class of a regular expression, where
	// '/' does not end the regular expression.
	inRegexpClass bool
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
	// parenthesesCloseCondition holds an entry for each open '(' in code, innermost last.
	// An entry is true when the parenthesis opens the condition of a statement such as
	// if or while. A statement follows the condition, so a '/' after the ')' starts a
	// regular expression, e.g. `if (a) /b/.test(s)`, while after other parentheses it is
	// division, e.g. `(a) / b`.
	parenthesesCloseCondition []bool
	// previousToken is the previous token read in code, including whitespace, so that
	// the second character of `++` and `--` can be recognized.
	previousToken string
	// previousKeyword is the previous non-whitespace token read in code if it is a
	// keyword, or empty otherwise. An identifier after '.' is a property name, not a
	// keyword, e.g. return in `b.return / 2`.
	previousKeyword string
	// previousTokenIsDot is true when the previous non-whitespace token read in code is
	// '.', so that the next identifier is a property name.
	previousTokenIsDot bool
	// atLineStart is true when only whitespace and comments have been read in code since
	// the start of the source or the last line terminator, where --> starts a comment.
	atLineStart bool
}

// New creates a Tokenizer positioned at the start of JavaScript source.
func New() *Tokenizer {
	return &Tokenizer{
		regexpAllowed: true,
		atLineStart:   true,
	}
}

// GetContext returns the context of the current position in the source.
func (t *Tokenizer) GetContext() Context {
	switch t.quote {
	case quoteBacktick:
		return ContextTemplateLiteral
	case quoteSingle, quoteDouble:
		return ContextString
	case quoteRegexp:
		return ContextRegexp
	}
	return ContextCode
}

// RecordValue records that a value has been inserted at the current position, such as
// a Go expression. In code, a value ends an expression, so a '/' after it is a division.
func (t *Tokenizer) RecordValue() {
	if t.quote != quoteNone {
		return
	}
	t.regexpAllowed = false
	t.recordCodeToken(insertedValueToken)
}

// insertedValueToken is recorded as the previous token in code when a value is
// inserted.
const insertedValueToken = "{{ }}"

// ReadComment reads a comment, if the input is at the start of one in code. Within a
// string, template literal, or regular expression, // and /* are data, so no comment is
// read.
func (t *Tokenizer) ReadComment(pi *parse.Input) (comment string, ok bool, err error) {
	if t.quote != quoteNone {
		return "", false, nil
	}
	comment, ok, err = jsComment.Parse(pi)
	if err != nil || ok {
		t.recordComment(comment)
		return comment, ok, err
	}
	if !t.atLineStart {
		return "", false, nil
	}
	comment, ok, err = jsHTMLCloseComment.Parse(pi)
	t.recordComment(comment)
	return comment, ok, err
}

// recordComment records that comment has been read in code. A comment is whitespace,
// so it doesn't change whether a '/' starts a regular expression, but a comment that
// contains a line terminator starts a new line.
func (t *Tokenizer) recordComment(comment string) {
	if strings.ContainsAny(comment, "\n\r  ") {
		t.atLineStart = true
	}
}

// Read reads the next token of JavaScript source, updates the lexical state, and returns
// the source text of the token. It returns false at the end of the input.
func (t *Tokenizer) Read(pi *parse.Input) (token string, ok bool, err error) {
	switch t.quote {
	case quoteNone:
		return t.readCodeToken(pi)
	case quoteRegexp:
		return t.readRegexpToken(pi)
	}
	return t.readStringLiteralToken(pi)
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
			t.recordCodeToken("${")
			return "${", true, nil
		}
	}
	if token == string(t.quote) {
		t.quote = quoteNone
		t.regexpAllowed = false
		t.recordCodeToken(token)
	}
	return token, true, nil
}

func (t *Tokenizer) readRegexpToken(pi *parse.Input) (token string, ok bool, err error) {
	token, ok, err = jsCharacter.Parse(pi)
	if err != nil || !ok {
		return token, ok, err
	}
	switch token {
	case "[":
		t.inRegexpClass = true
	case "]":
		t.inRegexpClass = false
	case "/":
		if t.inRegexpClass {
			break
		}
		// The flags that follow are read as an identifier, which is not a keyword, so a
		// '/' after them is division.
		t.quote = quoteNone
		t.regexpAllowed = false
		t.recordCodeToken(token)
	case "\n", "\r", " ", " ":
		// A line terminator in a regular expression is a syntax error. End the regular
		// expression, so that the rest of the source is read as code.
		t.quote = quoteNone
		t.inRegexpClass = false
		t.regexpAllowed = true
		t.recordCodeToken(token)
	}
	return token, true, nil
}

func (t *Tokenizer) readCodeToken(pi *parse.Input) (token string, ok bool, err error) {
	token, ok, err = t.ReadComment(pi)
	if err != nil || ok {
		return token, ok, err
	}

	token, ok, err = jsIdentifierName.Parse(pi)
	if err != nil {
		return token, false, err
	}
	if ok {
		t.regexpAllowed = !t.previousTokenIsDot && slices.Contains(jsKeywordsPrecedingExpression, token)
		t.recordCodeToken(token)
		return token, true, nil
	}

	token, ok, err = jsCharacter.Parse(pi)
	if err != nil || !ok {
		return token, ok, err
	}
	recorded := token
	switch token {
	case string(quoteSingle), string(quoteDouble), string(quoteBacktick):
		t.quote = quote(token)
	case "/":
		if t.regexpAllowed {
			t.quote = quoteRegexp
			t.inRegexpClass = false
			break
		}
		t.regexpAllowed = true
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
	case "(":
		t.parenthesesCloseCondition = append(t.parenthesesCloseCondition, slices.Contains(jsKeywordsPrecedingCondition, t.previousKeyword))
		t.regexpAllowed = true
	case ")":
		t.regexpAllowed = false
		if len(t.parenthesesCloseCondition) == 0 {
			break
		}
		innermost := len(t.parenthesesCloseCondition) - 1
		t.regexpAllowed = t.parenthesesCloseCondition[innermost]
		t.parenthesesCloseCondition = t.parenthesesCloseCondition[:innermost]
	case "]":
		t.regexpAllowed = false
	case "+", "-":
		// The second character of a ++ or -- operator. A postfix increment or decrement
		// ends an expression, e.g. `i++ / 2`. A '/' can't directly follow a prefix
		// increment or decrement, because its operand must be a reference.
		if t.previousToken == token {
			t.regexpAllowed = false
			recorded = t.previousToken + token
			break
		}
		t.regexpAllowed = true
	default:
		if strings.HasPrefix(token, `\`) {
			t.regexpAllowed = false
			break
		}
		if strings.TrimSpace(token) != "" {
			t.regexpAllowed = true
		}
	}
	t.recordCodeToken(recorded)
	return token, true, nil
}

// recordCodeToken records token as the previous token read in code.
func (t *Tokenizer) recordCodeToken(token string) {
	t.previousToken = token
	if strings.ContainsAny(token, "\n\r  ") {
		t.atLineStart = true
		return
	}
	if strings.TrimSpace(token) == "" {
		return
	}
	t.atLineStart = false
	if token == "await" && t.previousKeyword == "for" {
		// The condition of a for await statement follows the await keyword.
		t.previousTokenIsDot = false
		return
	}
	t.previousKeyword = ""
	if !t.previousTokenIsDot && (slices.Contains(jsKeywordsPrecedingExpression, token) || slices.Contains(jsKeywordsPrecedingCondition, token)) {
		t.previousKeyword = token
	}
	t.previousTokenIsDot = token == "."
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

// jsKeywordsPrecedingCondition are the keywords that are followed by a parenthesized
// condition and then a statement, so a '/' after the closing parenthesis of the
// condition starts a regular expression rather than a division.
var jsKeywordsPrecedingCondition = []string{
	"for",
	"if",
	"while",
	"with",
}

var jsIdentifierName = parse.StringFrom(parse.OneOrMore(parse.RuneWhere(isJSIdentifierRune)))

func isJSIdentifierRune(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

var jsCharacter = parse.Any(jsEscapedCharacter, parse.AnyRune)

// \uXXXX	Unicode code point escape	'a' = 'a'
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

// jsComment includes <!--, which starts a single line comment in a classic script.
var jsComment = parse.Any(jsSingleLineComment, jsMultiLineComment, jsHTMLOpenComment)

var (
	jsStartSingleLineComment = parse.String("//")
	jsEndOfSingleLineComment = parse.StringFrom(parse.Or(parse.NewLine, parse.EOF[string]()))
	jsSingleLineComment      = parse.StringFrom(jsStartSingleLineComment, parse.StringUntil(jsEndOfSingleLineComment), jsEndOfSingleLineComment)
	jsHTMLOpenComment        = parse.StringFrom(parse.String("<!--"), parse.StringUntil(jsEndOfSingleLineComment), jsEndOfSingleLineComment)
	// jsHTMLCloseComment is a comment in a classic script only at the start of a
	// line. Elsewhere, `a --> b` is `a-- > b`.
	jsHTMLCloseComment = parse.StringFrom(parse.String("-->"), parse.StringUntil(jsEndOfSingleLineComment), jsEndOfSingleLineComment)
)

var (
	jsStartMultiLineComment = parse.String("/*")
	jsEndOfMultiLineComment = parse.StringFrom(parse.Or(parse.String("*/"), parse.EOF[string]()))
	jsMultiLineComment      = parse.StringFrom(jsStartMultiLineComment, parse.StringUntil(jsEndOfMultiLineComment), jsEndOfMultiLineComment, parse.OptionalWhitespace)
)
