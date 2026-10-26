package jstokenizer

import (
	"slices"
	"unicode"

	"github.com/a-h/parse"
)

type quote string

const (
	quoteNone     quote = ""
	quoteSingle   quote = `'`
	quoteDouble   quote = `"`
	quoteBacktick quote = "`"
	quoteRegexp   quote = "/"
)

// lexer reads JavaScript source one token at a time, and tracks the lexical state of
// the source read so far, for one interpretation of the source.
type lexer struct {
	// quote is the delimiter of the string, template literal, or regular expression being
	// read, or quoteNone when code is being read.
	quote quote
	// inRegexpClass is true within a [ ] character class of a regular expression, where
	// '/' does not end the regular expression.
	inRegexpClass bool
	// atRegexpStart is true when nothing has been read since the '/' that opened the
	// regular expression being read.
	atRegexpStart bool
	// substitutionBraceDepths holds an entry for each open ${ ... } substitution within a
	// template literal, innermost last. The contents of a substitution are code, so quote
	// is quoteNone within one. Each entry counts the unmatched '{' tokens in code since
	// the substitution opened. When the count reaches zero, the substitution ends and the
	// enclosing template literal resumes.
	substitutionBraceDepths []int
	// slash decides whether a '/' in code starts a regular expression.
	slash slashClassifier
	// comments reads comments in code.
	comments commentReader
	// Alternative is set when the last token read was a '/' that could be either a
	// division or the start of a regular expression. The lexer reads the '/' as one, and
	// Alternative is a copy of the lexer that has read it as the other.
	Alternative *lexer
}

func newLexer() *lexer {
	return &lexer{
		slash:    newSlashClassifier(),
		comments: newCommentReader(),
	}
}

// Clone returns a copy of the lexer, without its Alternative.
func (l *lexer) Clone() *lexer {
	return &lexer{
		quote:                   l.quote,
		inRegexpClass:           l.inRegexpClass,
		atRegexpStart:           l.atRegexpStart,
		substitutionBraceDepths: slices.Clone(l.substitutionBraceDepths),
		slash:                   l.slash.Clone(),
		comments:                l.comments,
	}
}

// IsEqual returns true if the lexers have the same state, so that they read the rest
// of the source in the same way.
func (l *lexer) IsEqual(other *lexer) bool {
	return l.quote == other.quote &&
		l.inRegexpClass == other.inRegexpClass &&
		l.atRegexpStart == other.atRegexpStart &&
		slices.Equal(l.substitutionBraceDepths, other.substitutionBraceDepths) &&
		l.slash.IsEqual(other.slash) &&
		l.comments == other.comments
}

// GetContext returns the context of the current position in the source.
func (l *lexer) GetContext() Context {
	switch l.quote {
	case quoteBacktick:
		return ContextTemplateLiteral
	case quoteSingle, quoteDouble:
		return ContextString
	case quoteRegexp:
		if l.atRegexpStart {
			return ContextRegexpStart
		}
		return ContextRegexp
	}
	return ContextCode
}

// RecordValue records that a value has been inserted at the current position, such as
// a Go expression. In code, a value ends an expression, so a '/' after it is a division.
func (l *lexer) RecordValue() {
	l.atRegexpStart = false
	if l.quote != quoteNone {
		return
	}
	l.recordExpressionEnd(insertedValueToken)
}

// insertedValueToken is recorded as the previous token in code when a value is
// inserted.
const insertedValueToken = "{{ }}"

// ReadComment reads a comment, if the input is at the start of one in code. Within a
// string, template literal, or regular expression, // and /* are data, so no comment is
// read.
func (l *lexer) ReadComment(pi *parse.Input) (comment string, ok bool, err error) {
	if l.quote != quoteNone {
		return "", false, nil
	}
	comment, ok, err = l.comments.Read(pi)
	if err != nil || !ok {
		return comment, ok, err
	}
	l.slash.RecordComment(comment)
	return comment, true, nil
}

// Read reads the next token of JavaScript source, updates the lexical state, and returns
// the source text of the token. It returns false at the end of the input.
func (l *lexer) Read(pi *parse.Input) (token string, ok bool, err error) {
	switch l.quote {
	case quoteNone:
		return l.readCodeToken(pi)
	case quoteRegexp:
		return l.readRegexpToken(pi)
	}
	return l.readStringLiteralToken(pi)
}

func (l *lexer) readStringLiteralToken(pi *parse.Input) (token string, ok bool, err error) {
	token, ok, err = jsCharacter.Parse(pi)
	if err != nil || !ok {
		return token, ok, err
	}
	if l.quote == quoteBacktick && token == "$" {
		peeked, peekOK := pi.Peek(1)
		if peekOK && peeked == "{" {
			pi.Take(1)
			l.quote = quoteNone
			l.substitutionBraceDepths = append(l.substitutionBraceDepths, 1)
			l.slash.RecordExpressionStart("${")
			l.comments.Record("${")
			return "${", true, nil
		}
	}
	if token == string(l.quote) {
		l.quote = quoteNone
		l.recordExpressionEnd(token)
		return token, true, nil
	}
	if (l.quote == quoteSingle || l.quote == quoteDouble) && (token == "\n" || token == "\r") {
		// A line feed or carriage return in a single or double quoted string is a syntax
		// error. End the string, so that the rest of the source is read as code, as an
		// unterminated regular expression is. U+2028 and U+2029 are allowed in a string.
		l.quote = quoteNone
		l.recordCodeToken(token)
	}
	return token, true, nil
}

func (l *lexer) readRegexpToken(pi *parse.Input) (token string, ok bool, err error) {
	token, ok, err = jsCharacter.Parse(pi)
	if err != nil || !ok {
		return token, ok, err
	}
	l.atRegexpStart = false
	switch token {
	case "[":
		l.inRegexpClass = true
	case "]":
		l.inRegexpClass = false
	case "/":
		if l.inRegexpClass {
			break
		}
		// The flags that follow are read as an identifier, which is not a keyword, so a
		// '/' after them is division.
		l.quote = quoteNone
		l.recordExpressionEnd(token)
	case "\n", "\r", "\u2028", "\u2029":
		// A line terminator in a regular expression is a syntax error. End the regular
		// expression, so that the rest of the source is read as code.
		l.quote = quoteNone
		l.inRegexpClass = false
		l.recordCodeToken(token)
	}
	return token, true, nil
}

func (l *lexer) readCodeToken(pi *parse.Input) (token string, ok bool, err error) {
	token, ok, err = l.ReadComment(pi)
	if err != nil || ok {
		return token, ok, err
	}

	token, ok, err = jsCodeToken.Parse(pi)
	if err != nil || !ok {
		return token, ok, err
	}
	switch token {
	case string(quoteSingle), string(quoteDouble), string(quoteBacktick):
		l.quote = quote(token)
		l.comments.Record(token)
		return token, true, nil
	case "/":
		if l.slash.IsAmbiguous() {
			l.Alternative = l.Clone()
			l.Alternative.readSlash(!l.slash.RegexpAllowed)
		}
		l.readSlash(l.slash.RegexpAllowed)
		return token, true, nil
	case "{":
		if len(l.substitutionBraceDepths) > 0 {
			l.substitutionBraceDepths[len(l.substitutionBraceDepths)-1]++
		}
	case "}":
		if len(l.substitutionBraceDepths) == 0 {
			break
		}
		innermost := len(l.substitutionBraceDepths) - 1
		l.substitutionBraceDepths[innermost]--
		if l.substitutionBraceDepths[innermost] == 0 {
			l.substitutionBraceDepths = l.substitutionBraceDepths[:innermost]
			l.quote = quoteBacktick
		}
	}
	l.recordCodeToken(token)
	return token, true, nil
}

// readSlash records a '/' read in code, as the start of a regular expression if regexp
// is true, or as a division.
func (l *lexer) readSlash(regexp bool) {
	if !regexp {
		l.recordCodeToken("/")
		return
	}
	l.quote = quoteRegexp
	l.inRegexpClass = false
	l.atRegexpStart = true
	l.slash.RecordRegexpStart()
	l.comments.Record("/")
}

// recordCodeToken records a token read in code.
func (l *lexer) recordCodeToken(token string) {
	l.slash.Record(token)
	l.comments.Record(token)
}

// recordExpressionEnd records a token in code that ends an expression.
func (l *lexer) recordExpressionEnd(token string) {
	l.slash.RecordExpressionEnd(token)
	l.comments.Record(token)
}

// jsCodeToken reads a number, an identifier or keyword, an escape sequence, or a single
// character.
var jsCodeToken = parse.Any(jsNumericLiteralWithDecimalPoint, jsIdentifierName, jsCharacter)

// jsNumericLiteralWithDecimalPoint reads a decimal number that contains a decimal point,
// such as 1.5, 1., 1.e3, or .5. Without it, the digits are read as an identifier, and the
// '.' that follows as a property access, so a '/' after `1.` would start a regular
// expression rather than a division.
var jsNumericLiteralWithDecimalPoint = parse.Any(
	parse.StringFrom(
		parse.ZeroToNine,
		parse.StringFrom(parse.ZeroOrMore(parse.RuneWhere(isDecimalDigitOrSeparator))),
		parse.String("."),
		parse.StringFrom(parse.ZeroOrMore(jsIdentifierRune)),
	),
	parse.StringFrom(
		parse.String("."),
		parse.ZeroToNine,
		parse.StringFrom(parse.ZeroOrMore(jsIdentifierRune)),
	),
)

func isDecimalDigitOrSeparator(r rune) bool {
	return (r >= '0' && r <= '9') || r == '_'
}

var jsIdentifierName = parse.StringFrom(parse.OneOrMore(jsIdentifierRune))

var jsIdentifierRune parse.Parser[string] = utf8RuneParser{Predicate: isJSIdentifierRune}

// isJSIdentifierRune returns true if r can be part of an identifier, see
// https://tc39.es/ecma262/#prod-IdentifierPartChar. This includes combining marks, e.g.
// U+05B1 in `A\u05B1 / 2`, which are part of the identifier rather than punctuation.
func isJSIdentifierRune(r rune) bool {
	return r == '$' || r == '\u200c' || r == '\u200d' || unicode.In(r, unicode.L, unicode.Nl, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc)
}

var jsCharacter = parse.Any(jsEscapedCharacter, jsRune)

var jsRune parse.Parser[string] = utf8RuneParser{Predicate: isAnyRune}

func isAnyRune(r rune) bool {
	return true
}

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
var jsBackslashEscape = parse.StringFrom(parse.String("\\"), jsRune)

// All escapes.
var jsEscapedCharacter = parse.Any(jsBackslashEscape, jsUnicodeEscape, jsHexEscape, jsExtendedUnicodeEscape)
