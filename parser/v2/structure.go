package parser

// TemplateFileNodes are the top level nodes of a templ file.
var (
	// css name() { ... }
	_ TemplateFileNode = (*CSSTemplate)(nil)
	// templ name() { ... }
	_ TemplateFileNode = (*HTMLTemplate)(nil)
	// script name() { ... }
	_ TemplateFileNode = (*ScriptTemplate)(nil)
	// Go code within a templ file.
	_ TemplateFileNode = (*TemplateFileGoExpression)(nil)
)

// Nodes are all the nodes you can find in a `templ` component.
var (
	_ Node = (*Text)(nil)
	_ Node = (*Element)(nil)
	_ Node = (*ScriptElement)(nil)
	_ Node = (*RawElement)(nil)
	_ Node = (*GoComment)(nil)
	_ Node = (*HTMLComment)(nil)
	_ Node = (*CallTemplateExpression)(nil)
	_ Node = (*TemplElementExpression)(nil)
	_ Node = (*ChildrenExpression)(nil)
	_ Node = (*IfExpression)(nil)
	_ Node = (*SwitchExpression)(nil)
	_ Node = (*ForExpression)(nil)
	_ Node = (*StringExpression)(nil)
	_ Node = (*GoCode)(nil)
	_ Node = (*Whitespace)(nil)
	_ Node = (*DocType)(nil)
)

// Element nodes can have the following attributes.
var (
	_ Attribute = (*BoolConstantAttribute)(nil)
	_ Attribute = (*ConstantAttribute)(nil)
	_ Attribute = (*BoolExpressionAttribute)(nil)
	_ Attribute = (*ExpressionAttribute)(nil)
	_ Attribute = (*SpreadAttributes)(nil)
	_ Attribute = (*ConditionalAttribute)(nil)
)
