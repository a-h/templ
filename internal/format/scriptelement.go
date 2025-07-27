package format

import (
	"fmt"
	"strings"

	"github.com/a-h/templ/internal/htmlfind"
	"github.com/a-h/templ/internal/prettier"
	"github.com/a-h/templ/parser/v2"
	"golang.org/x/net/html"
)

// ScriptElement formats a ScriptElement node, replacing Go expressions with placeholders for formatting.
// After formatting, it updates the GoCode expressions and their ranges.
func ScriptElement(se *parser.ScriptElement, depth int) (err error) {
	// Skip empty script elements, as they don't need formatting.
	if len(se.Contents) == 0 {
		return nil
	}

	var scriptWithPlaceholders strings.Builder

	// Add divs to the start and end of the script to ensure that prettier formats the content with
	// correct indentation.
	for i := range depth + 1 {
		scriptWithPlaceholders.WriteString(fmt.Sprintf("<div data-templ-depth=\"%d\">", i))
	}

	se.WriteStart(&scriptWithPlaceholders, depth)
	for _, part := range se.Contents {
		if part.Value != nil {
			scriptWithPlaceholders.WriteString(*part.Value)
			continue
		}
		if part.GoCode != nil {
			scriptWithPlaceholders.WriteString("__templ_go_expr_()")
			continue
		}
	}
	se.WriteEnd(&scriptWithPlaceholders)

	for range depth + 1 {
		scriptWithPlaceholders.WriteString("</div>")
	}

	before := scriptWithPlaceholders.String()
	after, err := prettier.Run(before, "script_element.html")
	if err != nil {
		return fmt.Errorf("prettier error: %w", err)
	}
	if before == after {
		return nil
	}

	// Chop off the start and end divs we added to get prettier to format the content with correct
	// indentation.
	if depth > 0 {
		matcher := htmlfind.Element("script")
		nodes, err := htmlfind.AllReader(strings.NewReader(after), matcher)
		if err != nil {
			return fmt.Errorf("htmlfind error: %w", err)
		}
		if len(nodes) != 1 {
			return fmt.Errorf("expected 1 script node, got %d", len(nodes))
		}
		scriptNode := nodes[0]
		if scriptNode.FirstChild == nil {
			return fmt.Errorf("script node has no children")
		}
		var sb strings.Builder
		for node := range scriptNode.ChildNodes() {
			if node.Type == html.TextNode {
				sb.WriteString(node.Data)
			}
		}
		after = strings.TrimRight(sb.String(), " \t\r\n") + "\n" + strings.Repeat("\t", depth)
		fmt.Printf("After: %s\n", showWhitespace(after))
	}

	split := strings.Split(after, "__templ_go_expr_()")
	for i, part := range se.Contents {
		if part.Value != nil {
			se.Contents[i].Value = &split[i*2]
		}
	}

	return nil
}

func showWhitespace(s string) string {
	s = strings.ReplaceAll(s, "\n", "⏎\n")
	s = strings.ReplaceAll(s, "\t", "→")
	s = strings.ReplaceAll(s, " ", "·")
	return s
}
