package parser

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io"
	"strings"
	"unicode"

	"github.com/a-h/parse"
)

// package parser
//
// import "strings"
// import strs "strings"
//
// css AddressLineStyle() {
//   background-color: #ff0000;
//   color: #ffffff;
// }
//
// templ RenderAddress(addr Address) {
// 	<div style={ AddressLineStyle() }>{ addr.Address1 }</div>
// 	<div>{ addr.Address2 }</div>
// 	<div>{ addr.Address3 }</div>
// 	<div>{ addr.Address4 }</div>
// }
//
// templ Render(p Person) {
//    <div>
//      <div>{ p.Name() }</div>
//      <a href={ p.URL }>{ strings.ToUpper(p.Name()) }</a>
//      <div>
//          if p.Type == "test" {
//             <span>{ "Test user" }</span>
//          } else {
// 	    	<span>{ "Not test user" }</span>
//          }
//          for _, v := range p.Addresses {
//             {! call RenderAddress(v) }
//          }
//      </div>
//    </div>
// }

// Source mapping to map from the source code of the template to the
// in-memory representation.
type Position struct {
	Index int64
	Line  uint32
	Col   uint32
}

func (p Position) String() string {
	return fmt.Sprintf("line %d, col %d (index %d)", p.Line, p.Col, p.Index)
}

// NewPosition initialises a position.
func NewPosition(index int64, line, col uint32) Position {
	return Position{
		Index: index,
		Line:  line,
		Col:   col,
	}
}

// NewExpression creates a Go expression.
func NewExpression(value string, from, to parse.Position) Expression {
	return Expression{
		Value: value,
		Range: Range{
			From: Position{
				Index: int64(from.Index),
				Line:  uint32(from.Line),
				Col:   uint32(from.Col),
			},
			To: Position{
				Index: int64(to.Index),
				Line:  uint32(to.Line),
				Col:   uint32(to.Col),
			},
		},
	}
}

// Range of text within a file.
type Range struct {
	From Position
	To   Position
}

// Expression containing Go code.
type Expression struct {
	Value string
	Range Range
}

// Diagnostic for template file.
type Diagnostic struct {
	Message string
	Range   Range
}

type TemplateFile struct {
	// Header contains comments or whitespace at the top of the file.
	Header []TemplateFileGoExpression
	// Package expression.
	Package Package
	// Nodes in the file.
	Nodes []TemplateFileNode
	// Diagnostics contains any errors or warnings.
	Diagnostics []Diagnostic
}

func (tf TemplateFile) Write(w io.Writer) error {
	for _, n := range tf.Header {
		if err := n.Write(w, 0); err != nil {
			return err
		}
	}
	var indent int
	if err := tf.Package.Write(w, indent); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\n\n"); err != nil {
		return err
	}
	for i := 0; i < len(tf.Nodes); i++ {
		if err := tf.Nodes[i].Write(w, indent); err != nil {
			return err
		}
		if _, err := io.WriteString(w, getNodeWhitespace(tf.Nodes, i)); err != nil {
			return err
		}
	}
	return nil
}

func getNodeWhitespace(nodes []TemplateFileNode, i int) string {
	if i == len(nodes)-1 {
		return "\n"
	}
	if _, nextIsTemplate := nodes[i+1].(HTMLTemplate); nextIsTemplate {
		if e, isGo := nodes[i].(TemplateFileGoExpression); isGo && endsWithComment(e.Expression.Value) {
			return "\n"
		}
	}
	return "\n\n"
}

func endsWithComment(s string) bool {
	lineSlice := strings.Split(s, "\n")
	return strings.HasPrefix(lineSlice[len(lineSlice)-1], "//")
}

// TemplateFileNode can be a Template, CSS, Script or Go.
type TemplateFileNode interface {
	IsTemplateFileNode() bool
	Write(w io.Writer, indent int) error
}

// TemplateFileGoExpression within a TemplateFile
type TemplateFileGoExpression struct {
	Expression Expression
}

func (exp TemplateFileGoExpression) IsTemplateFileNode() bool { return true }
func (exp TemplateFileGoExpression) Write(w io.Writer, indent int) error {
	data, err := format.Source([]byte(exp.Expression.Value))
	if err != nil {
		return writeIndent(w, indent, exp.Expression.Value)
	}
	_, err = w.Write(data)
	return err
}

func writeIndent(w io.Writer, level int, s ...string) (err error) {
	indent := strings.Repeat("\t", level)
	if _, err = io.WriteString(w, indent); err != nil {
		return err
	}
	for _, ss := range s {
		_, err = io.WriteString(w, ss)
		if err != nil {
			return
		}
	}
	return
}

type Package struct {
	Expression Expression
}

func (p Package) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, p.Expression.Value)
}

// Whitespace.
type Whitespace struct {
	Value string
}

func (ws Whitespace) IsNode() bool { return true }

func (ws Whitespace) Write(w io.Writer, indent int) error {
	if ws.Value == "" || !strings.Contains(ws.Value, "\n") {
		return nil
	}
	// https://developer.mozilla.org/en-US/docs/Web/API/Document_Object_Model/Whitespace
	// - All spaces and tabs immediately before and after a line break are ignored.
	// - All tab characters are handled as space characters.
	// - Line breaks are converted to spaces.
	// Any space immediately following another space (even across two separate inline elements) is ignored.
	// Sequences of spaces at the beginning and end of an element are removed.

	// Notes: Since we only have whitespace in this node, we can strip anything that isn't a line break.
	// Since any space following another space is ignored, we can collapse to a single rule.
	// So, the rule is... if there's a newline, it becomes a single space, or it's stripped.
	// We have to remove the start and end space elsewhere.
	_, err := io.WriteString(w, " ")
	return err
}

// CSS definition.
//
//	css Name() {
//	  color: #ffffff;
//	  background-color: { constants.BackgroundColor };
//	  background-image: url('./somewhere.png');
//	}
type CSSTemplate struct {
	Name       Expression
	Properties []CSSProperty
}

func (css CSSTemplate) IsTemplateFileNode() bool { return true }
func (css CSSTemplate) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "css ", css.Name.Value, "() {\n"); err != nil {
		return err
	}
	for _, p := range css.Properties {
		if err := p.Write(w, indent+1); err != nil {
			return err
		}
	}
	if err := writeIndent(w, indent, "}"); err != nil {
		return err
	}
	return nil
}

// CSSProperty is a CSS property and value pair.
type CSSProperty interface {
	IsCSSProperty() bool
	Write(w io.Writer, indent int) error
}

// color: #ffffff;
type ConstantCSSProperty struct {
	Name  string
	Value string
}

func (c ConstantCSSProperty) IsCSSProperty() bool { return true }
func (c ConstantCSSProperty) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, c.String(false)); err != nil {
		return err
	}
	return nil
}

func (c ConstantCSSProperty) String(minified bool) string {
	sb := new(strings.Builder)
	sb.WriteString(c.Name)
	if minified {
		sb.WriteString(":")
	} else {
		sb.WriteString(": ")
	}
	sb.WriteString(c.Value)
	sb.WriteString(";")
	if !minified {
		sb.WriteString("\n")
	}
	return sb.String()
}

// background-color: { constants.BackgroundColor };
type ExpressionCSSProperty struct {
	Name  string
	Value StringExpression
}

func (c ExpressionCSSProperty) IsCSSProperty() bool { return true }
func (c ExpressionCSSProperty) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, c.Name, ": "); err != nil {
		return err
	}
	if err := c.Value.Write(w, 0); err != nil {
		return err
	}
	if _, err := w.Write([]byte(";\n")); err != nil {
		return err
	}
	return nil
}

// <!DOCTYPE html>
type DocType struct {
	Value string
}

func (dt DocType) IsNode() bool { return true }
func (dt DocType) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, "<!DOCTYPE ", dt.Value, ">")
}

// HTMLTemplate definition.
//
//	templ Name(p Parameter) {
//	  if ... {
//	      <Element></Element>
//	  }
//	}
type HTMLTemplate struct {
	Diagnostics []Diagnostic
	Expression  Expression
	Children    []Node
}

func (t HTMLTemplate) IsTemplateFileNode() bool { return true }

func (t HTMLTemplate) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "templ ", t.Expression.Value, " {\n"); err != nil {
		return err
	}
	if err := writeNodesIndented(w, indent+1, t.Children); err != nil {
		return err
	}
	if err := writeIndent(w, indent, "}"); err != nil {
		return err
	}
	return nil
}

// TrailingSpace defines the whitespace that may trail behind the close of an element, a
// text node, or string expression.
type TrailingSpace string

const (
	SpaceNone       TrailingSpace = ""
	SpaceHorizontal TrailingSpace = " "
	SpaceVertical   TrailingSpace = "\n"
)

var ErrNonSpaceCharacter = errors.New("non space character found")

func NewTrailingSpace(s string) (ts TrailingSpace, err error) {
	var hasHorizontalSpace bool
	for _, r := range s {
		if r == '\n' {
			return SpaceVertical, nil
		}
		if unicode.IsSpace(r) {
			hasHorizontalSpace = true
			continue
		}
		return ts, ErrNonSpaceCharacter
	}
	if hasHorizontalSpace {
		return SpaceHorizontal, nil
	}
	return SpaceNone, nil
}

type Nodes struct {
	Diagnostics []Diagnostic
	Nodes       []Node
}

// A Node appears within a template, e.g. an StringExpression, Element, IfExpression etc.
type Node interface {
	IsNode() bool
	// Write out the string.
	Write(w io.Writer, indent int) error
}

type WhitespaceTrailer interface {
	Trailing() TrailingSpace
}

var (
	_ WhitespaceTrailer = Element{}
	_ WhitespaceTrailer = Text{}
	_ WhitespaceTrailer = StringExpression{}
)

// Text node within the document.
type Text struct {
	// Value is the raw HTML encoded value.
	Value string
	// TrailingSpace lists what happens after the text.
	TrailingSpace TrailingSpace
}

func (t Text) Trailing() TrailingSpace {
	return t.TrailingSpace
}

func (t Text) IsNode() bool { return true }
func (t Text) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, t.Value)
}

// <a .../> or <div ...>...</div>
type Element struct {
	Name           string
	Attributes     []Attribute
	IndentAttrs    bool
	Children       []Node
	IndentChildren bool
	TrailingSpace  TrailingSpace
	Diagnostics    []Diagnostic
}

func (e Element) Trailing() TrailingSpace {
	return e.TrailingSpace
}

var voidElements = map[string]struct{}{
	"area": {}, "base": {}, "br": {}, "col": {}, "command": {}, "embed": {}, "hr": {}, "img": {}, "input": {}, "keygen": {}, "link": {}, "meta": {}, "param": {}, "source": {}, "track": {}, "wbr": {},
}

// https://www.w3.org/TR/2011/WD-html-markup-20110113/syntax.html#void-element
func (e Element) IsVoidElement() bool {
	_, ok := voidElements[e.Name]
	return ok
}

func (e Element) hasNonWhitespaceChildren() bool {
	for _, c := range e.Children {
		if _, isWhitespace := c.(Whitespace); !isWhitespace {
			return true
		}
	}
	return false
}

var blockElements = map[string]struct{}{
	"address": {}, "article": {}, "aside": {}, "body": {}, "blockquote": {}, "canvas": {}, "dd": {}, "div": {}, "dl": {}, "dt": {}, "fieldset": {}, "figcaption": {}, "figure": {}, "footer": {}, "form": {}, "h1": {}, "h2": {}, "h3": {}, "h4": {}, "h5": {}, "h6": {}, "head": {}, "header": {}, "hr": {}, "html": {}, "li": {}, "main": {}, "meta": {}, "nav": {}, "noscript": {}, "ol": {}, "p": {}, "pre": {}, "script": {}, "section": {}, "table": {}, "template": {}, "tfoot": {}, "turbo-stream": {}, "ul": {}, "video": {},
	// Not strictly block but for the purposes of layout, they are.
	"title": {}, "style": {}, "link": {}, "td": {}, "th": {}, "tr": {}, "br": {},
}

func (e Element) IsBlockElement() bool {
	_, ok := blockElements[e.Name]
	return ok
}

// Validate that no invalid expressions have been used.
func (e Element) Validate() (msgs []string, ok bool) {
	// Validate that style attributes are constant.
	for _, attr := range e.Attributes {
		if exprAttr, isExprAttr := attr.(ExpressionAttribute); isExprAttr {
			if strings.EqualFold(exprAttr.Name, "style") {
				msgs = append(msgs, "invalid style attribute: style attributes cannot be a templ expression")
			}
		}
	}
	// Validate that script and style tags don't contain expressions.
	if strings.EqualFold(e.Name, "script") || strings.EqualFold(e.Name, "style") {
		if containsNonTextNodes(e.Children) {
			msgs = append(msgs, "invalid node contents: script and style attributes must only contain text")
		}
	}
	return msgs, len(msgs) == 0
}

func containsNonTextNodes(nodes []Node) bool {
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		switch n.(type) {
		case Text:
			continue
		case Whitespace:
			continue
		default:
			return true
		}
	}
	return false
}

func (e Element) IsNode() bool { return true }
func (e Element) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "<", e.Name); err != nil {
		return err
	}
	for i := 0; i < len(e.Attributes); i++ {
		a := e.Attributes[i]
		// Only the conditional attributes get indented.
		var attrIndent int
		if e.IndentAttrs {
			if _, err := w.Write([]byte("\n")); err != nil {
				return err
			}
			attrIndent = indent + 1
		} else {
			if _, err := w.Write([]byte(" ")); err != nil {
				return err
			}
		}
		if err := a.Write(w, attrIndent); err != nil {
			return err
		}
	}
	var closeAngleBracketIndent int
	if e.IndentAttrs {
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
		closeAngleBracketIndent = indent
	}
	if e.hasNonWhitespaceChildren() {
		if e.IndentChildren {
			if err := writeIndent(w, closeAngleBracketIndent, ">\n"); err != nil {
				return err
			}
			if err := writeNodesIndented(w, indent+1, e.Children); err != nil {
				return err
			}
			if err := writeIndent(w, indent, "</", e.Name, ">"); err != nil {
				return err
			}
			return nil
		}
		if err := writeIndent(w, closeAngleBracketIndent, ">"); err != nil {
			return err
		}
		if err := writeNodesWithoutIndentation(w, e.Children); err != nil {
			return err
		}
		if _, err := w.Write([]byte("</" + e.Name + ">")); err != nil {
			return err
		}
		return nil
	}
	if e.IsVoidElement() {
		if err := writeIndent(w, closeAngleBracketIndent, "/>"); err != nil {
			return err
		}
		return nil
	}
	if err := writeIndent(w, closeAngleBracketIndent, "></", e.Name, ">"); err != nil {
		return err
	}
	return nil
}

func writeNodesWithoutIndentation(w io.Writer, nodes []Node) error {
	return writeNodes(w, 0, nodes, false)
}

func writeNodesIndented(w io.Writer, level int, nodes []Node) error {
	return writeNodes(w, level, nodes, true)
}

func writeNodes(w io.Writer, level int, nodes []Node, indent bool) error {
	startLevel := level
	for i := 0; i < len(nodes); i++ {
		_, isWhitespace := nodes[i].(Whitespace)

		// Skip whitespace nodes.
		if isWhitespace {
			continue
		}
		if err := nodes[i].Write(w, level); err != nil {
			return err
		}

		// Apply trailing whitespace if present.
		trailing := SpaceVertical
		if wst, isWhitespaceTrailer := nodes[i].(WhitespaceTrailer); isWhitespaceTrailer {
			trailing = wst.Trailing()
		}
		// Put a newline after the last node in indentation mode.
		if indent && ((nextNodeIsBlock(nodes, i) || i == len(nodes)-1) || shouldAlwaysBreakAfter(nodes[i])) {
			trailing = SpaceVertical
		}
		switch trailing {
		case SpaceNone:
			level = 0
		case SpaceHorizontal:
			level = 0
		case SpaceVertical:
			level = startLevel
		}
		if _, err := w.Write([]byte(trailing)); err != nil {
			return err
		}
	}
	return nil
}

func shouldAlwaysBreakAfter(node Node) bool {
	if el, isElement := node.(Element); isElement {
		return strings.EqualFold(el.Name, "br") || strings.EqualFold(el.Name, "hr")
	}
	return false
}

func nextNodeIsBlock(nodes []Node, i int) bool {
	if len(nodes)-1 < i+1 {
		return false
	}
	return isBlockNode(nodes[i+1])
}

func isBlockNode(node Node) bool {
	switch n := node.(type) {
	case IfExpression:
		return true
	case SwitchExpression:
		return true
	case ForExpression:
		return true
	case Element:
		return n.IsBlockElement() || n.IndentChildren
	}
	return false
}

type RawElement struct {
	Name       string
	Attributes []Attribute
	Contents   string
}

func (e RawElement) IsNode() bool { return true }
func (e RawElement) Write(w io.Writer, indent int) error {
	// Start.
	if err := writeIndent(w, indent, "<", e.Name); err != nil {
		return err
	}
	for i := 0; i < len(e.Attributes); i++ {
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
		a := e.Attributes[i]
		// Don't indent the attributes, only the conditional attributes get indented.
		if err := a.Write(w, 0); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte(">")); err != nil {
		return err
	}
	// Contents.
	if _, err := w.Write([]byte(e.Contents)); err != nil {
		return err
	}
	// Close.
	if _, err := w.Write([]byte("</" + e.Name + ">")); err != nil {
		return err
	}
	return nil
}

type Attribute interface {
	// Write out the string.
	Write(w io.Writer, indent int) error
}

// <hr noshade/>
type BoolConstantAttribute struct {
	Name string
}

func (bca BoolConstantAttribute) String() string {
	return bca.Name
}

func (bca BoolConstantAttribute) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, bca.String())
}

// href=""
type ConstantAttribute struct {
	Name        string
	Value       string
	SingleQuote bool
}

func (ca ConstantAttribute) String() string {
	quote := `"`
	if ca.SingleQuote {
		quote = `'`
	}
	return ca.Name + `=` + quote + ca.Value + quote
}

func (ca ConstantAttribute) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, ca.String())
}

// noshade={ templ.Bool(...) }
type BoolExpressionAttribute struct {
	Name       string
	Expression Expression
}

func (ea BoolExpressionAttribute) String() string {
	return ea.Name + `?={ ` + ea.Expression.Value + ` }`
}

func (ea BoolExpressionAttribute) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, ea.String())
}

// href={ ... }
type ExpressionAttribute struct {
	Name       string
	Expression Expression
}

func (ea ExpressionAttribute) String() string {
	sb := new(strings.Builder)
	_ = ea.Write(sb, 0)
	return sb.String()
}

func (ea ExpressionAttribute) formatExpression() (exp []string) {
	trimmed := strings.TrimSpace(ea.Expression.Value)
	if !strings.Contains(trimmed, "\n") {
		formatted, err := format.Source([]byte(trimmed))
		if err != nil {
			return []string{trimmed}
		}
		return []string{string(formatted)}
	}

	buf := bytes.NewBufferString("[]any{\n")
	buf.WriteString(trimmed)
	buf.WriteString("\n}")

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return []string{trimmed}
	}

	// Trim prefix and suffix.
	lines := strings.Split(string(formatted), "\n")
	if len(lines) < 3 {
		return []string{trimmed}
	}

	// Return.
	return lines[1 : len(lines)-1]
}

func (ea ExpressionAttribute) Write(w io.Writer, indent int) (err error) {
	lines := ea.formatExpression()
	if len(lines) == 1 {
		return writeIndent(w, indent, ea.Name, `={ `, lines[0], ` }`)
	}

	if err = writeIndent(w, indent, ea.Name, "={\n"); err != nil {
		return err
	}
	for _, line := range lines {
		if err = writeIndent(w, indent, line, "\n"); err != nil {
			return err
		}
	}
	return writeIndent(w, indent, "}")
}

// <a { spread... } />
type SpreadAttributes struct {
	Expression Expression
}

func (sa SpreadAttributes) String() string {
	return `{ ` + sa.Expression.Value + `... }`
}

func (sa SpreadAttributes) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, sa.String())
}

//	<a href="test" \
//		if active {
//	   class="isActive"
//	 }
type ConditionalAttribute struct {
	Expression Expression
	Then       []Attribute
	Else       []Attribute
}

func (ca ConditionalAttribute) String() string {
	sb := new(strings.Builder)
	_ = ca.Write(sb, 0)
	return sb.String()
}

func (ca ConditionalAttribute) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "if "); err != nil {
		return err
	}
	if _, err := w.Write([]byte(ca.Expression.Value)); err != nil {
		return err
	}
	if _, err := w.Write([]byte(" {\n")); err != nil {
		return err
	}
	{
		indent++
		for _, attr := range ca.Then {
			if err := attr.Write(w, indent); err != nil {
				return err
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				return err
			}
		}
		indent--
	}
	if err := writeIndent(w, indent, "}"); err != nil {
		return err
	}
	if len(ca.Else) == 0 {
		return nil
	}
	// Write the else blocks.
	if _, err := w.Write([]byte(" else {\n")); err != nil {
		return err
	}
	{
		indent++
		for _, attr := range ca.Else {
			if err := attr.Write(w, indent); err != nil {
				return err
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				return err
			}
		}
		indent--
	}
	if err := writeIndent(w, indent, "}"); err != nil {
		return err
	}
	return nil
}

// GoComment.
type GoComment struct {
	Contents  string
	Multiline bool
}

func (c GoComment) IsNode() bool { return true }
func (c GoComment) Write(w io.Writer, indent int) error {
	if c.Multiline {
		return writeIndent(w, indent, "/*", c.Contents, "*/")
	}
	return writeIndent(w, indent, "//", c.Contents)
}

// HTMLComment.
type HTMLComment struct {
	Contents string
}

func (c HTMLComment) IsNode() bool { return true }
func (c HTMLComment) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, "<!--", c.Contents, "-->")
}

// Nodes.

// CallTemplateExpression can be used to create and render a template using data.
// {! Other(p.First, p.Last) }
// or it can be used to render a template parameter.
// {! v }
type CallTemplateExpression struct {
	// Expression returns a template to execute.
	Expression Expression
}

func (cte CallTemplateExpression) IsNode() bool { return true }
func (cte CallTemplateExpression) Write(w io.Writer, indent int) error {
	// Rewrite to new call syntax
	return writeIndent(w, indent, `@`, cte.Expression.Value)
}

// TemplElementExpression can be used to create and render a template using data.
// @Other(p.First, p.Last)
// or it can be used to render a template parameter.
// @v
type TemplElementExpression struct {
	// Expression returns a template to execute.
	Expression Expression
	// Children returns the elements in a block element.
	Children    []Node
	Diagnostics []Diagnostic
}

func (tee TemplElementExpression) IsNode() bool { return true }
func (tee TemplElementExpression) Write(w io.Writer, indent int) error {
	if len(tee.Children) == 0 {
		return writeIndent(w, indent, "@", tee.Expression.Value)
	}
	if err := writeIndent(w, indent, "@", tee.Expression.Value, " {\n"); err != nil {
		return err
	}
	if err := writeNodesIndented(w, indent+1, tee.Children); err != nil {
		return err
	}
	if err := writeIndent(w, indent, "}"); err != nil {
		return err
	}
	return nil
}

// ChildrenExpression can be used to rended the children of a templ element.
// { children ... }
type ChildrenExpression struct{}

func (ChildrenExpression) IsNode() bool { return true }
func (ChildrenExpression) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "{ children... }"); err != nil {
		return err
	}
	return nil
}

// if p.Type == "test" && p.thing {
// }
type IfExpression struct {
	Expression  Expression
	Then        []Node
	ElseIfs     []ElseIfExpression
	Else        []Node
	Diagnostics []Diagnostic
}

type ElseIfExpression struct {
	Expression  Expression
	Then        []Node
	Diagnostics []Diagnostic
}

func (n IfExpression) IsNode() bool { return true }
func (n IfExpression) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "if ", n.Expression.Value, " {\n"); err != nil {
		return err
	}
	indent++
	if err := writeNodesIndented(w, indent, n.Then); err != nil {
		return err
	}
	indent--
	for _, elseIf := range n.ElseIfs {
		if err := writeIndent(w, indent, "} else if ", elseIf.Expression.Value, " {\n"); err != nil {
			return err
		}
		indent++
		if err := writeNodesIndented(w, indent, elseIf.Then); err != nil {
			return err
		}
		indent--
	}
	if len(n.Else) > 0 {
		if err := writeIndent(w, indent, "} else {\n"); err != nil {
			return err
		}
		if err := writeNodesIndented(w, indent+1, n.Else); err != nil {
			return err
		}
	}
	if err := writeIndent(w, indent, "}"); err != nil {
		return err
	}
	return nil
}

//	switch p.Type {
//	 case "Something":
//	}
type SwitchExpression struct {
	Expression Expression
	Cases      []CaseExpression
}

func (se SwitchExpression) IsNode() bool { return true }
func (se SwitchExpression) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "switch ", se.Expression.Value, " {\n"); err != nil {
		return err
	}
	indent++
	for i := 0; i < len(se.Cases); i++ {
		c := se.Cases[i]
		if err := writeIndent(w, indent, c.Expression.Value, "\n"); err != nil {
			return err
		}
		if err := writeNodesIndented(w, indent+1, c.Children); err != nil {
			return err
		}
	}
	indent--
	if err := writeIndent(w, indent, "}"); err != nil {
		return err
	}
	return nil
}

// case "Something":
type CaseExpression struct {
	Expression  Expression
	Children    []Node
	Diagnostics []Diagnostic
}

//	for i, v := range p.Addresses {
//	  {! Address(v) }
//	}
type ForExpression struct {
	Expression  Expression
	Children    []Node
	Diagnostics []Diagnostic
}

func (fe ForExpression) IsNode() bool { return true }
func (fe ForExpression) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "for ", fe.Expression.Value, " {\n"); err != nil {
		return err
	}
	if err := writeNodesIndented(w, indent+1, fe.Children); err != nil {
		return err
	}
	if err := writeIndent(w, indent, "}"); err != nil {
		return err
	}
	return nil
}

// StringExpression is used within HTML elements, and for style values.
// { ... }
type StringExpression struct {
	Expression Expression
	// TrailingSpace lists what happens after the expression.
	TrailingSpace TrailingSpace
}

func (se StringExpression) Trailing() TrailingSpace {
	return se.TrailingSpace
}

func (se StringExpression) IsNode() bool                  { return true }
func (se StringExpression) IsStyleDeclarationValue() bool { return true }
func (se StringExpression) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, `{ `, se.Expression.Value, ` }`)
}

// ScriptTemplate is a script block.
type ScriptTemplate struct {
	Name       Expression
	Parameters Expression
	Value      string
}

func (s ScriptTemplate) IsTemplateFileNode() bool { return true }
func (s ScriptTemplate) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "script ", s.Name.Value, "(", s.Parameters.Value, ") {\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, s.Value); err != nil {
		return err
	}
	if err := writeIndent(w, indent, "}"); err != nil {
		return err
	}
	return nil
}
