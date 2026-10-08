package parser

import (
	"strings"

	"github.com/a-h/parse"
	"github.com/a-h/templ/internal/jstokenizer"
)

var scriptElement = scriptElementParser{}

var jsTokenizerContextToScriptContentsContext = map[jstokenizer.Context]ScriptContentsContext{
	jstokenizer.ContextCode:            ScriptContentsContextExpression,
	jstokenizer.ContextString:          ScriptContentsContextString,
	jstokenizer.ContextTemplateLiteral: ScriptContentsContextTemplateLiteral,
}

type scriptElementParser struct{}

func (p scriptElementParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	start := pi.Index()

	// <
	if _, ok, err = lt.Parse(pi); err != nil || !ok {
		return
	}
	openTagStart := pi.PositionAt(start)

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
	e.OpenTagRange = NewRange(openTagStart, pi.Position())

	// If there's a type attribute and it's not a JS attribute (e.g. text/javascript), we need to parse the contents as raw text.
	if !hasJavaScriptType(e.Attributes) {
		var contents string
		if contents, ok, err = parse.StringUntil(jsEndTag).Parse(pi); err != nil || !ok {
			return e, true, parse.Error("<script>: expected end tag not present", pi.Position())
		}
		e.Contents = append(e.Contents, NewScriptContentsScriptCode(contents))

		// Cut the end element.
		closeTagStart := pi.Position()
		_, _, _ = jsEndTag.Parse(pi)
		e.CloseTagRange = NewRange(closeTagStart, pi.Position())
		e.Range = NewRange(pi.PositionAt(start), pi.Position())
		return e, true, nil
	}

	// Parse the contents, we should get script text or Go expressions up until the closing tag.
	var sb strings.Builder
	tokenizer := jstokenizer.New()
	for {
		closeTagStartIndex := pi.Index()
		_, ok, err = jsEndTag.Parse(pi)
		if err != nil {
			return nil, false, err
		}
		if ok {
			e.CloseTagRange = NewRange(pi.PositionAt(closeTagStartIndex), pi.Position())
			break
		}
		pi.Seek(closeTagStartIndex)

		if tokenizer.GetContext() == jstokenizer.ContextCode {
			_, ok, err = endTagStart.Parse(pi)
			if err != nil {
				return nil, false, err
			}
			if ok {
				return nil, false, parse.Error("<script>: invalid end tag, expected </script> not found", pi.Position())
			}
		}

		code, ok, err := goCodeInJavaScript.Parse(pi)
		if err != nil {
			return nil, false, err
		}
		if ok {
			if sb.Len() > 0 {
				e.Contents = append(e.Contents, NewScriptContentsScriptCode(sb.String()))
				sb.Reset()
			}
			context := jsTokenizerContextToScriptContentsContext[tokenizer.GetContext()]
			e.Contents = append(e.Contents, NewScriptContentsGo(code.(*GoCode), context))
			tokenizer.RecordValue()
			continue
		}

		comment, ok, err := tokenizer.ReadComment(pi)
		if err != nil {
			return nil, false, err
		}
		if ok {
			if sb.Len() > 0 {
				e.Contents = append(e.Contents, NewScriptContentsScriptCode(sb.String()))
				sb.Reset()
			}
			e.Contents = append(e.Contents, NewScriptContentsScriptCode(comment))
			continue
		}

		token, ok, err := tokenizer.Read(pi)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return nil, false, parse.Error("script: unclosed <script> element", pi.Position())
		}
		sb.WriteString(token)
	}
	if sb.Len() > 0 {
		e.Contents = append(e.Contents, NewScriptContentsScriptCode(sb.String()))
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
