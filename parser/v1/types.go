package parser

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/a-h/lexical/parse"
)

// {% package parser %}
//
// {% import "strings" %}
// {% import strs "strings" %}
//
// {% css AddressLineStyle() %}
//   background-color: #ff0000;
//   color: #ffffff;
// {% endcss %}
//
// {% templ RenderAddress(addr Address) %}
// 	<div style={%= AddressLineStyle() %}>{%= addr.Address1 %}</div>
// 	<div>{%= addr.Address2 %}</div>
// 	<div>{%= addr.Address3 %}</div>
// 	<div>{%= addr.Address4 %}</div>
// {% endtempl %}
//
// {% templ Render(p Person) %}
//    <div>
//      <div>{%= p.Name() %}</div>
//      <a href={%= p.URL %}>{%= strings.ToUpper(p.Name()) %}</a>
//      <div>
//          {% if p.Type == "test" %}
//             <span>{%= "Test user" %}</span>
//          {% else %}
// 	    <span>{%= "Not test user" %}</span>
//          {% endif %}
//          {% for _, v := range p.Addresses %}
//             {% call RenderAddress(v) %}
//          {% endfor %}
//      </div>
//    </div>
// {% endtempl %}

// Source mapping to map from the source code of the template to the
// in-memory representation.
type Position struct {
	Index int64
	Line  int
	Col   int
}

func (p Position) String() string {
	return fmt.Sprintf("line %d, col %d (index %d)", p.Line, p.Col, p.Index)
}

// NewPosition initialises a position.
func NewPosition() Position {
	return Position{
		Index: 0,
		Line:  1,
		Col:   0,
	}
}

// NewPositionFromValues initialises a position.
func NewPositionFromValues(index int64, line, col int) Position {
	return Position{
		Index: index,
		Line:  line,
		Col:   col,
	}
}

// NewPositionFromInput creates a position from a parse input.
func NewPositionFromInput(pi parse.Input) Position {
	l, c := pi.Position()
	return Position{
		Index: pi.Index(),
		Line:  l,
		Col:   c,
	}
}

// NewExpression creates a Go expression.
func NewExpression(value string, from, to Position) Expression {
	return Expression{
		Value: value,
		Range: Range{
			From: from,
			To:   to,
		},
	}
}

// NewRange creates a range.
func NewRange(from, to Position) Range {
	return Range{
		From: from,
		To:   to,
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
	Imports []Import
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
	if len(tf.Imports) > 0 {
		for i := 0; i < len(tf.Imports); i++ {
			if err := tf.Imports[i].Write(w, indent); err != nil {
				return err
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
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

// TemplateFileNode can be a Template or a CSS.
type TemplateFileNode interface {
	IsTemplateFileNode() bool
	Write(w io.Writer, indent int) error
}

// {% package parser %}
type Package struct {
	Expression Expression
}

func (p Package) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, "{% package "+p.Expression.Value+" %}")
}

func writeIndent(w io.Writer, level int, s string) (err error) {
	if _, err = w.Write([]byte(strings.Repeat("\t", level))); err != nil {
		return
	}
	_, err = w.Write([]byte(s))
	return
}

// Whitespace.
type Whitespace struct {
	Value string
}

func (ws Whitespace) IsNode() bool { return true }
func (ws Whitespace) Write(w io.Writer, indent int) error {
	// Explicit whitespace nodes are removed from templates because they're auto-formatted.
	return nil
}

// {% import "strings" %}
// {% import strs "strings" %}
type Import struct {
	Expression Expression
}

func (imp Import) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, "{% import "+imp.Expression.Value+" %}")
}

// CSS definition.
// {% css Name() %}
//
//	color: #ffffff;
//	background-color: {%= constants.BackgroundColor %};
//	background-image: url('./somewhere.png');
//
// {% endcss %}
type CSSTemplate struct {
	Name       Expression
	Properties []CSSProperty
}

func (css CSSTemplate) IsTemplateFileNode() bool { return true }
func (css CSSTemplate) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "{% css "+css.Name.Value+"() %}\n"); err != nil {
		return err
	}
	for _, p := range css.Properties {
		if err := p.Write(w, indent+1); err != nil {
			return err
		}
	}
	if err := writeIndent(w, indent, "{% endcss %}"); err != nil {
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
	if err := writeIndent(w, indent, c.Name+": "+c.Value+";\n"); err != nil {
		return err
	}
	return nil
}

// background-color: {%= constants.BackgroundColor %};
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
// {% templ Name(p Parameter) %}
//
//	{% if ... %}
//	<Element></Element>
//
// {% endtempl %}
type HTMLTemplate struct {
	Name       Expression
	Parameters Expression
	Children   []Node
}

func (t HTMLTemplate) IsTemplateFileNode() bool { return true }

func (t HTMLTemplate) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "{% templ "+t.Name.Value+"("+t.Parameters.Value+") %}\n"); err != nil {
		return err
	}
	if err := writeNodesBlock(w, indent+1, t.Children); err != nil {
		return err
	}
	if err := writeIndent(w, indent, "{% endtempl %}"); err != nil {
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
	Name       string
	Attributes []Attribute
	Children   []Node
}

var voidElements = map[string]struct{}{
	"area": {}, "base": {}, "br": {}, "col": {}, "command": {}, "embed": {}, "hr": {}, "img": {}, "input": {}, "keygen": {}, "link": {}, "meta": {}, "param": {}, "source": {}, "track": {}, "wbr": {}}

// https://www.w3.org/TR/2011/WD-html-markup-20110113/syntax.html#void-element
func (e Element) IsVoidElement() bool {
	_, ok := voidElements[e.Name]
	return ok
}

var blockElements = map[string]struct{}{
	"address": {}, "article": {}, "aside": {}, "body": {}, "blockquote": {}, "canvas": {}, "dd": {}, "div": {}, "dl": {}, "dt": {}, "fieldset": {}, "figcaption": {}, "figure": {}, "footer": {}, "form": {}, "h1": {}, "h2": {}, "h3": {}, "h4": {}, "h5": {}, "h6": {}, "head": {}, "header": {}, "hr": {}, "html": {}, "li": {}, "main": {}, "meta": {}, "nav": {}, "noscript": {}, "ol": {}, "p": {}, "pre": {}, "script": {}, "section": {}, "table": {}, "template": {}, "tfoot": {}, "turbo-stream": {}, "ul": {}, "video": {},
}

func (e Element) isBlockElement() bool {
	_, ok := blockElements[e.Name]
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

func (e Element) containsBlockElement() bool {
	for _, c := range e.Children {
		switch n := c.(type) {
		case Whitespace:
			continue
		case Element:
			if n.isBlockElement() {
				return true
			}
			continue
		case StringExpression:
			continue
		case Text:
			continue
		}
		// Any template elements should be considered block.
		return true
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
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
		a := e.Attributes[i]
		if _, err := w.Write([]byte(a.String())); err != nil {
			return err
		}
	}
	if e.hasNonWhitespaceChildren() {
		if e.containsBlockElement() {
			if _, err := w.Write([]byte(">\n")); err != nil {
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
		if _, err := w.Write([]byte(">")); err != nil {
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
		if _, err := w.Write([]byte("/>")); err != nil {
			return err
		}
		return nil
	}
	if _, err := w.Write([]byte("></" + e.Name + ">")); err != nil {
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
		if _, isWhitespace := nodes[i].(Whitespace); isWhitespace {
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

type Attribute interface {
	IsAttribute() bool
	String() string
}

// <hr noshade/>
type BoolConstantAttribute struct {
	Name string
}

func (bca BoolConstantAttribute) IsAttribute() bool { return true }
func (bca BoolConstantAttribute) String() string {
	return bca.Name
}

// href=""
type ConstantAttribute struct {
	Name  string
	Value string
}

func (ca ConstantAttribute) IsAttribute() bool { return true }
func (ca ConstantAttribute) String() string {
	return ca.Name + `="` + html.EscapeString(ca.Value) + `"`
}

// href={%= templ.Bool(...) }
type BoolExpressionAttribute struct {
	Name       string
	Expression Expression
}

func (ea BoolExpressionAttribute) IsAttribute() bool { return true }
func (ea BoolExpressionAttribute) String() string {
	return ea.Name + `?={%= ` + ea.Expression.Value + ` %}`
}

// href={%= ... }
type ExpressionAttribute struct {
	Name       string
	Expression Expression
}

func (ea ExpressionAttribute) IsAttribute() bool { return true }
func (ea ExpressionAttribute) String() string {
	return ea.Name + `={%= ` + ea.Expression.Value + ` %}`
}

// Nodes.

// CallTemplateExpression can be used to create and render a template using data.
// {%! Other(p.First, p.Last) %}
// or it can be used to render a template parameter.
// {%! v %}
type CallTemplateExpression struct {
	// Expression returns a template to execute.
	Expression Expression
}

func (cte CallTemplateExpression) IsNode() bool { return true }
func (cte CallTemplateExpression) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, `{%! `+cte.Expression.Value+` %}`)
}

// {% if p.Type == "test" && p.thing %}
// {% endif %}
type IfExpression struct {
	Expression Expression
	Then       []Node
	Else       []Node
}

func (n IfExpression) IsNode() bool { return true }
func (n IfExpression) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "{% if "+n.Expression.Value+" %}\n"); err != nil {
		return err
	}
	indent++
	if err := writeNodesBlock(w, indent, n.Then); err != nil {
		return err
	}
	indent--
	if len(n.Else) > 0 {
		if err := writeIndent(w, indent, "{% else %}\n"); err != nil {
			return err
		}
		if err := writeNodesBlock(w, indent+1, n.Else); err != nil {
			return err
		}
	}
	if err := writeIndent(w, indent, "{% endif %}"); err != nil {
		return err
	}
	return nil
}

// {% switch p.Type %}
//
//	{% case "Something" %}
//	{% endcase %}
//
// {% endswitch %}
type SwitchExpression struct {
	Expression Expression
	Cases      []CaseExpression
	Default    []Node
}

func (se SwitchExpression) IsNode() bool { return true }
func (se SwitchExpression) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "{% switch "+se.Expression.Value+" %}\n"); err != nil {
		return err
	}
	indent++
	for i := 0; i < len(se.Cases); i++ {
		c := se.Cases[i]
		if err := writeIndent(w, indent, "{% case "+c.Expression.Value+" %}\n"); err != nil {
			return err
		}
		if err := writeNodesBlock(w, indent+1, c.Children); err != nil {
			return err
		}
		if err := writeIndent(w, indent, "{% endcase %}\n"); err != nil {
			return err
		}
	}
	if len(se.Default) > 0 {
		if err := writeIndent(w, indent, "{% default %}\n"); err != nil {
			return err
		}
		if err := writeNodesBlock(w, indent+1, se.Default); err != nil {
			return err
		}
		if err := writeIndent(w, indent, "{% enddefault %}\n"); err != nil {
			return err
		}
	}
	indent--
	if err := writeIndent(w, indent, "{% endswitch %}"); err != nil {
		return err
	}
	return nil
}

// {% case "Something" %}
// ...
// {% endcase %}
type CaseExpression struct {
	Expression Expression
	Children   []Node
}

// {% for i, v := range p.Addresses %}
//
//	{% call Address(v) %}
//
// {% endfor %}
type ForExpression struct {
	Expression Expression
	Children   []Node
}

func (fe ForExpression) IsNode() bool { return true }
func (fe ForExpression) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "{% for "+fe.Expression.Value+" %}\n"); err != nil {
		return err
	}
	if err := writeNodesBlock(w, indent+1, fe.Children); err != nil {
		return err
	}
	if err := writeIndent(w, indent, "{% endfor %}"); err != nil {
		return err
	}
	return nil
}

// StringExpression is used within HTML elements, and for style values.
// {%= ... %}
type StringExpression struct {
	Expression Expression
}

func (se StringExpression) IsNode() bool                  { return true }
func (se StringExpression) IsStyleDeclarationValue() bool { return true }
func (se StringExpression) Write(w io.Writer, indent int) error {
	return writeIndent(w, indent, `{%= `+se.Expression.Value+` %}`)
}

// ScriptTemplate is a script block.
type ScriptTemplate struct {
	Name       Expression
	Parameters Expression
	Value      string
}

func (s ScriptTemplate) IsTemplateFileNode() bool { return true }
func (s ScriptTemplate) Write(w io.Writer, indent int) error {
	if err := writeIndent(w, indent, "{% script "+s.Name.Value+"("+s.Parameters.Value+") %}\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, s.Value); err != nil {
		return err
	}
	if err := writeIndent(w, indent, "{% endscript %}"); err != nil {
		return err
	}
	return nil
}
