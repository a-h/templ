package parser

// TemplateFileNodes are the top level nodes of a templ file.
var (
	// css name() { ... }
	_ TemplateFileNode = CSSTemplate{}
	// templ name() { ... }
	_ TemplateFileNode = HTMLTemplate{}
	// script name() { ... }
	_ TemplateFileNode = ScriptTemplate{}
	// Go code within a templ file.
	_ TemplateFileNode = TemplateFileGoExpression{}
)

// Nodes are all the nodes you can find in a `templ` component.
var (
	_ Node = Text{}
	_ Node = Element{}
	_ Node = RawElement{}
	_ Node = GoComment{}
	_ Node = HTMLComment{}
	_ Node = CallTemplateExpression{}
	_ Node = TemplElementExpression{}
	_ Node = ChildrenExpression{}
	_ Node = IfExpression{}
	_ Node = SwitchExpression{}
	_ Node = ForExpression{}
	_ Node = StringExpression{}
	_ Node = GoCode{}
	_ Node = Whitespace{}
	_ Node = DocType{}
)

// Element nodes can have the following attributes.
var (
	_ Attribute = BoolConstantAttribute{}
	_ Attribute = ConstantAttribute{}
	_ Attribute = BoolExpressionAttribute{}
	_ Attribute = ExpressionAttribute{}
	_ Attribute = SpreadAttributes{}
	_ Attribute = ConditionalAttribute{}
)
