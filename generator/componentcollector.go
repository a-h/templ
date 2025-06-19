package generator

import (
	goparser "go/parser"
	"go/token"
	"strings"

	"github.com/a-h/templ/parser/v2"
	"github.com/a-h/templ/parser/v2/visitor"
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

	// Second pass: collect components using visitor pattern
	v := visitor.New()
	v.ElementComponent = cc.visitElementComponent
	_ = tf.Visit(v)

	return cc.components
}

// visitElementComponent handles ElementComponent nodes using the visitor pattern
func (cc *ComponentCollector) visitElementComponent(n *parser.ElementComponent) error {
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

	// The default visitor implementation will handle visiting children
	return nil
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

// collectImports extracts import aliases from the template file using Go AST parser
func (cc *ComponentCollector) collectImports(tf *parser.TemplateFile) {
	fset := token.NewFileSet()

	for _, node := range tf.Nodes {
		if importNode, ok := node.(*parser.TemplateFileGoExpression); ok {
			// Check if this contains import statements
			if strings.Contains(importNode.Expression.Value, "import") {
				cc.parseImportStatementsWithAST(importNode.Expression.Value, fset)
			}
		}
	}
}

// parseImportStatementsWithAST extracts import aliases from Go code using proper AST parsing
func (cc *ComponentCollector) parseImportStatementsWithAST(goCode string, fset *token.FileSet) {
	// Try to parse as a complete Go file first
	fullGoCode := "package main\n" + goCode

	astFile, err := goparser.ParseFile(fset, "", fullGoCode, goparser.ImportsOnly)
	if err != nil {
		// If that fails, try parsing just the import block
		if strings.Contains(goCode, "import (") {
			// Extract just the import block
			start := strings.Index(goCode, "import (")
			if start != -1 {
				end := strings.Index(goCode[start:], ")")
				if end != -1 {
					importBlock := goCode[start : start+end+1]
					fullGoCode = "package main\n" + importBlock
					astFile, err = goparser.ParseFile(fset, "", fullGoCode, goparser.ImportsOnly)
				}
			}
		}

		if err != nil {
			// Fall back to simple string parsing for edge cases
			cc.parseImportStatementsFallback(goCode)
			return
		}
	}

	// Extract import aliases from AST
	for _, imp := range astFile.Imports {
		if imp.Path != nil {
			pkgPath := strings.Trim(imp.Path.Value, `"`)
			var alias string

			if imp.Name != nil {
				// Explicit alias: import alias "path"
				alias = imp.Name.Name
				if alias != "." && alias != "_" {
					cc.imports[alias] = true
				}
			} else {
				// No explicit alias: import "path" -> derive alias from path
				if lastSlash := strings.LastIndex(pkgPath, "/"); lastSlash != -1 {
					alias = pkgPath[lastSlash+1:]
					cc.imports[alias] = true
				}
			}
		}
	}
}

// parseImportStatementsFallback provides fallback parsing for edge cases
func (cc *ComponentCollector) parseImportStatementsFallback(goCode string) {
	lines := strings.Split(goCode, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "import ") {
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
