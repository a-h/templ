package templ

import (
	"fmt"

	"github.com/a-h/lexical/parse"
)

//TODO: Add comment line?

// {% package templ %}
//
// {% import "strings" %}
// {% import strs "strings" %}
//
// {% templ Person(p Person) %}
//    <div>
//      <div>{%= p.Name() %}</div>
//      <a href={%= p.URL %}>{%= strings.ToUpper(p.Name()) %}</a>
//      <div>
//          {% call Other(p) %}
//          {% if p.Type == "test" && p.thing %}
//      	<span>{ p.type }</span>
//          {% else %}
//      	<span>Not test</span>
//          {% endif %}
//          {% switch p.Type %}
//             {% case "Something" %}
//              <h1>Something</h1>
//             {% end case %}
//          {% endswitch %}
//          {% for i, v := range p.Addresses %}
//             {% call Address(v) %}
//          {% endfor %}
//          {%= Complex(TypeName{Field: p.PhoneNumber}) %}
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

// Item defines that the item is a template item.
type Item interface {
	IsItem() bool
}

type TemplateFile struct {
	Package   Package
	Imports   []Import
	Templates []Template
}

// {% package templ %}
type Package struct {
	Expression Expression
}

func (p Package) IsItem() bool { return true }

// Whitespace.
type Whitespace struct {
	Value string
}

func (w Whitespace) IsItem() bool { return true }
func (w Whitespace) IsNode() bool { return true }

// {% import "strings" %}
// {% import strs "strings" %}
type Import struct {
	Expression Expression
}

func (imp Import) IsItem() bool { return true }

// Template definition.
// {% templ Name(p Parameter) %}
//   {% if ... %}
//   <Element></Element>
// {% endtempl %}
type Template struct {
	Name       Expression
	Parameters Expression
	Children   []Node
}

func (t Template) IsItem() bool { return true }

// A Node appears within a template, e.g. a StringExpression, Element, IfExpression etc.
type Node interface {
	IsNode() bool
}

// <a .../> or <div ...>...</div>
type Element struct {
	Name       string
	Attributes []Attribute
	Children   []Node
}

func (e Element) IsItem() bool { return true }
func (e Element) IsNode() bool { return true }

type Attribute interface {
	IsItem() bool
	IsAttribute() bool
}

// href=""
type ConstantAttribute struct {
	Name  string
	Value string
}

func (ca ConstantAttribute) IsItem() bool      { return true }
func (ca ConstantAttribute) IsAttribute() bool { return true }

// href={%= ... }
type ExpressionAttribute struct {
	Name  string
	Value StringExpression
}

func (ea ExpressionAttribute) IsItem() bool      { return true }
func (ea ExpressionAttribute) IsAttribute() bool { return true }

// Nodes.

// {%= ... %}
type StringExpression struct {
	Expression Expression
}

func (se StringExpression) IsItem() bool { return true }
func (se StringExpression) IsNode() bool { return true }

// {% call Other(p.First, p.Last) %}
type CallTemplateExpression struct {
	// Name of the template to execute.
	Name Expression
	// Arguments to pass.
	Arguments Expression
}

func (cte CallTemplateExpression) IsItem() bool { return true }
func (cte CallTemplateExpression) IsNode() bool { return true }

// {% if p.Type == "test" && p.thing %}
// {% endif %}
type IfExpression struct {
	Expression Expression
	Then       []Node
	Else       []Node
}

func (n IfExpression) IsItem() bool { return true }
func (n IfExpression) IsNode() bool { return true }

// {% switch p.Type %}
//  {% case "Something" %}
//  {% endcase %}
// {% endswitch %}
type SwitchExpression struct {
	Expression Expression
	Cases      []CaseExpression
	Default    *CaseExpression
}

func (se SwitchExpression) IsItem() bool { return true }
func (se SwitchExpression) IsNode() bool { return true }

// {% case "Something" %}
// ...
// {% endcase %}
type CaseExpression struct {
	Expression Expression
	Children   []Node
}

func (ce CaseExpression) IsItem() bool { return true }

// {% for i, v := range p.Addresses %}
//   {% call Address(v) %}
// {% endfor %}
type ForExpression struct {
	Expression Expression
	Children   []Node
}

func (fe ForExpression) IsItem() bool { return true }
