package parser

import (
	"strings"

	"github.com/a-h/parse"
)

var scriptElement = scriptElementParser{}

type jsQuote string

const (
	jsQuoteNone     jsQuote = ""
	jsQuoteSingle   jsQuote = `'`
	jsQuoteDouble   jsQuote = `"`
	jsQuoteBacktick jsQuote = "`"
)

type scriptElementParser struct{}

func (p scriptElementParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	start := pi.Index()

	// <
	if _, ok, err = lt.Parse(pi); err != nil || !ok {
		return
	}

	// Element name.
	var e ScriptElement
	var name string
	if name, ok, err = elementNameParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	if name != "script" {
		pi.Seek(start)
		ok = false
		return
	}

	if e.Attributes, ok, err = (attributesParser{}).Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Optional whitespace.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		pi.Seek(start)
		return
	}

	// >
	if _, ok, err = gt.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Parse the contents, we should get script text or Go expressions up until the closing tag.
	var sb strings.Builder
	var stringLiteralDelimiter jsQuote

loop:
	for {
		// Read and decide whether we're we've hit a:
		//  - {{ - Start of a Go expression, read the contents with the `goCode` function.
		//  - </script> - End of the script, break out of the loop.
		//  - ' - Start of a single quoted string.
		//  - " - Start of a double quoted string.
		//  - ` - Start of a backtick quoted string.
		//  - // - Start of a single line comment - can read to the end of the line without parsing.
		//  - /* - Start of a multi-line comment - can read to the end of the comment without parsing.
		//  - \ - Start of an escape sequence, we can just take the value.
		//  - Anything else - Add it to the script.

		if _, ok, err = jsEndTag.Parse(pi); err != nil || ok {
			// We've reached the end of the script.
			break loop
		}

		if _, ok, err = endTagStart.Parse(pi); err != nil || ok {
			// We've reached the end of the script, but the end tag is probably invalid.
			break loop
		}

		var code Node
		code, ok, err = goCodeInJavaScript.Parse(pi)
		if err != nil {
			return nil, false, err
		}
		if ok {
			e.Contents = append(e.Contents, NewScriptContentsGo(code.(GoCode), stringLiteralDelimiter != jsQuoteNone))
			continue loop
		}

		// Try for a comment.
		var comment string
		comment, ok, err = jsComment.Parse(pi)
		if err != nil {
			return nil, false, err
		}
		if ok {
			e.Contents = append(e.Contents, NewScriptContentsJS(comment))
			continue loop
		}

		// Read JavaScript chracaters.
		for {
			before := pi.Index()
			var c string
			c, ok, err := jsCharacter.Parse(pi)
			if err != nil {
				return nil, false, err
			}
			if ok {
				_, isEOF, _ := parse.EOF[string]().Parse(pi)
				if c == `"` || c == "'" || c == "`" {
					// Start or exit a string literal.
					if stringLiteralDelimiter == jsQuoteNone {
						stringLiteralDelimiter = jsQuote(c)
					} else if stringLiteralDelimiter == jsQuote(c) {
						stringLiteralDelimiter = jsQuoteNone
					}
				}
				peeked, _ := pi.Peek(1)
				peeked = c + peeked

				breakForGo := peeked == "{{"
				breakForHTML := stringLiteralDelimiter == jsQuoteNone && (peeked == "</" || peeked == "//" || peeked == "/*")

				if isEOF || breakForGo || breakForHTML {
					if sb.Len() > 0 {
						e.Contents = append(e.Contents, NewScriptContentsJS(sb.String()))
						sb.Reset()
					}
					if isEOF {
						break loop
					}
					pi.Seek(before)
					continue loop
				}
				sb.WriteString(c)
			}
			if _, ok, _ = parse.EOF[string]().Parse(pi); ok {
				return nil, false, parse.Error("script: unclosed <script> element", pi.Position())
			}
		}
	}

	return e, true, nil
}

var jsEndTag = parse.String("</script>")
var endTagStart = parse.String("</")

var jsCharacter = parse.Any(jsEscapedCharacter, parse.AnyRune)

// \uXXXX	Unicode code point escape	'\u0061' = 'a'
var hexDigit = parse.Any(parse.ZeroToNine, parse.RuneIn("abcdef"), parse.RuneIn("ABCDEF"))
var jsUnicodeEscape = parse.StringFrom(parse.String("\\u"), hexDigit, hexDigit, hexDigit, hexDigit)

// \u{X...}	ES6+ extended Unicode escape	'\u{1F600}' = '😀'
var jsExtendedUnicodeEscape = parse.StringFrom(parse.String("\\u{"), hexDigit, parse.StringFrom(parse.AtLeast(1, parse.ZeroOrMore(hexDigit))), parse.String("}"))

// \xXX	Hex code (2-digit)	'\x41' = 'A'
var jsHexEscape = parse.StringFrom(parse.String("\\x"), hexDigit, hexDigit)

// \x Backslash escape	'\\' = '\'
var jsBackslashEscape = parse.StringFrom(parse.String("\\"), parse.AnyRune)

// All escapes.
var jsEscapedCharacter = parse.Any(jsBackslashEscape, jsUnicodeEscape, jsHexEscape, jsExtendedUnicodeEscape)

var jsComment = parse.Any(jsSingleLineComment, jsMultiLineComment)

var jsStartSingleLineComment = parse.String("//")
var jsEndOfSingleLineComment = parse.StringFrom(parse.Or(parse.NewLine, parse.EOF[string]()))
var jsSingleLineComment = parse.StringFrom(jsStartSingleLineComment, parse.StringUntil(jsEndOfSingleLineComment), jsEndOfSingleLineComment)

var jsStartMultiLineComment = parse.String("/*")
var jsEndOfMultiLineComment = parse.StringFrom(parse.Or(parse.String("*/"), parse.EOF[string]()))
var jsMultiLineComment = parse.StringFrom(jsStartMultiLineComment, parse.StringUntil(jsEndOfMultiLineComment), jsEndOfMultiLineComment, parse.OptionalWhitespace)
