package parser

import (
	"errors"
	"strings"
)

// ARCHITECTURE NOTE: Diagnostics Duplication
//
// This file contains parser-level diagnostics that operate without external dependencies
// like golang.org/x/tools/packages. These are "fast" diagnostics that can run during
// parsing without needing to load Go modules or resolve import paths.
//
// There is intentional duplication with generator/enhanced_diagnostics.go because:
//
// 1. PARSER LAYER (this file):
//    - Runs during template parsing
//    - No external dependencies (no x/tools/packages)
//    - Fast execution, minimal setup
//    - Uses simple string parsing for Go code analysis
//    - Can only validate local components (same file)
//    - Limited to basic pattern matching for Go structs with Render methods
//
// 2. GENERATOR LAYER (enhanced_diagnostics.go):
//    - Runs during code generation with full context
//    - Uses x/tools/packages for proper Go type information
//    - Slower but more accurate
//    - Can resolve cross-package components and imports
//    - Full type checking and interface implementation validation
//    - Can validate working directory and module structure
//
// The parser diagnostics provide immediate feedback during editing, while the generator
// diagnostics provide comprehensive validation during the build process. Both are needed
// for a good user experience in editors/LSPs vs command-line builds.

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

	// Find all Go types that implement Component interface in this template
	goComponents := collectGoComponents(t)

	// Check each component reference
	for _, ref := range componentRefs {
		// Skip components with package prefixes (like pkg.Component)
		// as we can't validate those without the full import context
		if strings.Contains(ref.Name, ".") {
			continue
		}

		// Check if component is defined in this template (either as templ component or Go struct)
		if !definedComponents[ref.Name] && !goComponents[ref.Name] {
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

// collectGoComponents finds all Go types that implement the Component interface in the template file
// NOTE: This is a simplified version compared to generator/enhanced_diagnostics.go which uses
// x/tools/packages for full type checking. This version uses basic string parsing for speed.
func collectGoComponents(t *TemplateFile) map[string]bool {
	goComponents := make(map[string]bool)

	// Look through Go expressions for type definitions and Render methods
	for _, node := range t.Nodes {
		if goExpr, ok := node.(*TemplateFileGoExpression); ok {
			typeNames := parseGoTypesWithRenderMethod(goExpr.Expression.Value)
			for _, typeName := range typeNames {
				goComponents[typeName] = true
			}
		}
	}

	return goComponents
}

// parseGoTypesWithRenderMethod parses Go code and finds types that have a Render method
// This is a simplified parser that looks for patterns like:
// type Foo struct{} ... func (f *Foo) Render(ctx context.Context, w io.Writer) error
func parseGoTypesWithRenderMethod(goCode string) []string {
	var typeNames []string

	// Split into lines for simple pattern matching
	lines := strings.Split(goCode, "\n")

	// First pass: collect type definitions
	definedTypes := make(map[string]bool)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "type ") && (strings.Contains(line, "struct") || strings.Contains(line, "interface")) {
			// Extract type name
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				typeName := parts[1]
				definedTypes[typeName] = true
			}
		}
	}

	// Second pass: look for Render methods on these types
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "func ") && strings.Contains(line, "Render(") {
			// Look for receiver pattern: func (receiver Type) Render or func (receiver *Type) Render
			parenStart := strings.Index(line, "(")
			parenEnd := strings.Index(line, ")")
			if parenStart != -1 && parenEnd != -1 && parenEnd > parenStart {
				receiver := strings.TrimSpace(line[parenStart+1 : parenEnd])
				if receiver != "" {
					// Parse receiver: "f Foo" or "f *Foo"
					parts := strings.Fields(receiver)
					if len(parts) >= 2 {
						typeName := strings.TrimPrefix(parts[1], "*") // Remove pointer indicator
						if definedTypes[typeName] {
							typeNames = append(typeNames, typeName)
						}
					}
				}
			}
		}
	}

	return typeNames
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
