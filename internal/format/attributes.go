package format

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/a-h/templ/internal/htmlfind"
	"github.com/a-h/templ/internal/prettier"
	parser "github.com/a-h/templ/parser/v2"
	"github.com/a-h/templ/parser/v2/visitor"
	nethtml "golang.org/x/net/html"
)

// templIDMatcher matches <div> elements that have a data-templ-id attribute.
var templIDMatcher = func(n *nethtml.Node) bool {
	if n.Type != nethtml.ElementNode || n.Data != "div" {
		return false
	}
	for _, a := range n.Attr {
		if a.Key == "data-templ-id" {
			return true
		}
	}
	return false
}

type constantAttributeEntry struct {
	attr *parser.ConstantAttribute
	key  string
}

// Attributes sends constant attribute values through prettier so that plugins
// like prettier-plugin-tailwindcss can process them. Each attribute becomes a
// separate synthetic <div> element to avoid duplicate-key issues.
func Attributes(children []parser.Node, prettierCommand string) error {
	var entries []constantAttributeEntry
	collector := visitor.New()
	collector.ConstantAttribute = func(n *parser.ConstantAttribute) error {
		key, ok := n.Key.(parser.ConstantAttributeKey)
		if !ok {
			return nil
		}
		entries = append(entries, constantAttributeEntry{attr: n, key: key.Name})
		return nil
	}
	for _, child := range children {
		if err := child.Visit(collector); err != nil {
			return err
		}
	}
	if len(entries) == 0 {
		return nil
	}

	// Build synthetic HTML: one <div> per attribute.
	var sb strings.Builder
	for i, e := range entries {
		sb.WriteString(fmt.Sprintf(`<div data-templ-id="%d" %s="%s"></div>`, i, e.key, html.EscapeString(e.attr.Value)))
		sb.WriteByte('\n')
	}

	formatted, err := prettier.Run(sb.String(), "templ_content.html", prettierCommand)
	if err != nil {
		return fmt.Errorf("prettier attribute formatting error: %w", err)
	}

	// Parse the formatted output and read back attribute values.
	doc, err := nethtml.Parse(strings.NewReader(formatted))
	if err != nil {
		return fmt.Errorf("failed to parse prettier output for attributes: %w", err)
	}

	nodes := htmlfind.All(doc, templIDMatcher)
	for _, node := range nodes {
		idVal, ok := getAttr(node, "data-templ-id")
		if !ok {
			return fmt.Errorf("matched element missing data-templ-id attribute")
		}
		id, err := strconv.Atoi(idVal)
		if err != nil {
			return fmt.Errorf("failed to parse data-templ-id: %w", err)
		}
		if id < 0 || id >= len(entries) {
			return fmt.Errorf("data-templ-id %d out of range [0, %d)", id, len(entries))
		}
		e := entries[id]
		val, ok := getAttr(node, e.key)
		if !ok {
			return fmt.Errorf("prettier removed attribute %q from element with data-templ-id %d", e.key, id)
		}
		e.attr.Value = val
		e.attr.SingleQuote = strings.Contains(val, `"`)
	}

	return nil
}

func getAttr(n *nethtml.Node, key string) (val string, ok bool) {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val, true
		}
	}
	return "", false
}
