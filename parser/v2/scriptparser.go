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
	e := &ScriptElement{}
	var name string
	if name, ok, err = elementNameParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return n, false, err
	}

	if name != "script" {
		pi.Seek(start)
		return n, false, nil
	}

	if e.Attributes, ok, err = (attributesParser{}).Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return n, false, err
	}

	// Optional whitespace.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		pi.Seek(start)
		return n, false, err
	}

	// >
	if _, ok, err = gt.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return n, false, parse.Error("<script>: unclosed element - missing '>'", pi.Position())
	}

	// If there's a type attribute and it's not a JS attribute (e.g. text/javascript), we need to parse the contents as raw text.
	if !hasJavaScriptType(e.Attributes) {
		var contents string
		if contents, ok, err = parse.StringUntil(jsEndTag).Parse(pi); err != nil || !ok {
			return e, true, parse.Error("<script>: expected end tag not present", pi.Position())
		}
		e.Contents = append(e.Contents, NewScriptContentsScriptCode(contents))

		// Cut the end element.
		_, _, _ = jsEndTag.Parse(pi)

		return e, true, nil
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

		_, ok, err = jsEndTag.Parse(pi)
		if err != nil {
			return nil, false, err
		}
		if ok {
			// We've reached the end of the script.
			break loop
		}

		_, ok, err = endTagStart.Parse(pi)
		if err != nil {
			return nil, false, err
		}
		if ok {
			return nil, false, parse.Error("<script>: invalid end tag, expected </script> not found", pi.Position())
		}

		// Try for a Go code expression, i.e. {{ goCode }}.
		code, ok, err := goCodeInJavaScript.Parse(pi)
		if err != nil {
			return nil, false, err
		}
		if ok {
			e.Contents = append(e.Contents, NewScriptContentsGo(code.(*GoCode), stringLiteralDelimiter != jsQuoteNone))
			continue loop
		}

		// Try for a comment.
		comment, ok, err := jsComment.Parse(pi)
		if err != nil {
			return nil, false, err
		}
		if ok {
			e.Contents = append(e.Contents, NewScriptContentsScriptCode(comment))
			continue loop
		}

		// Read JavaScript characters.
	charLoop:
		for {
			before := pi.Index()

			// If we're outside of a string literal, check for a regexp literal.
			// Check for a regular expression literal.
			if stringLiteralDelimiter == jsQuoteNone {
				r, ok, err := regexpLiteral.Parse(pi)
				if err != nil {
					return nil, false, err
				}
				if ok {
					sb.WriteString(r)
					continue charLoop
				}
			}

			// Check for EOF.
			if _, ok, _ = parse.EOF[string]().Parse(pi); ok {
				return nil, false, parse.Error("script: unclosed <script> element", pi.Position())
			}

			// Check for a character.
			c, ok, err := jsCharacter.Parse(pi)
			if err != nil {
				return nil, false, err
			}
			if !ok {
				return nil, false, parse.Error("script: expected to parse a character, but didn't", pi.Position())
			}
			if c == string(jsQuoteDouble) || c == string(jsQuoteSingle) || c == string(jsQuoteBacktick) {
				// Start or exit a string literal.
				if stringLiteralDelimiter == jsQuoteNone {
					stringLiteralDelimiter = jsQuote(c)
				} else if stringLiteralDelimiter == jsQuote(c) {
					stringLiteralDelimiter = jsQuoteNone
				}
			}

			peeked, peekOK := pi.Peek(1)
			isEOF := !peekOK
			peeked = c + peeked
			breakForGo := peeked == "{{"
			breakForHTML := stringLiteralDelimiter == jsQuoteNone && peeked == "</"
			breakForComment := stringLiteralDelimiter == jsQuoteNone && (peeked == "//" || peeked == "/*")
			if isEOF || breakForGo || breakForHTML || breakForComment {
				if sb.Len() > 0 {
					e.Contents = append(e.Contents, NewScriptContentsScriptCode(sb.String()))
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
	}

	e.Range = NewRange(pi.PositionAt(start), pi.Position())
	return e, true, nil
}

var javaScriptTypeAttributeValues = []string{
	"", // If the type is not set, it is JavaScript.
	"text/javascript",
	"javascript", // Obsolete, but still used.
	"module",
}

func hasJavaScriptType(attrs []Attribute) bool {
	for _, attr := range attrs {
		ca, isCA := attr.(*ConstantAttribute)
		if !isCA {
			continue
		}
		caKey, isCAKey := ca.Key.(ConstantAttributeKey)
		if !isCAKey {
			continue
		}
		if !strings.EqualFold(caKey.Name, "type") {
			continue
		}
		for _, v := range javaScriptTypeAttributeValues {
			if strings.EqualFold(ca.Value, v) {
				return true
			}
		}
		// If there's a type attribute but the value doesn't match any
		// known JavaScript type, it's not JavaScript.
		return false
	}
	// If there's no type attribute, it's JavaScript.
	return true
}

var (
	jsEndTag    = parse.String("</script>")
	endTagStart = parse.String("</")
)

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
