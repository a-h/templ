package jstokenizer

import (
	"strings"

	"github.com/a-h/parse"
)

// commentReader reads comments in code. Whether --> and #! start a comment depends on
// the position in the source, so the reader tracks the position from the comments and
// code tokens read.
type commentReader struct {
	// atLineStart is true when only whitespace and comments have been read since the
	// start of the source or the last line terminator, where --> starts a comment.
	atLineStart bool
	// atSourceStart is true when nothing has been read, where #! starts a comment.
	atSourceStart bool
}

func newCommentReader() commentReader {
	return commentReader{
		atLineStart:   true,
		atSourceStart: true,
	}
}

// Read reads a comment, if the input is at the start of one.
func (r *commentReader) Read(pi *parse.Input) (comment string, ok bool, err error) {
	if r.atSourceStart {
		comment, ok, err = jsHashbangComment.Parse(pi)
		if err != nil || ok {
			r.recordComment(comment)
			return comment, ok, err
		}
	}
	comment, ok, err = jsComment.Parse(pi)
	if err != nil || ok {
		r.recordComment(comment)
		return comment, ok, err
	}
	if !r.atLineStart {
		return "", false, nil
	}
	comment, ok, err = jsHTMLCloseComment.Parse(pi)
	r.recordComment(comment)
	return comment, ok, err
}

// Record records that a token has been read in code.
func (r *commentReader) Record(token string) {
	r.atSourceStart = false
	if strings.ContainsAny(token, "\n\r\u2028\u2029") {
		r.atLineStart = true
		return
	}
	if strings.TrimSpace(token) != "" {
		r.atLineStart = false
	}
}

// recordComment records that a comment has been read. A comment is whitespace, so it
// only changes the position if it contains a line terminator.
func (r *commentReader) recordComment(comment string) {
	r.atSourceStart = false
	if strings.ContainsAny(comment, "\n\r\u2028\u2029") {
		r.atLineStart = true
	}
}

// jsComment includes <!--, which starts a single line comment in a classic script.
var jsComment = parse.Any(jsSingleLineComment, jsMultiLineComment, jsHTMLOpenComment)

var (
	jsStartSingleLineComment = parse.String("//")
	// jsLineTerminator includes \u2028 and \u2029, which end a single line comment.
	jsLineTerminator         = parse.Any(parse.NewLine, parse.Parser[string](utf8RuneParser{Predicate: isJSLineTerminatorRune}))
	jsEndOfSingleLineComment = parse.StringFrom(parse.Or(jsLineTerminator, parse.EOF[string]()))
	jsSingleLineComment      = parse.StringFrom(jsStartSingleLineComment, parse.StringUntil(jsEndOfSingleLineComment), jsEndOfSingleLineComment)
	jsHTMLOpenComment        = parse.StringFrom(parse.String("<!--"), parse.StringUntil(jsEndOfSingleLineComment), jsEndOfSingleLineComment)
	// jsHTMLCloseComment is a comment in a classic script only at the start of a
	// line. Elsewhere, `a --> b` is `a-- > b`.
	jsHTMLCloseComment = parse.StringFrom(parse.String("-->"), parse.StringUntil(jsEndOfSingleLineComment), jsEndOfSingleLineComment)
	// jsHashbangComment is a comment only at the start of the source.
	jsHashbangComment = parse.StringFrom(parse.String("#!"), parse.StringUntil(jsEndOfSingleLineComment), jsEndOfSingleLineComment)
)

var (
	jsStartMultiLineComment = parse.String("/*")
	jsEndOfMultiLineComment = parse.StringFrom(parse.Or(parse.String("*/"), parse.EOF[string]()))
	jsMultiLineComment      = parse.StringFrom(jsStartMultiLineComment, parse.StringUntil(jsEndOfMultiLineComment), jsEndOfMultiLineComment, parse.OptionalWhitespace)
)

func isJSLineTerminatorRune(r rune) bool {
	return r == '\r' || r == '\u2028' || r == '\u2029'
}
