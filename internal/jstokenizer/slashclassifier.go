package jstokenizer

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// slashClassifier decides whether a '/' in code starts a regular expression literal or
// is a division operator. A regular expression can only appear where an expression
// starts, which depends on the grammar around the '/'. The classifier approximates the
// grammar from the tokens read in code, and reports a '/' as ambiguous where the tokens
// don't determine it. See ADR 0005 and ADR 0009.
type slashClassifier struct {
	// RegexpAllowed is true when a '/' in code starts a regular expression literal, and
	// false when it is a division operator: `x = /a/` is a regular expression, while
	// `x / a / b` and `(x) / a` are divisions. Where IsAmbiguous returns true, it is the
	// more likely of the two.
	RegexpAllowed bool
	// parenthesesCloseCondition holds an entry for each open '(' in code, innermost last.
	// An entry is true when the parenthesis opens the condition of a statement such as
	// if or while. A statement follows the condition, so a '/' after the ')' starts a
	// regular expression, e.g. `if (a) /b/.test(s)`, while after other parentheses it is
	// division, e.g. `(a) / b`.
	parenthesesCloseCondition []bool
	// previousToken is the previous token, including whitespace, so that the second
	// character of `++` and `--` can be recognized.
	previousToken string
	// previousKeyword is the previous non-whitespace token if it is a keyword, or empty
	// otherwise. An identifier after '.' is a property name, not a keyword, e.g. return
	// in `b.return / 2`.
	previousKeyword string
	// previousTokenIsDot is true when the previous non-whitespace token is '.', so that
	// the next identifier is a property name.
	previousTokenIsDot bool
	// lineTerminatorSincePreviousToken is true when a line terminator, or a comment that
	// contains one, has been read since the previous non-whitespace token.
	lineTerminatorSincePreviousToken bool
	// isAmbiguousAfterPreviousToken is true when the previous non-whitespace token doesn't
	// determine whether a '/' after it is a division:
	//
	//   - '}' ends a block, after which a '/' starts a regular expression, or an object
	//     literal or a function expression, after which it is division.
	//   - yield, await, and of are keywords in a generator, an async function or module,
	//     and a for ... of statement respectively, and identifiers elsewhere.
	isAmbiguousAfterPreviousToken bool
	// isAmbiguousAfterLineTerminator is true when the previous non-whitespace token is an
	// identifier. Automatic semicolon insertion can end a statement after an identifier
	// at a line terminator, e.g. after a declaration without an initializer, in which
	// case a '/' on the next line starts a regular expression: `var a\n/x/`.
	isAmbiguousAfterLineTerminator bool
	// operatorFollowsExpression is true when the previous token is a '+' or '-' that
	// follows an expression on the same line, so that a second '+' or '-' makes a postfix
	// increment or decrement rather than a prefix one.
	operatorFollowsExpression bool
}

func newSlashClassifier() slashClassifier {
	return slashClassifier{
		RegexpAllowed: true,
	}
}

// Clone returns a copy of the classifier.
func (c slashClassifier) Clone() slashClassifier {
	c.parenthesesCloseCondition = slices.Clone(c.parenthesesCloseCondition)
	return c
}

// IsEqual returns true if the classifiers have the same state.
func (c slashClassifier) IsEqual(other slashClassifier) bool {
	return c.isEqualExceptParentheses(other) && slices.Equal(c.parenthesesCloseCondition, other.parenthesesCloseCondition)
}

func (c slashClassifier) isEqualExceptParentheses(other slashClassifier) bool {
	return c.RegexpAllowed == other.RegexpAllowed &&
		c.previousToken == other.previousToken &&
		c.previousKeyword == other.previousKeyword &&
		c.previousTokenIsDot == other.previousTokenIsDot &&
		c.lineTerminatorSincePreviousToken == other.lineTerminatorSincePreviousToken &&
		c.isAmbiguousAfterPreviousToken == other.isAmbiguousAfterPreviousToken &&
		c.isAmbiguousAfterLineTerminator == other.isAmbiguousAfterLineTerminator &&
		c.operatorFollowsExpression == other.operatorFollowsExpression
}

// IsAmbiguous returns true if a '/' at the current position could be either a division
// or the start of a regular expression, depending on the grammar.
func (c *slashClassifier) IsAmbiguous() bool {
	return c.isAmbiguousAfterPreviousToken || (c.isAmbiguousAfterLineTerminator && c.lineTerminatorSincePreviousToken)
}

// Record records a token read in code: an identifier, a keyword, a number, an escape
// sequence, whitespace, or a single punctuation character. A '/' that starts a regular
// expression is recorded with RecordRegexpStart instead.
func (c *slashClassifier) Record(token string) {
	if strings.TrimSpace(token) == "" {
		c.recordWhitespace(token)
		return
	}
	recorded := token
	var isAmbiguousAfterToken, isAmbiguousAfterLineTerminator bool
	first, _ := utf8.DecodeRuneInString(token)
	switch {
	case len(token) > 1 && first == '.':
		// A number that starts with a decimal point, e.g. .5.
		c.RegexpAllowed = false
	case isJSIdentifierRune(first):
		// A number, an identifier, or a keyword. Only a keyword can be followed by an
		// expression.
		isPropertyName := c.previousTokenIsDot
		isKeyword := !isPropertyName && slices.Contains(jsKeywordsPrecedingExpression, token)
		c.RegexpAllowed = isKeyword
		isAmbiguousAfterToken = !isPropertyName && slices.Contains(jsContextualKeywords, token)
		isAmbiguousAfterLineTerminator = !isPropertyName && !isKeyword && !unicode.IsDigit(first)
	case token == "(":
		c.parenthesesCloseCondition = append(c.parenthesesCloseCondition, slices.Contains(jsKeywordsPrecedingCondition, c.previousKeyword))
		c.RegexpAllowed = true
	case token == ")":
		c.RegexpAllowed = false
		if len(c.parenthesesCloseCondition) == 0 {
			break
		}
		innermost := len(c.parenthesesCloseCondition) - 1
		c.RegexpAllowed = c.parenthesesCloseCondition[innermost]
		c.parenthesesCloseCondition = c.parenthesesCloseCondition[:innermost]
	case token == "]":
		c.RegexpAllowed = false
	case token == "}":
		c.RegexpAllowed = true
		isAmbiguousAfterToken = true
	case (token == "+" || token == "-") && c.previousToken == token:
		// The second character of a ++ or -- operator. A postfix increment or decrement
		// ends an expression, e.g. `i++ / 2`. A prefix increment or decrement is followed
		// by its operand, which can be a regular expression, e.g. `++/x/.lastIndex`.
		c.RegexpAllowed = !c.operatorFollowsExpression
		recorded = c.previousToken + token
	case token == "+" || token == "-":
		// A postfix increment or decrement can't follow a line terminator, so `a\n++b` is
		// `a; ++b`.
		c.operatorFollowsExpression = !c.RegexpAllowed && !c.lineTerminatorSincePreviousToken
		c.RegexpAllowed = true
	case strings.HasPrefix(token, `\`):
		// An escape sequence in code is part of an identifier.
		c.RegexpAllowed = false
	default:
		c.RegexpAllowed = true
	}
	c.recordPreviousToken(recorded)
	c.isAmbiguousAfterPreviousToken = isAmbiguousAfterToken
	c.isAmbiguousAfterLineTerminator = isAmbiguousAfterLineTerminator
}

// RecordComment records a comment read in code. A comment is whitespace, and one that
// contains a line terminator is a line terminator.
func (c *slashClassifier) RecordComment(comment string) {
	c.recordWhitespace(comment)
}

// RecordRegexpStart records the '/' that starts a regular expression literal.
func (c *slashClassifier) RecordRegexpStart() {
	c.recordPreviousToken("/")
}

// RecordExpressionEnd records a token that ends an expression: the closing quote of a
// string or template literal, the closing '/' of a regular expression, or an inserted
// value. A '/' after it is division.
func (c *slashClassifier) RecordExpressionEnd(token string) {
	c.RegexpAllowed = false
	c.recordPreviousToken(token)
}

// RecordExpressionStart records a token after which an expression starts, such as the
// ${ of a template literal substitution. A '/' after it starts a regular expression.
func (c *slashClassifier) RecordExpressionStart(token string) {
	c.RegexpAllowed = true
	c.recordPreviousToken(token)
}

func (c *slashClassifier) recordWhitespace(whitespace string) {
	c.previousToken = whitespace
	if strings.ContainsAny(whitespace, "\n\r\u2028\u2029") {
		c.lineTerminatorSincePreviousToken = true
	}
}

// recordPreviousToken records a non-whitespace token.
func (c *slashClassifier) recordPreviousToken(token string) {
	c.previousToken = token
	c.lineTerminatorSincePreviousToken = false
	c.isAmbiguousAfterPreviousToken = false
	c.isAmbiguousAfterLineTerminator = false
	if token == "await" && c.previousKeyword == "for" {
		// The condition of a for await statement follows the await keyword.
		c.previousTokenIsDot = false
		return
	}
	c.previousKeyword = ""
	if !c.previousTokenIsDot && (slices.Contains(jsKeywordsPrecedingExpression, token) || slices.Contains(jsKeywordsPrecedingCondition, token)) {
		c.previousKeyword = token
	}
	c.previousTokenIsDot = token == "."
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

// jsContextualKeywords are the keywords in jsKeywordsPrecedingExpression that are
// identifiers in some contexts, e.g. yield outside a generator, so a '/' after one of
// them could be a division.
var jsContextualKeywords = []string{
	"await",
	"of",
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
