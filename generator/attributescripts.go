package generator

import (
	"github.com/a-h/templ/internal/htmlattr"
	"github.com/a-h/templ/parser/v2"
)

// attributeScripts holds the attribute expressions of an element that can
// return a templ.ComponentScript. The functions of the scripts are rendered
// before the opening tag of the element, so that they are defined before the
// element calls them.
type attributeScripts struct {
	// EventHandlers are the values of event handler attributes with constant
	// names. The values are always a templ.ComponentScript.
	EventHandlers []string
	// DynamicKeyValues are the values of attributes with names that are only
	// known at render time. The values may be a templ.ComponentScript.
	DynamicKeyValues []parser.Expression
	// Spreads are spread attributes. The values of the attributes may be a
	// templ.ComponentScript.
	Spreads []parser.Expression
}

// Collect adds the script expressions of attr, including those within
// conditional attributes, to s.
func (s *attributeScripts) Collect(attr parser.Attribute) {
	switch attr := attr.(type) {
	case *parser.ConditionalAttribute:
		for _, attr := range attr.Then {
			s.Collect(attr)
		}
		for _, attr := range attr.Else {
			s.Collect(attr)
		}
	case *parser.ExpressionAttribute:
		switch key := attr.Key.(type) {
		case parser.ConstantAttributeKey:
			if htmlattr.Classify("", key.Name) == htmlattr.ContextEventHandler {
				s.EventHandlers = append(s.EventHandlers, attr.Expression.Value)
			}
		case parser.ExpressionAttributeKey:
			s.DynamicKeyValues = append(s.DynamicKeyValues, attr.Expression)
		}
	case *parser.SpreadAttributes:
		s.Spreads = append(s.Spreads, attr.Expression)
	}
}
