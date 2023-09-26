package parser

import (
	"fmt"
	"html"
	"io"
	"strings"

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

type TemplateFile struct {
	Package Package
	Nodes   []TemplateFileNode
}

func (tf TemplateFile) Write(w io.Writer) error {
	var indent int
	if err := tf.Package.Write(w, indent); err != nil {
		return err
	}
	if _, err := w.Write([]byte("\n\n")); err != nil {
		return err
	}
	for i := 0; i < len(tf.Nodes); i++ {
		if err := tf.Nodes[i].Write(w, indent); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n\n")); err != nil {
			return err
		}
	}
	return nil
}

// TemplateFileNode can be a Template, CSS, Script or Go.
type TemplateFileNode interface {
	IsTemplateFileNode() bool
	Write(w io.Writer, indent int) error
}

// GoExpression within a TemplateFile
type GoExpression struct {
	Expression Expression
}

func (exp GoExpression) IsTemplateFileNode() bool { return true }
func (exp GoExpression) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, exp.Expression.Value)
}

func writeIndent(w io.Writer, level int, s string) (err error) {
	if _, err = w.Write([]byte(strings.Repeat("\t", level))); err != nil {
		return
	}
	_, err = w.Write([]byte(s))
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
	if err := writeIndent(w, indent, "css "+css.Name.Value+"() {\n"); err != nil {
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
	var sb strings.Builder
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
	if err := writeIndent(w, indent, c.Name+": "); err != nil {
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
	return writeIndent(w, indent, "<!DOCTYPE "+dt.Value+">")
}

// HTMLTemplate definition.
//
//	templ Name(p Parameter) {
//	  if ... {
//	      <Element></Element>
//	  }
//	}
type HTMLTemplate struct {
	Expression Expression
	Children   []Node
}

func (t HTMLTemplate) IsTemplateFileNode() bool { return true }

func (t HTMLTemplate) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "templ "+t.Expression.Value+" {\n"); err != nil {
		return err
	}
	if err := writeNodesBlock(w, indent+1, t.Children); err != nil {
		return err
	}
	if err := writeIndent(w, indent, "}"); err != nil {
		return err
	}
	return nil
}

// A Node appears within a template, e.g. an StringExpression, Element, IfExpression etc.
type Node interface {
	IsNode() bool
	// Write out the string.
	Write(w io.Writer, indent int) error
}

// Text node within the document.
type Text struct {
	// Value is the raw HTML encoded value.
	Value string
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
	if err := writeIndent(w, indent, "<"+e.Name); err != nil {
		return err
	}
	for i := 0; i < len(e.Attributes); i++ {
		a := e.Attributes[i]
		// Only the conditional attributes get indented.
		var attrIndent int
		if e.IndentAttrs {
			w.Write([]byte("\n"))
			attrIndent = indent + 1
		}
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
		if err := a.Write(w, attrIndent); err != nil {
			return err
		}
	}
	var closeAngleBracketIndent int
	if e.IndentAttrs {
		w.Write([]byte("\n"))
		closeAngleBracketIndent = indent
	}
	if e.hasNonWhitespaceChildren() {
		if e.IndentChildren {
			if err := writeIndent(w, closeAngleBracketIndent, ">\n"); err != nil {
				return err
			}
			if err := writeNodesBlock(w, indent+1, e.Children); err != nil {
				return err
			}
			if err := writeIndent(w, indent, "</"+e.Name+">"); err != nil {
				return err
			}
			return nil
		}
		if err := writeIndent(w, closeAngleBracketIndent, ">"); err != nil {
			return err
		}
		if err := writeNodesInline(w, e.Children); err != nil {
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
	if err := writeIndent(w, closeAngleBracketIndent, "></"+e.Name+">"); err != nil {
		return err
	}
	return nil
}

func writeNodesInline(w io.Writer, nodes []Node) error {
	return writeNodes(w, 0, nodes, false)
}

func writeNodesBlock(w io.Writer, indent int, nodes []Node) error {
	return writeNodes(w, indent, nodes, true)
}

func writeNodes(w io.Writer, indent int, nodes []Node, block bool) error {
	for i := 0; i < len(nodes); i++ {
		// Terminating and leading whitespace is stripped.
		_, isWhitespace := nodes[i].(Whitespace)
		if isWhitespace && (i == 0 || i == len(nodes)-1) {
			continue
		}
		// Whitespace is stripped from block elements.
		if isWhitespace && block {
			continue
		}
		if err := nodes[i].Write(w, indent); err != nil {
			return err
		}
		if block {
			if _, err := w.Write([]byte("\n")); err != nil {
				return err
			}
		}
	}
	return nil
}

type RawElement struct {
	Name       string
	Attributes []Attribute
	Contents   string
}

func (e RawElement) IsNode() bool { return true }
func (e RawElement) Write(w io.Writer, indent int) error {
	// Start.
	if err := writeIndent(w, indent, "<"+e.Name); err != nil {
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
	Name  string
	Value string
}

func (ca ConstantAttribute) String() string {
	return ca.Name + `="` + html.EscapeString(ca.Value) + `"`
}

func (ca ConstantAttribute) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, ca.String())
}

// href={ templ.Bool(...) }
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
	return ea.Name + `={ ` + ea.Expression.Value + ` }`
}

func (ea ExpressionAttribute) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, ea.String())
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
	return writeIndent(w, indent, `{! `+cte.Expression.Value+` }`)
}

// TemplElementExpression can be used to create and render a template using data.
// @Other(p.First, p.Last)
// or it can be used to render a template parameter.
// @v
type TemplElementExpression struct {
	// Expression returns a template to execute.
	Expression Expression
	// Children returns the elements in a block element.
	Children []Node
}

func (tee TemplElementExpression) IsNode() bool { return true }
func (tee TemplElementExpression) Write(w io.Writer, indent int) error {
	if len(tee.Children) == 0 {
		return writeIndent(w, indent, fmt.Sprintf("@%s", tee.Expression.Value))
	}
	if err := writeIndent(w, indent, fmt.Sprintf("@%s {\n", tee.Expression.Value)); err != nil {
		return err
	}
	if err := writeNodesBlock(w, indent+1, tee.Children); err != nil {
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
	Expression Expression
	Then       []Node
	ElseIfs    []ElseIfExpression
	Else       []Node
}

type ElseIfExpression struct {
	Expression Expression
	Then       []Node
}

func (n IfExpression) IsNode() bool { return true }
func (n IfExpression) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "if "+n.Expression.Value+" {\n"); err != nil {
		return err
	}
	indent++
	if err := writeNodesBlock(w, indent, n.Then); err != nil {
		return err
	}
	indent--
	for _, elseIf := range n.ElseIfs {
		if err := writeIndent(w, indent, "} else if "+elseIf.Expression.Value+" {\n"); err != nil {
			return err
		}
		indent++
		if err := writeNodesBlock(w, indent, elseIf.Then); err != nil {
			return err
		}
		indent--
	}
	if len(n.Else) > 0 {
		if err := writeIndent(w, indent, "} else {\n"); err != nil {
			return err
		}
		if err := writeNodesBlock(w, indent+1, n.Else); err != nil {
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
	if err := writeIndent(w, indent, "switch "+se.Expression.Value+" {\n"); err != nil {
		return err
	}
	indent++
	for i := 0; i < len(se.Cases); i++ {
		c := se.Cases[i]
		if err := writeIndent(w, indent, c.Expression.Value+"\n"); err != nil {
			return err
		}
		if err := writeNodesBlock(w, indent+1, c.Children); err != nil {
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
	Expression Expression
	Children   []Node
}

//	for i, v := range p.Addresses {
//	  {! Address(v) }
//	}
type ForExpression struct {
	Expression Expression
	Children   []Node
}

func (fe ForExpression) IsNode() bool { return true }
func (fe ForExpression) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "for "+fe.Expression.Value+" {\n"); err != nil {
		return err
	}
	if err := writeNodesBlock(w, indent+1, fe.Children); err != nil {
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
}

func (se StringExpression) IsNode() bool                  { return true }
func (se StringExpression) IsStyleDeclarationValue() bool { return true }
func (se StringExpression) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, `{ `+se.Expression.Value+` }`)
}

// ScriptTemplate is a script block.
type ScriptTemplate struct {
	Name       Expression
	Parameters Expression
	Value      string
}

func (s ScriptTemplate) IsTemplateFileNode() bool { return true }
func (s ScriptTemplate) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "script "+s.Name.Value+"("+s.Parameters.Value+") {\n"); err != nil {
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
