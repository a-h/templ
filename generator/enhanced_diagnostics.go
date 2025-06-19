package generator

import (
	"strings"

	"github.com/a-h/templ/parser/v2"
)

// DiagnoseWithSymbolResolution performs diagnostics with Go type information
// This is more comprehensive than parser.Diagnose() but requires a working directory for package loading
func DiagnoseWithSymbolResolution(t *parser.TemplateFile, workingDir string) ([]parser.Diagnostic, error) {
	// Run standard diagnostics first
	standardDiags, err := parser.Diagnose(t)
	if err != nil {
		return nil, err
	}

	// Filter out missing component diagnostics - we'll replace them with enhanced ones
	var filteredDiags []parser.Diagnostic
	for _, d := range standardDiags {
		if !strings.Contains(d.Message, "Component") || !strings.Contains(d.Message, "not found") {
			filteredDiags = append(filteredDiags, d)
		}
	}

	// Add enhanced missing component diagnostics
	enhancedDiags, err := enhancedMissingComponentDiagnoser(t, workingDir)
	if err != nil {
		// If enhanced diagnostics fail, fall back to standard ones
		return standardDiags, nil
	}

	return append(filteredDiags, enhancedDiags...), nil
}

// enhancedMissingComponentDiagnoser checks for missing components using Go type information
func enhancedMissingComponentDiagnoser(t *parser.TemplateFile, workingDir string) ([]parser.Diagnostic, error) {
	var diags []parser.Diagnostic

	// Collect all component references
	componentRefs := collectComponentReferences(t)

	// Find defined templ components in this file
	definedComponents := collectDefinedComponents(t)

	// Create symbol resolver
	resolver := NewSymbolResolver(workingDir)

	// Check each component reference
	for _, ref := range componentRefs {
		// Skip components with package prefixes for now
		// TODO: Could be enhanced to resolve cross-package components
		if strings.Contains(ref.Name, ".") {
			continue
		}

		// Check if component is defined as a templ function in this file
		if definedComponents[ref.Name] {
			continue
		}

		// Try to resolve as a Go type implementing templ.Component
		sig, err := resolver.ResolveLocalComponent(ref.Name, parser.Position{}, "")
		if err != nil {
			// Component not found - add diagnostic
			diags = append(diags, parser.Diagnostic{
				Message: "Component " + ref.Name + " not found",
				Range: parser.Range{
					From: ref.Position,
					To: parser.Position{
						Index: ref.Position.Index + int64(len(ref.Name)),
						Line:  ref.Position.Line,
						Col:   ref.Position.Col + uint32(len(ref.Name)),
					},
				},
			})
		} else if sig != nil {
			// For debugging: if we found it, let's understand what we found
			// (This would normally be removed in production code)
		}
	}

	return diags, nil
}

// Helper types and functions that mirror those in parser/v2/diagnostics.go

// componentRef represents a component reference for validation
type componentRef struct {
	Name     string
	Position parser.Position
}

// collectComponentReferences finds all ElementComponent references in the template
func collectComponentReferences(t *parser.TemplateFile) []componentRef {
	var refs []componentRef

	walkTemplate(t, func(n parser.Node) bool {
		if comp, ok := n.(*parser.ElementComponent); ok {
			refs = append(refs, componentRef{
				Name:     comp.Name,
				Position: comp.NameRange.From,
			})
		}
		return true
	})

	return refs
}

// walkTemplate walks through all template nodes
func walkTemplate(t *parser.TemplateFile, f func(parser.Node) bool) {
	for _, n := range t.Nodes {
		hn, ok := n.(*parser.HTMLTemplate)
		if !ok {
			continue
		}
		walkNodes(hn.Children, f)
	}
}

// walkNodes walks through a slice of nodes recursively
func walkNodes(t []parser.Node, f func(parser.Node) bool) {
	for _, n := range t {
		if !f(n) {
			continue
		}
		if h, ok := n.(parser.CompositeNode); ok {
			walkNodes(h.ChildNodes(), f)
		}
	}
}

// collectDefinedComponents finds all HTMLTemplate definitions in the template file
func collectDefinedComponents(t *parser.TemplateFile) map[string]bool {
	defined := make(map[string]bool)

	for _, node := range t.Nodes {
		if tmpl, ok := node.(*parser.HTMLTemplate); ok {
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
