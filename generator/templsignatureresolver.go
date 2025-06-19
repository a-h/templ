package generator

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"strings"

	"github.com/a-h/templ/parser/v2"
)

// TemplSignatureResolver extracts component signatures from templ template definitions
type TemplSignatureResolver map[string]ComponentSignature

// ExtractSignatures walks through a templ file and extracts all template signatures
func (tsr *TemplSignatureResolver) ExtractSignatures(tf *parser.TemplateFile) {
	for _, node := range tf.Nodes {
		switch n := node.(type) {
		case *parser.HTMLTemplate:
			sig, ok := tsr.extractHTMLTemplateSignature(n)
			if ok {
				tsr.Add(sig)
			}
		case *parser.TemplateFileGoExpression:
			// Extract type definitions that might implement Component
			tsr.extractGoTypeSignatures(n)
		}
	}
}

// Get returns the signature for a template name
func (tsr TemplSignatureResolver) Get(name string) (ComponentSignature, bool) {
	sig, ok := tsr[name]
	return sig, ok
}

// Add adds a signature to the resolver using sig.Name as the key
func (tsr TemplSignatureResolver) Add(sig ComponentSignature) {
	tsr[sig.Name] = sig
}

// GetAllNames returns all signature names for debugging
func (tsr TemplSignatureResolver) GetAllNames() []string {
	names := make([]string, 0, len(tsr))
	for name := range tsr {
		names = append(names, name)
	}
	return names
}

// AddAlias adds an alias mapping for a signature
func (tsr TemplSignatureResolver) AddAlias(alias, target string) {
	if sig, ok := tsr[target]; ok {
		tsr[alias] = sig
	}
}

// extractHTMLTemplateSignature extracts the signature from an HTML template
func (tsr *TemplSignatureResolver) extractHTMLTemplateSignature(tmpl *parser.HTMLTemplate) (ComponentSignature, bool) {
	// Parse the template declaration from Expression.Value using Go AST parser
	// This leverages the same parsing logic used by parseTemplFuncDecl
	exprValue := tmpl.Expression.Value
	if exprValue == "" {
		return ComponentSignature{}, false
	}

	name, params, err := tsr.parseTemplateSignatureFromAST(exprValue)
	if err != nil || name == "" {
		return ComponentSignature{}, false
	}

	return ComponentSignature{
		PackagePath: "", // Local package
		Name:        name,
		Parameters:  params,
	}, true
}

// parseTemplateSignatureFromAST parses a templ template signature using Go AST parser
// This follows the same approach as parseTemplFuncDecl in goparser.go
func (tsr *TemplSignatureResolver) parseTemplateSignatureFromAST(exprValue string) (name string, params []ParameterInfo, err error) {
	// Add "func " prefix to make it a valid Go function declaration for parsing
	// This mirrors what parseTemplFuncDecl does with goexpression.Func
	funcDecl := "func " + exprValue

	// Create a temporary package to parse the function
	src := "package main\n" + funcDecl

	// Parse the source
	fset := token.NewFileSet()
	node, parseErr := goparser.ParseFile(fset, "", src, goparser.AllErrors)
	if parseErr != nil || node == nil {
		return "", nil, parseErr
	}

	// Extract function declaration from AST
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			name = fn.Name.Name
			params = tsr.extractParametersFromAST(fn.Type.Params)

			// If this is a receiver method, create a composite name
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				receiverType := tsr.astTypeToString(fn.Recv.List[0].Type)
				// Remove pointer indicator if present for consistent naming
				receiverType = strings.TrimPrefix(receiverType, "*")
				name = receiverType + "." + name
			}

			return name, params, nil
		}
	}

	return "", nil, nil
}

// extractParametersFromAST extracts parameter information from AST field list
func (tsr *TemplSignatureResolver) extractParametersFromAST(fieldList *ast.FieldList) []ParameterInfo {
	if fieldList == nil || len(fieldList.List) == 0 {
		return nil
	}

	var params []ParameterInfo

	for _, field := range fieldList.List {
		fieldType := tsr.astTypeToString(field.Type)

		// Handle multiple names with the same type (e.g., "a, b string")
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				params = append(params, ParameterInfo{
					Name: name.Name,
					Type: fieldType,
				})
			}
		} else {
			// TODO: Handle anonymous parameters if needed: maybe use the fieldtype, but sanitized?
			params = append(params, ParameterInfo{
				Name: "",
				Type: fieldType,
			})
		}
	}

	return params
}

// astTypeToString converts AST type expressions to their string representation
// This returns the type name as it appears in the Go source code
func (tsr *TemplSignatureResolver) astTypeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		// Basic types like string, int, bool, etc.
		return t.Name
	case *ast.StarExpr:
		// Pointer types like *string
		return "*" + tsr.astTypeToString(t.X)
	case *ast.ArrayType:
		// Array or slice types like []string, [10]int
		if t.Len == nil {
			// Slice
			return "[]" + tsr.astTypeToString(t.Elt)
		} else {
			// Array - e.g., [10]int or [...]string{}
			return "[...]" + tsr.astTypeToString(t.Elt)
		}
	case *ast.MapType:
		// Map types like map[string]int
		return "map[" + tsr.astTypeToString(t.Key) + "]" + tsr.astTypeToString(t.Value)
	case *ast.SelectorExpr:
		// Qualified types like time.Time, context.Context
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.InterfaceType:
		return "any"
	default:
		return ""
	}
}

// extractGoTypeSignatures extracts type definitions from Go code that might implement Component
func (tsr *TemplSignatureResolver) extractGoTypeSignatures(goExpr *parser.TemplateFileGoExpression) {
	// Parse the Go code
	src := "package main\n" + goExpr.Expression.Value
	fset := token.NewFileSet()
	node, err := goparser.ParseFile(fset, "", src, goparser.AllErrors)
	if err != nil || node == nil {
		return
	}

	// Look for type declarations and methods
	typeNames := make(map[string]bool)

	// First pass: collect all type names
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					typeNames[typeSpec.Name.Name] = true
				}
			}
		}
	}

	// Second pass: look for Render methods on these types
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "Render" && fn.Recv != nil {
			// Check if this is a method on one of our types
			if len(fn.Recv.List) > 0 {
				receiverType := tsr.astTypeToString(fn.Recv.List[0].Type)
				receiverType = strings.TrimPrefix(receiverType, "*")

				if typeNames[receiverType] {
					// Check if the signature matches Component.Render
					if tsr.isComponentRenderMethod(fn) {
						// Check if receiver is a pointer
						isPointerRecv := strings.HasPrefix(tsr.astTypeToString(fn.Recv.List[0].Type), "*")

						// This type implements Component
						sig := ComponentSignature{
							PackagePath:   "",
							Name:          receiverType,
							Parameters:    []ParameterInfo{}, // Component types have no parameters
							IsStruct:      true,
							IsPointerRecv: isPointerRecv,
						}
						tsr.Add(sig)
					}
				}
			}
		}
	}
}

// isComponentRenderMethod checks if a function declaration matches the Component.Render signature
func (tsr *TemplSignatureResolver) isComponentRenderMethod(fn *ast.FuncDecl) bool {
	// Check parameters: (ctx context.Context, w io.Writer)
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 2 {
		return false
	}

	// Check return type: error
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}

	// Check if return type is error
	if retType, ok := fn.Type.Results.List[0].Type.(*ast.Ident); !ok || retType.Name != "error" {
		return false
	}

	return true
}
