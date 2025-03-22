package parser

import (
	"strings"

	"github.com/a-h/parse"
)

var scriptElement = scriptElementParser{}

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
	var isInsideStringLiteral bool
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

		var code Node
		code, ok, err = goCodeInJavaScript.Parse(pi)
		if err != nil {
			return nil, false, err
		}
		if ok {
			e.Contents = append(e.Contents, NewScriptContentsGo(code.(GoCode), isInsideStringLiteral))
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
					isInsideStringLiteral = !isInsideStringLiteral
				}
				peeked, _ := pi.Peek(1)
				peeked = c + peeked
				if isEOF || peeked == "{{" || peeked == "</" || peeked == "//" || peeked == "/*" {
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

var jsCharacter = parse.Any(jsEscapedCharacter, parse.AnyRune)

var jsEscapedCharacter = parse.StringFrom(parse.String("\\"), parse.AnyRune)

var jsComment = parse.Any(jsSingleLineComment, jsMultiLineComment)

var jsStartSingleLineComment = parse.String("//")
var jsSingleLineComment = parse.StringFrom(jsStartSingleLineComment, parse.StringUntil(parse.NewLine), parse.NewLine)

var jsStartMultiLineComment = parse.String("/*")
var jsMultiLineComment = parse.StringFrom(jsStartMultiLineComment, parse.StringUntil(parse.String("*/")), parse.String("*/"), parse.OptionalWhitespace)
