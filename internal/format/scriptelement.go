package format

import (
	"strings"

	"github.com/a-h/templ/parser/v2"
)

const templScriptPlaceholder = "templ_go_expression_7331"

// ScriptElement formats a ScriptElement node, replacing Go expressions with placeholders for formatting.
// After formatting, it updates the GoCode expressions and their ranges.
func ScriptElement(se *parser.ScriptElement, depth int) (err error) {
	// Skip empty script elements, as they don't need formatting.
	if len(se.Contents) == 0 {
		return nil
	}

	// ScriptElements may contain Go expressions in {{ }} blocks. Prettier has no idea how to handle
	// that, so we replace them with a placeholder, format the script, and then replace the placeholders
	// with the original Go expressions.
	var scriptWithPlaceholders strings.Builder
	for _, part := range se.Contents {
		if part.Value != nil {
			scriptWithPlaceholders.WriteString(*part.Value)
			continue
		}
		if part.GoCode != nil {
			scriptWithPlaceholders.WriteString(templScriptPlaceholder)
			continue
		}
	}

	// <script> elements can have a type attribute that specifies the script type. This is important
	// to avoid formatting errors, e.g. attempting to format <script type="text/hyperscript"> with a
	// JavaScript formatter.
	var scriptType string
loop:
	for _, attr := range se.Attributes {
		switch attr := attr.(type) {
		case *parser.ConstantAttribute:
			if attr.Key.String() == "type" {
				scriptType = attr.Value
			}
			break loop
		}
	}

	// Use the prettifyElement function to format the script contents.
	after, err := prettifyElement("script", scriptType, scriptWithPlaceholders.String(), depth)
	if err != nil {
		return err
	}

	// After formatting, replace the placeholders with the original Go expressions.
	split := strings.Split(after, templScriptPlaceholder)
	var splitIndex int
	for i, part := range se.Contents {
		if part.Value != nil {
			se.Contents[i].Value = &split[splitIndex]
			splitIndex++
		}
		if part.GoCode != nil && part.GoCode.TrailingSpace != parser.SpaceNone {
			se.Contents[i].GoCode.TrailingSpace = parser.SpaceNone
		}
	}

	return nil
}
