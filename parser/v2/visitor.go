package parser

// Visitor is an interface for visiting nodes in the parse tree.
type Visitor interface {
	VisitTemplateFile(*TemplateFile) error
	VisitTemplateFileGoExpression(*TemplateFileGoExpression) error
	VisitPackage(*Package) error
	VisitWhitespace(*Whitespace) error
	VisitCSSTemplate(*CSSTemplate) error
	VisitConstantCSSProperty(*ConstantCSSProperty) error
	VisitExpressionCSSProperty(*ExpressionCSSProperty) error
	VisitDocType(*DocType) error
	VisitHTMLTemplate(*HTMLTemplate) error
	VisitText(*Text) error
	VisitElement(*Element) error
	VisitScriptElement(*ScriptElement) error
	VisitRawElement(*RawElement) error
	VisitBoolConstantAttribute(*BoolConstantAttribute) error
	VisitConstantAttribute(*ConstantAttribute) error
	VisitBoolExpressionAttribute(*BoolExpressionAttribute) error
	VisitExpressionAttribute(*ExpressionAttribute) error
	VisitSpreadAttributes(*SpreadAttributes) error
	VisitConditionalAttribute(*ConditionalAttribute) error
	VisitGoComment(*GoComment) error
	VisitHTMLComment(*HTMLComment) error
	VisitCallTemplateExpression(*CallTemplateExpression) error
	VisitTemplElementExpression(*TemplElementExpression) error
	VisitChildrenExpression(*ChildrenExpression) error
	VisitIfExpression(*IfExpression) error
	VisitSwitchExpression(*SwitchExpression) error
	VisitForExpression(*ForExpression) error
	VisitGoCode(*GoCode) error
	VisitStringExpression(*StringExpression) error
	VisitScriptTemplate(*ScriptTemplate) error
}
