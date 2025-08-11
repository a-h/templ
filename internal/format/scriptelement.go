package format

import (
	"strings"

	"github.com/a-h/templ/internal/prettier"
	"github.com/a-h/templ/parser/v2"
)

const templScriptPlaceholder = "templ_go_expression_7331"

// ScriptElement formats a ScriptElement node, replacing Go expressions with placeholders for formatting.
// After formatting, it updates the GoCode expressions and their ranges.
func ScriptElement(se *parser.ScriptElement, depth int, prettierCommand string) (err error) {
	// Skip empty script elements, as they don't need formatting.
	if len(se.Contents) == 0 {
		return nil
	}

	// ScriptElements may contain Go expressions in {{ }} blocks. Prettier has no idea how to handle
	// that, so we replace them with a placeholder, format the script, and then replace the placeholders
	// with the original Go expressions.
	var placeholderContent []parser.ScriptContents
	var scriptWithPlaceholders strings.Builder
	for _, part := range se.Contents {
		if part.Value != nil {
			scriptWithPlaceholders.WriteString(*part.Value)
			continue
		}
		if part.GoCode != nil {
			scriptWithPlaceholders.WriteString(templScriptPlaceholder)
			placeholderContent = append(placeholderContent, part)
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
	after, err := prettier.Element("script", scriptType, scriptWithPlaceholders.String(), depth, prettierCommand)
	if err != nil {
		return err
	}

	// If there were no placeholders, set the contents to the formatted script.
	if len(placeholderContent) == 0 {
		se.Contents = []parser.ScriptContents{
			{
				Value:               &after,
				GoCode:              nil,
				InsideStringLiteral: false,
			},
		}
		return nil
	}

	split := strings.Split(after, templScriptPlaceholder)
	var appliedPlaceholderCount int
	var newContents []parser.ScriptContents
	for _, part := range split {
		if part == "" {
			continue
		}
		newContents = append(newContents, parser.ScriptContents{
			Value:               &part,
			GoCode:              nil,
			InsideStringLiteral: false,
		})
		// If we had a GoCode part, we need to add it back in.
		if appliedPlaceholderCount < len(placeholderContent) {
			placeholder := placeholderContent[appliedPlaceholderCount]
			// Trim horizontal space from the GoCode trailing space, as prettier will add its own.
			if placeholder.GoCode != nil && strings.Contains(string(placeholder.GoCode.TrailingSpace), " ") {
				placeholder.GoCode.TrailingSpace = parser.SpaceNone
			}
			newContents = append(newContents, placeholder)
			appliedPlaceholderCount++
		}
	}
	se.Contents = newContents

	return nil
}
