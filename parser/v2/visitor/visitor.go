package visitor

import "github.com/a-h/templ/parser/v2"

// New returns a default Visitor. Each function in the Visitor struct can be
// overridden to provide custom behavior when visiting nodes in the parse tree.
func New() *Visitor {
	v := &Visitor{}

	// Set default implementations for all visitor functions.
	v.TemplateFile = func(n *parser.TemplateFile) error {
		for _, header := range n.Header {
			if err := v.VisitTemplateFileGoExpression(header); err != nil {
				return err
			}
		}
		if err := v.VisitPackage(&n.Package); err != nil {
			return err
		}
		for _, node := range n.Nodes {
			if err := node.Visit(v); err != nil {
				return err
			}
		}
		return nil
	}
	v.TemplateFileGoExpression = func(n *parser.TemplateFileGoExpression) error {
		return nil
	}
	v.Package = func(n *parser.Package) error {
		return nil
	}
	v.Whitespace = func(n *parser.Whitespace) error {
		return nil
	}
	v.CSSTemplate = func(n *parser.CSSTemplate) error {
		for _, prop := range n.Properties {
			if err := prop.Visit(v); err != nil {
				return err
			}
		}
		return nil
	}
	v.ConstantCSSProperty = func(n *parser.ConstantCSSProperty) error {
		return nil
	}
	v.ExpressionCSSProperty = func(n *parser.ExpressionCSSProperty) error {
		if err := n.Value.Visit(v); err != nil {
			return err
		}
		return nil
	}
	v.DocType = func(n *parser.DocType) error {
		return nil
	}
	v.HTMLTemplate = func(n *parser.HTMLTemplate) error {
		for _, child := range n.Children {
			if err := child.Visit(v); err != nil {
				return err
			}
		}
		return nil
	}
	v.Text = func(n *parser.Text) error {
		return nil
	}
	v.Element = func(n *parser.Element) error {
		for _, attr := range n.Attributes {
			if err := attr.Visit(v); err != nil {
				return err
			}
		}
		for _, child := range n.Children {
			if err := child.Visit(v); err != nil {
				return err
			}
		}
		return nil
	}
	v.RawElement = func(n *parser.RawElement) error {
		for _, attr := range n.Attributes {
			if err := attr.Visit(v); err != nil {
				return err
			}
		}
		return nil
	}
	v.ScriptElement = func(n *parser.ScriptElement) error {
		for _, attr := range n.Attributes {
			if err := attr.Visit(v); err != nil {
				return err
			}
		}
		return nil
	}
	v.BoolConstantAttribute = func(n *parser.BoolConstantAttribute) error {
		return nil
	}
	v.ConstantAttribute = func(n *parser.ConstantAttribute) error {
		return nil
	}
	v.BoolExpressionAttribute = func(n *parser.BoolExpressionAttribute) error {
		return nil
	}
	v.ExpressionAttribute = func(n *parser.ExpressionAttribute) error {
		return nil
	}
	v.SpreadAttributes = func(n *parser.SpreadAttributes) error {
		return nil
	}
	v.ConditionalAttribute = func(n *parser.ConditionalAttribute) error {
		for _, child := range n.Then {
			if err := child.Visit(v); err != nil {
				return err
			}
		}
		for _, child := range n.Else {
			if err := child.Visit(v); err != nil {
				return err
			}
		}
		return nil
	}
	v.GoComment = func(n *parser.GoComment) error {
		return nil
	}
	v.HTMLComment = func(n *parser.HTMLComment) error {
		return nil
	}
	v.CallTemplateExpression = func(n *parser.CallTemplateExpression) error {
		return nil
	}
	v.TemplElementExpression = func(n *parser.TemplElementExpression) error {
		for _, child := range n.Children {
			if err := child.Visit(v); err != nil {
				return err
			}
		}
		return nil
	}
	v.ChildrenExpression = func(n *parser.ChildrenExpression) error {
		return nil
	}
	v.IfExpression = func(n *parser.IfExpression) error {
		for _, child := range n.Then {
			if err := child.Visit(v); err != nil {
				return err
			}
		}
		for _, child := range n.ElseIfs {
			for _, child := range child.Then {
				if err := child.Visit(v); err != nil {
					return err
				}
			}
		}
		for _, child := range n.Else {
			if err := child.Visit(v); err != nil {
				return err
			}
		}
		return nil
	}
	v.SwitchExpression = func(n *parser.SwitchExpression) error {
		for _, node := range n.Cases {
			for _, child := range node.Children {
				if err := child.Visit(v); err != nil {
					return err
				}
			}
		}
		return nil
	}
	v.ForExpression = func(n *parser.ForExpression) error {
		for _, child := range n.Children {
			if err := child.Visit(v); err != nil {
				return err
			}
		}
		return nil
	}
	v.GoCode = func(n *parser.GoCode) error {
		return nil
	}
	v.StringExpression = func(n *parser.StringExpression) error {
		return nil
	}
	v.ScriptTemplate = func(n *parser.ScriptTemplate) error {
		return nil
	}

	return v
}

// Visitor implements the parser.Visitor interface. Each function corresponds to a node type in the parse tree.
// Override these functions to provide custom behavior when visiting nodes.
type Visitor struct {
	TemplateFile             func(n *parser.TemplateFile) error
	TemplateFileGoExpression func(n *parser.TemplateFileGoExpression) error
	Package                  func(n *parser.Package) error
	Whitespace               func(n *parser.Whitespace) error
	CSSTemplate              func(n *parser.CSSTemplate) error
	ConstantCSSProperty      func(n *parser.ConstantCSSProperty) error
	ExpressionCSSProperty    func(n *parser.ExpressionCSSProperty) error
	DocType                  func(n *parser.DocType) error
	HTMLTemplate             func(n *parser.HTMLTemplate) error
	Text                     func(n *parser.Text) error
	Element                  func(n *parser.Element) error
	RawElement               func(n *parser.RawElement) error
	ScriptElement            func(n *parser.ScriptElement) error
	BoolConstantAttribute    func(n *parser.BoolConstantAttribute) error
	ConstantAttribute        func(n *parser.ConstantAttribute) error
	BoolExpressionAttribute  func(n *parser.BoolExpressionAttribute) error
	ExpressionAttribute      func(n *parser.ExpressionAttribute) error
	SpreadAttributes         func(n *parser.SpreadAttributes) error
	ConditionalAttribute     func(n *parser.ConditionalAttribute) error
	GoComment                func(n *parser.GoComment) error
	HTMLComment              func(n *parser.HTMLComment) error
	CallTemplateExpression   func(n *parser.CallTemplateExpression) error
	TemplElementExpression   func(n *parser.TemplElementExpression) error
	ChildrenExpression       func(n *parser.ChildrenExpression) error
	IfExpression             func(n *parser.IfExpression) error
	SwitchExpression         func(n *parser.SwitchExpression) error
	ForExpression            func(n *parser.ForExpression) error
	GoCode                   func(n *parser.GoCode) error
	StringExpression         func(n *parser.StringExpression) error
	ScriptTemplate           func(n *parser.ScriptTemplate) error
}

var _ parser.Visitor = (*Visitor)(nil)

func (v *Visitor) VisitTemplateFile(n *parser.TemplateFile) error {
	return v.TemplateFile(n)
}

func (v *Visitor) VisitTemplateFileGoExpression(n *parser.TemplateFileGoExpression) error {
	return v.TemplateFileGoExpression(n)
}

func (v *Visitor) VisitPackage(n *parser.Package) error {
	return v.Package(n)
}

func (v *Visitor) VisitWhitespace(n *parser.Whitespace) error {
	return v.Whitespace(n)
}

func (v *Visitor) VisitCSSTemplate(n *parser.CSSTemplate) error {
	return v.CSSTemplate(n)
}
func (v *Visitor) VisitConstantCSSProperty(n *parser.ConstantCSSProperty) error {
	return v.ConstantCSSProperty(n)
}

func (v *Visitor) VisitExpressionCSSProperty(n *parser.ExpressionCSSProperty) error {
	return v.ExpressionCSSProperty(n)
}

func (v *Visitor) VisitDocType(n *parser.DocType) error {
	return v.DocType(n)
}

func (v *Visitor) VisitHTMLTemplate(n *parser.HTMLTemplate) error {
	return v.HTMLTemplate(n)
}

func (v *Visitor) VisitText(n *parser.Text) error {
	return v.Text(n)
}

func (v *Visitor) VisitElement(n *parser.Element) error {
	return v.Element(n)
}

func (v *Visitor) VisitRawElement(n *parser.RawElement) error {
	return v.RawElement(n)
}

func (v *Visitor) VisitScriptElement(n *parser.ScriptElement) error {
	return v.ScriptElement(n)
}

func (v *Visitor) VisitBoolConstantAttribute(n *parser.BoolConstantAttribute) error {
	return v.BoolConstantAttribute(n)
}

func (v *Visitor) VisitConstantAttribute(n *parser.ConstantAttribute) error {
	return v.ConstantAttribute(n)
}

func (v *Visitor) VisitBoolExpressionAttribute(n *parser.BoolExpressionAttribute) error {
	return v.BoolExpressionAttribute(n)
}

func (v *Visitor) VisitExpressionAttribute(n *parser.ExpressionAttribute) error {
	return v.ExpressionAttribute(n)
}

func (v *Visitor) VisitSpreadAttributes(n *parser.SpreadAttributes) error {
	return v.SpreadAttributes(n)
}

func (v *Visitor) VisitConditionalAttribute(n *parser.ConditionalAttribute) error {
	return v.ConditionalAttribute(n)
}

func (v *Visitor) VisitGoComment(n *parser.GoComment) error {
	return v.GoComment(n)
}

func (v *Visitor) VisitHTMLComment(n *parser.HTMLComment) error {
	return v.HTMLComment(n)
}

func (v *Visitor) VisitCallTemplateExpression(n *parser.CallTemplateExpression) error {
	return v.CallTemplateExpression(n)
}

func (v *Visitor) VisitTemplElementExpression(n *parser.TemplElementExpression) error {
	return v.TemplElementExpression(n)
}

func (v *Visitor) VisitChildrenExpression(n *parser.ChildrenExpression) error {
	return v.ChildrenExpression(n)
}

func (v *Visitor) VisitIfExpression(n *parser.IfExpression) error {
	return v.IfExpression(n)
}

func (v *Visitor) VisitSwitchExpression(n *parser.SwitchExpression) error {
	return v.SwitchExpression(n)
}

func (v *Visitor) VisitForExpression(n *parser.ForExpression) error {
	return v.ForExpression(n)
}

func (v *Visitor) VisitGoCode(n *parser.GoCode) error {
	return v.GoCode(n)
}

func (v *Visitor) VisitStringExpression(n *parser.StringExpression) error {
	return v.StringExpression(n)
}

func (v *Visitor) VisitScriptTemplate(n *parser.ScriptTemplate) error {
	return v.ScriptTemplate(n)
}
