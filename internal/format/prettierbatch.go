package format

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/a-h/templ/internal/prettier"
	parser "github.com/a-h/templ/parser/v2"
	"github.com/a-h/templ/parser/v2/visitor"
	nethtml "golang.org/x/net/html"
)

const (
	attributeIDKey = "data-templ-attribute-id"
	scriptIDKey    = "data-templ-script-id"
	styleIDKey     = "data-templ-style-id"
)

const templScriptPlaceholder = "templ_go_expression_7331"

// prettierBatch collects the constant attributes, script elements, and style elements of a
// templ file so that a single prettier process formats all of them. Starting prettier takes
// over 100ms, so starting it once per template or element makes formatting slow.
type prettierBatch struct {
	command    string
	attributes []constantAttributeEntry
	scripts    []scriptElementEntry
	styles     []styleElementEntry
}

// constantAttributeEntry is a constant attribute whose value prettier formats, so that
// plugins like prettier-plugin-tailwindcss can process it.
type constantAttributeEntry struct {
	attr *parser.ConstantAttribute
	key  string
}

func newConstantAttributeEntry(attr *parser.ConstantAttribute, key string) constantAttributeEntry {
	return constantAttributeEntry{attr: attr, key: key}
}

// Apply sets the attribute value to the formatted value.
func (e constantAttributeEntry) Apply(val string) {
	e.attr.Value = val
	e.attr.SingleQuote = strings.Contains(val, `"`)
}

// scriptElementEntry is a script element whose contents prettier formats.
//
// ScriptElements may contain Go expressions in {{ }} blocks. Prettier has no idea how to handle
// that, so the content replaces them with a placeholder, and Apply replaces the placeholders
// with the original Go expressions.
type scriptElementEntry struct {
	se           *parser.ScriptElement
	depth        int
	content      string
	scriptType   string
	placeholders []parser.ScriptContents
}

func newScriptElementEntry(se *parser.ScriptElement, depth int) scriptElementEntry {
	e := scriptElementEntry{se: se, depth: depth}

	var content strings.Builder
	for _, part := range se.Contents {
		if part.Value != nil {
			content.WriteString(*part.Value)
			continue
		}
		if part.GoCode != nil {
			content.WriteString(templScriptPlaceholder)
			e.placeholders = append(e.placeholders, part)
			continue
		}
	}
	e.content = content.String()

	// <script> elements can have a type attribute that specifies the script type. This is important
	// to avoid formatting errors, e.g. attempting to format <script type="text/hyperscript"> with a
	// JavaScript formatter.
loop:
	for _, attr := range se.Attributes {
		switch attr := attr.(type) {
		case *parser.ConstantAttribute:
			if attr.Key.String() == "type" {
				e.scriptType = attr.Value
			}
			break loop
		}
	}

	return e
}

// Apply sets the script element contents to the formatted contents, and restores the Go
// expressions.
func (e scriptElementEntry) Apply(after string) {
	if len(e.placeholders) == 0 {
		e.se.Contents = []parser.ScriptContents{
			{
				Value:  &after,
				GoCode: nil,
			},
		}
		return
	}

	split := strings.Split(after, templScriptPlaceholder)
	var appliedPlaceholderCount int
	var newContents []parser.ScriptContents
	for _, part := range split {
		if part == "" {
			continue
		}
		newContents = append(newContents, parser.ScriptContents{
			Value:  &part,
			GoCode: nil,
		})
		// If we had a GoCode part, we need to add it back in.
		if appliedPlaceholderCount < len(e.placeholders) {
			placeholder := e.placeholders[appliedPlaceholderCount]
			// Trim horizontal space from the GoCode trailing space, as prettier will add its own.
			if placeholder.GoCode != nil && strings.Contains(string(placeholder.GoCode.TrailingSpace), " ") {
				placeholder.GoCode.TrailingSpace = parser.SpaceNone
			}
			newContents = append(newContents, placeholder)
			appliedPlaceholderCount++
		}
	}
	e.se.Contents = newContents
}

// styleElementEntry is a style element whose contents prettier formats.
type styleElementEntry struct {
	re    *parser.RawElement
	depth int
}

func newStyleElementEntry(re *parser.RawElement, depth int) styleElementEntry {
	return styleElementEntry{re: re, depth: depth}
}

// Apply sets the style element contents to the formatted contents.
func (e styleElementEntry) Apply(after string) {
	e.re.Contents = after
}

func newPrettierBatch(command string) *prettierBatch {
	return &prettierBatch{command: command}
}

// AddAttributes adds the constant attributes within the children to the batch.
func (b *prettierBatch) AddAttributes(children []parser.Node) error {
	collector := visitor.New()
	collector.ConstantAttribute = func(n *parser.ConstantAttribute) error {
		key, ok := n.Key.(parser.ConstantAttributeKey)
		if !ok {
			return nil
		}
		b.attributes = append(b.attributes, newConstantAttributeEntry(n, key.Name))
		return nil
	}
	for _, child := range children {
		if err := child.Visit(collector); err != nil {
			return err
		}
	}
	return nil
}

// AddScriptElement adds the contents of a script element to the batch. The depth is the
// number of elements the script element is nested within, and sets its indentation.
func (b *prettierBatch) AddScriptElement(se *parser.ScriptElement, depth int) {
	if len(se.Contents) == 0 {
		return
	}
	b.scripts = append(b.scripts, newScriptElementEntry(se, depth))
}

// AddStyleElement adds the contents of a style element to the batch. The depth is the
// number of elements the style element is nested within, and sets its indentation.
func (b *prettierBatch) AddStyleElement(re *parser.RawElement, depth int) {
	if re.Name != "style" || len(re.Contents) == 0 {
		return
	}
	b.styles = append(b.styles, newStyleElementEntry(re, depth))
}

// Run formats everything in the batch with a single prettier process, and updates the
// templ nodes with the result.
func (b *prettierBatch) Run() error {
	if len(b.attributes) == 0 && len(b.scripts) == 0 && len(b.styles) == 0 {
		return nil
	}

	// Each attribute becomes a separate element to avoid duplicate-key issues.
	var sb strings.Builder
	for i, e := range b.attributes {
		fmt.Fprintf(&sb, `<div %s="%d" %s="%s"></div>`, attributeIDKey, i, e.key, html.EscapeString(e.attr.Value))
		sb.WriteByte('\n')
	}
	for i, e := range b.scripts {
		writeElementHTML(&sb, "script", scriptIDKey, i, e.scriptType, e.content, e.depth)
	}
	for i, e := range b.styles {
		writeElementHTML(&sb, "style", styleIDKey, i, "text/css", e.re.Contents, e.depth)
	}
	input := sb.String()

	formatted, err := prettier.Run(input, "templ_content.html", b.command)
	if err != nil {
		return fmt.Errorf("prettier error: %w", err)
	}
	doc, err := nethtml.Parse(strings.NewReader(formatted))
	if err != nil {
		return fmt.Errorf("failed to parse prettier output: %w", err)
	}

	attributeNodes, err := findEntryNodes(doc, attributeIDKey, len(b.attributes))
	if err != nil {
		return err
	}
	for i, node := range attributeNodes {
		e := b.attributes[i]
		val, ok := getAttr(node, e.key)
		if !ok {
			return fmt.Errorf("prettier removed attribute %q from element with %s %d", e.key, attributeIDKey, i)
		}
		e.Apply(val)
	}

	// A command that passes the input through unchanged, such as cat, leaves the element
	// contents as written.
	if formatted == input {
		return nil
	}

	scriptNodes, err := findEntryNodes(doc, scriptIDKey, len(b.scripts))
	if err != nil {
		return err
	}
	for i, node := range scriptNodes {
		after, err := readElementContents(node, b.scripts[i].depth)
		if err != nil {
			return fmt.Errorf("script element: %w", err)
		}
		b.scripts[i].Apply(after)
	}

	styleNodes, err := findEntryNodes(doc, styleIDKey, len(b.styles))
	if err != nil {
		return err
	}
	for i, node := range styleNodes {
		after, err := readElementContents(node, b.styles[i].depth)
		if err != nil {
			return fmt.Errorf("style element: %w", err)
		}
		b.styles[i].Apply(after)
	}

	return nil
}

// writeElementHTML writes an element wrapped in depth div elements, so that prettier indents
// its contents to match its position in the templ file.
func writeElementHTML(sb *strings.Builder, name, idKey string, entryIndex int, typeAttrValue, content string, depth int) {
	for i := range depth {
		fmt.Fprintf(sb, "<div data-templ-depth=\"%d\">", i)
	}
	fmt.Fprintf(sb, "<%s %s=\"%d\"", name, idKey, entryIndex)
	if typeAttrValue != "" {
		fmt.Fprintf(sb, " type=\"%s\"", html.EscapeString(typeAttrValue))
	}
	sb.WriteString(">")
	sb.WriteString(content)
	fmt.Fprintf(sb, "</%s>", name)
	for range depth {
		sb.WriteString("</div>")
	}
	sb.WriteByte('\n')
}

// readElementContents returns the text contents of a formatted element, with the trailing
// whitespace set so that the closing tag is indented to the depth.
func readElementContents(n *nethtml.Node, depth int) (contents string, err error) {
	if n.FirstChild == nil {
		return "", fmt.Errorf("%q node has no children", n.Data)
	}
	var sb strings.Builder
	for child := range n.ChildNodes() {
		sb.WriteString(child.Data)
	}
	return strings.TrimRight(sb.String(), " \t\r\n") + "\n" + strings.Repeat("\t", depth), nil
}

// findEntryNodes returns the element in the prettier output for each of the entryCount batch
// entries of one kind. Run writes the position of each entry in its batch slice into the idKey
// attribute, so entryNodes[i] is the element whose idKey attribute is i. It returns an error if
// prettier dropped or duplicated an element.
func findEntryNodes(doc *nethtml.Node, idKey string, entryCount int) (entryNodes []*nethtml.Node, err error) {
	entryNodes = make([]*nethtml.Node, entryCount)
	for n := range doc.Descendants() {
		if n.Type != nethtml.ElementNode {
			continue
		}
		val, ok := getAttr(n, idKey)
		if !ok {
			continue
		}
		entryIndex, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", idKey, err)
		}
		if entryIndex < 0 || entryIndex >= entryCount {
			return nil, fmt.Errorf("%s %d out of range [0, %d)", idKey, entryIndex, entryCount)
		}
		if entryNodes[entryIndex] != nil {
			return nil, fmt.Errorf("prettier output contains %s %d more than once", idKey, entryIndex)
		}
		entryNodes[entryIndex] = n
	}
	for i, n := range entryNodes {
		if n == nil {
			return nil, fmt.Errorf("prettier output is missing the element with %s %d", idKey, i)
		}
	}
	return entryNodes, nil
}

// getAttr returns the value of the attribute with the given key. The comparison
// is case-insensitive because golang.org/x/net/html lowercases attribute keys
// during parsing.
func getAttr(n *nethtml.Node, key string) (val string, ok bool) {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val, true
		}
	}
	return "", false
}
