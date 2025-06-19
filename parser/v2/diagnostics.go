package parser

import (
	"errors"
	"strings"
)

type diagnoser func(Node) ([]Diagnostic, error)
type templateDiagnoser func(*TemplateFile) ([]Diagnostic, error)

// Diagnostic for template file.
type Diagnostic struct {
	Message string
	Range   Range
}

func walkTemplate(t *TemplateFile, f func(Node) bool) {
	for _, n := range t.Nodes {
		hn, ok := n.(*HTMLTemplate)
		if !ok {
			continue
		}
		walkNodes(hn.Children, f)
	}
}
func walkNodes(t []Node, f func(Node) bool) {
	for _, n := range t {
		if !f(n) {
			continue
		}
		if h, ok := n.(CompositeNode); ok {
			walkNodes(h.ChildNodes(), f)
		}
	}
}

var diagnosers = []diagnoser{
	useOfLegacyCallSyntaxDiagnoser,
}

var templateDiagnosers = []templateDiagnoser{
	missingComponentDiagnoser,
}

func Diagnose(t *TemplateFile) ([]Diagnostic, error) {
	var diags []Diagnostic
	var errs error
	
	// Run node-level diagnosers
	walkTemplate(t, func(n Node) bool {
		for _, d := range diagnosers {
			diag, err := d(n)
			if err != nil {
				errs = errors.Join(errs, err)
				return false
			}
			diags = append(diags, diag...)
		}
		return true
	})
	
	// Run template-level diagnosers
	for _, d := range templateDiagnosers {
		diag, err := d(t)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		diags = append(diags, diag...)
	}
	
	return diags, errs
}

func useOfLegacyCallSyntaxDiagnoser(n Node) ([]Diagnostic, error) {
	if c, ok := n.(*CallTemplateExpression); ok {
		return []Diagnostic{{
			Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
			Range:   c.Expression.Range,
		}}, nil
	}
	return nil, nil
}

func missingComponentDiagnoser(t *TemplateFile) ([]Diagnostic, error) {
	var diags []Diagnostic
	
	// Collect all component references in the template
	componentRefs := collectComponentReferences(t)
	
	// Find all defined components in this template
	definedComponents := collectDefinedComponents(t)
	
	// Check each component reference
	for _, ref := range componentRefs {
		// Skip components with package prefixes (like pkg.Component) 
		// as we can't validate those without the full import context
		if strings.Contains(ref.Name, ".") {
			continue
		}
		
		// Check if component is defined in this template
		if !definedComponents[ref.Name] {
			diags = append(diags, Diagnostic{
				Message: "Component " + ref.Name + " not found",
				Range: Range{
					From: ref.Position,
					To: Position{
						Index: ref.Position.Index + int64(len(ref.Name)),
						Line:  ref.Position.Line,
						Col:   ref.Position.Col + uint32(len(ref.Name)),
					},
				},
			})
		}
	}
	
	return diags, nil
}

// componentRef represents a component reference for validation
type componentRef struct {
	Name     string
	Position Position
}

// collectComponentReferences finds all ElementComponent references in the template
func collectComponentReferences(t *TemplateFile) []componentRef {
	var refs []componentRef
	
	walkTemplate(t, func(n Node) bool {
		if comp, ok := n.(*ElementComponent); ok {
			refs = append(refs, componentRef{
				Name:     comp.Name,
				Position: comp.NameRange.From,
			})
		}
		return true
	})
	
	return refs
}

// collectDefinedComponents finds all HTMLTemplate definitions in the template file
func collectDefinedComponents(t *TemplateFile) map[string]bool {
	defined := make(map[string]bool)
	
	for _, node := range t.Nodes {
		if tmpl, ok := node.(*HTMLTemplate); ok {
			name := extractTemplateName(tmpl.Expression.Value)
			if name != "" {
				defined[name] = true
			}
		}
	}
	
	return defined
}

// extractTemplateName extracts the template name from a template expression
// e.g., "Foo()" -> "Foo", "(x Data) Bar()" -> "Bar"
func extractTemplateName(expr string) string {
	// Remove leading/trailing whitespace
	expr = strings.TrimSpace(expr)
	
	// Find the function name by looking for the opening parenthesis
	parenIndex := strings.Index(expr, "(")
	if parenIndex == -1 {
		return ""
	}
	
	// Get everything before the opening parenthesis
	beforeParen := strings.TrimSpace(expr[:parenIndex])
	
	// Split by spaces to handle receiver syntax like "(x Data) Foo"
	parts := strings.Fields(beforeParen)
	if len(parts) == 0 {
		return ""
	}
	
	// The function name is the last part
	return parts[len(parts)-1]
}
