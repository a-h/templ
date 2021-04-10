package templ

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

type TemplateFile struct {
	Package   PackageExpression
	Imports   []ImportExpression
	Templates []Template
}

// {% package templ %}
type PackageExpression struct {
	Expression string
}

// {% import "strings" %}
// {% import strs "strings" %}
type ImportExpression struct {
	Expression string
}

// Template definition.
// {% templ Name(p Parameter) %}
//   {% if ... %}
//   <Element></Element>
// {% endtempl %}
type Template struct {
	Name                string
	ParameterExpression string
	Children            []Node
}

// Node can be an expression or an element.
type Node interface {
	IsNode() bool
}

// <a .../> or <div ...>...</div>
type Element struct {
	Name       string
	Attributes []Attribute
	Children   []Node
}

func (e Element) IsNode() bool {
	return true
}

type Attribute interface {
	IsAttribute() bool
}

// href=""
type ConstantAttribute struct {
	Name  string
	Value string
}

func (ca ConstantAttribute) IsAttribute() bool { return true }

// href={%= ... }
type ExpressionAttribute struct {
	Name  string
	Value AttributeStringExpression
}

func (ea ExpressionAttribute) IsAttribute() bool { return true }

// {%= ... %}
type AttributeStringExpression struct {
	Expression string
}

// Nodes.

// {%= ... %}
type NodeStringExpression struct {
	Expression string
}

func (ns NodeStringExpression) IsNode() bool {
	return true
}

// {% call Other(p.First, p.Last) %}
type CallTemplateExpression struct {
	// Name of the template to execute
	Name               string
	ArgumentExpression string
}

// {% if p.Type == "test" && p.thing %}
// {% endif %}
type IfExpression struct {
	Expression string
	True       []Node
	ElseIf     []ElseIfExpression
	Else       []Node
}

type ElseIfExpression struct {
	Expression string
	True       []Node
}

// {% switch p.Type %}
//  {% case "Something" %}
//  {% endcase %}
// {% endswitch %}
type SwitchExpression struct {
	Expression string
	Cases      []CaseExpression
	Default    *CaseExpression
}

// {% case "Something" %}
// ...
// {% endcase %}
type CaseExpression struct {
	Expression string
	Children   []Node
}

// {% for i, v := range p.Addresses %}
//   {% call Address(v) %}
// {% endfor %}
type ForRangeExpression struct {
	Expression string
	Children   []Node
}
