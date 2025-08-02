package format

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	"github.com/a-h/templ/internal/htmlfind"
	"github.com/a-h/templ/internal/imports"
	"github.com/a-h/templ/internal/prettier"
	parser "github.com/a-h/templ/parser/v2"
	"github.com/a-h/templ/parser/v2/visitor"
)

func calculateNodeDepth(e parser.Node, nodeToDepth map[parser.Node]int, depth int) {
	switch e := e.(type) {
	case *parser.ScriptElement:
		nodeToDepth[e] = depth
	case *parser.RawElement:
		nodeToDepth[e] = depth
	case parser.CompositeNode:
		for _, child := range e.ChildNodes() {
			calculateNodeDepth(child, nodeToDepth, depth+1)
		}
	}
}

// Templ formats templ source, returning the formatted output, whether it changed, and an error if any.
// The fileName is used for Go import processing, use an empty name if the source is not from a file.
func Templ(src []byte, fileName string) (output []byte, changed bool, err error) {
	t, err := parser.ParseString(string(src))
	if err != nil {
		return nil, false, err
	}
	t.Filepath = fileName
	t, err = imports.Process(t)
	if err != nil {
		return nil, false, err
	}

	nodeFormatter := visitor.New()
	// Calculate the depth of each ScriptElement and RawElement in the tree so that the formatting is properly indented.
	nodeToDepth := make(map[parser.Node]int)
	nodeFormatter.HTMLTemplate = func(n *parser.HTMLTemplate) error {
		// Visit the children first to calculate their depth.
		for _, child := range n.Children {
			calculateNodeDepth(child, nodeToDepth, 1)
		}
		// Now that we have the depth of each node, we can format them.
		for _, child := range n.Children {
			if err := child.Visit(nodeFormatter); err != nil {
				return err
			}
		}
		return nil
	}
	nodeFormatter.ScriptElement = func(se *parser.ScriptElement) error {
		depth := nodeToDepth[se]
		return ScriptElement(se, depth)
	}
	nodeFormatter.RawElement = func(re *parser.RawElement) error {
		depth := nodeToDepth[re]
		if re.Name != "style" {
			return nil
		}
		return StyleElement(re, depth)
	}

	if err = nodeFormatter.VisitTemplateFile(t); err != nil {
		return nil, false, err
	}

	w := new(bytes.Buffer)
	if err = t.Write(w); err != nil {
		return nil, false, fmt.Errorf("formatting error: %w", err)
	}
	out := w.Bytes()
	changed = !bytes.Equal(src, out)
	return out, changed, nil
}

func prettifyElement(name string, typeAttrValue string, content string, depth int) (after string, err error) {
	var indentationWrapper strings.Builder

	// Add divs to the start and end of the script to ensure that prettier formats the content with
	// correct indentation.
	for i := range depth {
		indentationWrapper.WriteString(fmt.Sprintf("<div data-templ-depth=\"%d\">", i))
	}

	// Write start tag with type attribute if present.
	indentationWrapper.WriteString("<")
	indentationWrapper.WriteString(name)
	if typeAttrValue != "" {
		indentationWrapper.WriteString(" type=\"")
		indentationWrapper.WriteString(html.EscapeString(typeAttrValue))
		indentationWrapper.WriteString("\"")
	}
	indentationWrapper.WriteString(">")

	// Write contents.
	indentationWrapper.WriteString(content)

	// Write end tag.
	indentationWrapper.WriteString("</")
	indentationWrapper.WriteString(name)
	indentationWrapper.WriteString(">")

	for range depth {
		indentationWrapper.WriteString("</div>")
	}

	before := indentationWrapper.String()
	after, err = prettier.Run(before, "templ_content.html")
	if err != nil {
		return "", fmt.Errorf("prettier error: %w", err)
	}
	if before == after {
		return before, nil
	}

	// Chop off the start and end divs we added to get prettier to format the content with correct
	// indentation.
	matcher := htmlfind.Element(name)
	nodes, err := htmlfind.AllReader(strings.NewReader(after), matcher)
	if err != nil {
		return before, fmt.Errorf("htmlfind error: %w", err)
	}
	if len(nodes) != 1 {
		return before, fmt.Errorf("expected 1 %q node, got %d", name, len(nodes))
	}
	scriptNode := nodes[0]
	if scriptNode.FirstChild == nil {
		return before, fmt.Errorf("%q node has no children", name)
	}
	var sb strings.Builder
	for node := range scriptNode.ChildNodes() {
		sb.WriteString(node.Data)
	}
	after = strings.TrimRight(sb.String(), " \t\r\n") + "\n" + strings.Repeat("\t", depth)

	return after, nil
}
