package generator

import (
	"strings"

	"github.com/a-h/templ/parser/v2"
)

// ComponentReference represents a reference to a component in HTML Element syntax
type ComponentReference struct {
	Name        string
	PackageName string // empty for local components
	Position    parser.Position
	Attributes  []parser.Attribute
}

// ComponentCollector collects all Element component references from a templ file
type ComponentCollector struct {
	components []ComponentReference
	imports    map[string]bool // Track import aliases
}

// NewElementComponentCollector creates a new component collector
func NewElementComponentCollector() *ComponentCollector {
	return &ComponentCollector{
		components: make([]ComponentReference, 0),
		imports:    make(map[string]bool),
	}
}

// Collect walks the template file and collects all Element component references
func (cc *ComponentCollector) Collect(tf *parser.TemplateFile) []ComponentReference {
	cc.components = make([]ComponentReference, 0)
	cc.imports = make(map[string]bool)

	// First pass: collect imports
	cc.collectImports(tf)

	// Second pass: collect components
	for _, node := range tf.Nodes {
		switch n := node.(type) {
		case *parser.HTMLTemplate:
			cc.collectFromNodes(n.Children)
		case *parser.CSSTemplate:
			// CSS templates don't contain components
		case *parser.ScriptTemplate:
			// Script templates don't contain components
		}
	}

	return cc.components
}

// collectFromNodes recursively collects components from a slice of nodes
func (cc *ComponentCollector) collectFromNodes(nodes []parser.Node) {
	for _, node := range nodes {
		cc.collectFromNode(node)
	}
}

// collectFromNode collects components from a single node
func (cc *ComponentCollector) collectFromNode(node parser.Node) {
	switch n := node.(type) {
	case *parser.ElementComponent:
		// Split component name by dots and check if first part is an import
		parts := strings.Split(n.Name, ".")
		
		var pkgName, componentName string
		
		if len(parts) > 1 && cc.imports[parts[0]] {
			// First part is an import alias, treat as package.Component
			pkgName = parts[0]
			componentName = strings.Join(parts[1:], ".")
		} else {
			// Not an import, treat as local component (could be structVar.Method)
			componentName = n.Name
		}

		cc.components = append(cc.components, ComponentReference{
			Name:        componentName,
			PackageName: pkgName,
			Position:    n.NameRange.From,
			Attributes:  n.Attributes,
		})

		// Recursively collect from children
		cc.collectFromNodes(n.Children)

	case *parser.Element:
		// Regular HTML elements might contain Element components in their children
		cc.collectFromNodes(n.Children)

	case *parser.IfExpression:
		cc.collectFromNodes(n.Then)
		cc.collectFromNodes(n.Else)

	case *parser.SwitchExpression:
		for _, c := range n.Cases {
			cc.collectFromNodes(c.Children)
		}

	case *parser.ForExpression:
		cc.collectFromNodes(n.Children)

	case *parser.CallTemplateExpression:
		// Template calls don't have children in the AST

	case *parser.TemplElementExpression:
		cc.collectFromNodes(n.Children)

		// Add other node types that might contain children
	}
}

// GetUniqueComponents returns unique component references
func (cc *ComponentCollector) GetUniqueComponents() []ComponentReference {
	seen := make(map[string]bool)
	unique := make([]ComponentReference, 0)

	for _, comp := range cc.components {
		key := comp.PackageName + "." + comp.Name
		if !seen[key] {
			seen[key] = true
			unique = append(unique, comp)
		}
	}

	return unique
}

// collectImports extracts import aliases from the template file
func (cc *ComponentCollector) collectImports(tf *parser.TemplateFile) {
	for _, node := range tf.Nodes {
		if importNode, ok := node.(*parser.TemplateFileGoExpression); ok {
			// Check if this contains import statements
			if strings.Contains(importNode.Expression.Value, "import ") {
				cc.parseImportStatements(importNode.Expression.Value)
			}
		}
	}
}

// parseImportStatements extracts import aliases from Go import code
func (cc *ComponentCollector) parseImportStatements(goCode string) {
	lines := strings.Split(goCode, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "import ") {
			// Handle different import formats:
			// import "package"
			// import alias "package" 
			// import . "package"
			
			// Remove "import " prefix
			importPart := strings.TrimSpace(line[7:])
			
			// Handle quoted import without alias
			if strings.HasPrefix(importPart, `"`) && strings.HasSuffix(importPart, `"`) {
				// import "github.com/pkg/name" -> alias is "name"
				pkgPath := importPart[1 : len(importPart)-1]
				if lastSlash := strings.LastIndex(pkgPath, "/"); lastSlash != -1 {
					alias := pkgPath[lastSlash+1:]
					cc.imports[alias] = true
				}
			} else {
				// Handle import with explicit alias
				// alias "package" or . "package"
				parts := strings.Fields(importPart)
				if len(parts) >= 2 {
					alias := parts[0]
					if alias != "." && alias != "_" {
						cc.imports[alias] = true
					}
				}
			}
		}
	}
}

