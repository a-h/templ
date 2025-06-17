package generator

import (
	"strings"

	"github.com/a-h/templ/parser/v2"
)

// ComponentReference represents a reference to a component in JSX syntax
type ComponentReference struct {
	Name        string
	PackageName string // empty for local components
	Position    parser.Position
	Attributes  []parser.Attribute
}

// ComponentCollector collects all JSX component references from a templ file
type ComponentCollector struct {
	components []ComponentReference
}

// NewComponentCollector creates a new component collector
func NewComponentCollector() *ComponentCollector {
	return &ComponentCollector{
		components: make([]ComponentReference, 0),
	}
}

// Collect walks the template file and collects all JSX component references
func (cc *ComponentCollector) Collect(tf *parser.TemplateFile) []ComponentReference {
	cc.components = make([]ComponentReference, 0)
	
	// Walk through all templates in the file
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
	case *parser.JSXComponentElement:
		// Extract package name and component name
		pkgName := ""
		componentName := n.Name
		
		if idx := strings.LastIndex(n.Name, "."); idx != -1 {
			pkgName = n.Name[:idx]
			componentName = n.Name[idx+1:]
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
		// Regular HTML elements might contain JSX components in their children
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