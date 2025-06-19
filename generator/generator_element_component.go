package generator

import (
	"fmt"
	goparser "go/parser"
	"go/token"
	"slices"
	"strings"

	_ "embed"

	"github.com/a-h/templ/parser/v2"
)

func (g *generator) writeElementComponent(indentLevel int, n *parser.ElementComponent) (err error) {
	if len(n.Children) == 0 {
		return g.writeSelfClosingElementComponent(indentLevel, n)
	}
	return g.writeBlockElementComponent(indentLevel, n)
}

func (g *generator) writeSelfClosingElementComponent(indentLevel int, n *parser.ElementComponent) (err error) {
	// templ_7745c5c3_Err = Component(arg1, arg2, ...)
	if err = g.writeElementComponentFunctionCall(indentLevel, n); err != nil {
		return err
	}
	// .Render(ctx, templ_7745c5c3_Buffer)
	if _, err = g.w.Write(".Render(ctx, templ_7745c5c3_Buffer)\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeBlockElementComponent(indentLevel int, n *parser.ElementComponent) (err error) {
	childrenName := g.createVariableName()
	if _, err = g.w.WriteIndent(indentLevel, childrenName+" := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {\n"); err != nil {
		return err
	}
	indentLevel++
	if _, err = g.w.WriteIndent(indentLevel, "templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context\n"); err != nil {
		return err
	}
	if err := g.writeTemplBuffer(indentLevel); err != nil {
		return err
	}
	// ctx = templ.InitializeContext(ctx)
	if _, err = g.w.WriteIndent(indentLevel, "ctx = templ.InitializeContext(ctx)\n"); err != nil {
		return err
	}
	if err = g.writeNodes(indentLevel, stripLeadingAndTrailingWhitespace(n.Children), nil); err != nil {
		return err
	}
	// return nil
	if _, err = g.w.WriteIndent(indentLevel, "return nil\n"); err != nil {
		return err
	}
	indentLevel--
	if _, err = g.w.WriteIndent(indentLevel, "})\n"); err != nil {
		return err
	}
	if err = g.writeElementComponentFunctionCall(indentLevel, n); err != nil {
		return err
	}

	// .Render(templ.WithChildren(ctx, children), templ_7745c5c3_Buffer)
	if _, err = g.w.Write(".Render(templ.WithChildren(ctx, " + childrenName + "), templ_7745c5c3_Buffer)\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

type elementComponentAttributes struct {
	keys      []parser.ConstantAttributeKey
	attrs     []parser.Attribute
	params    []ParameterInfo
	restAttrs []parser.Attribute
	restParam ParameterInfo
}

func (g *generator) writeElementComponentAttrVars(indentLevel int, sigs *ComponentSignature, n *parser.ElementComponent) ([]string, error) {
	orderedAttrs, err := g.reorderElementComponentAttributes(sigs, n)
	if err != nil {
		return nil, err
	}

	var restVarName string
	if orderedAttrs.restParam.Name != "" {
		restVarName = g.createVariableName()
		if _, err = g.w.WriteIndent(indentLevel, "var "+restVarName+" = templ.OrderedAttributes{}\n"); err != nil {
			return nil, err
		}
	}

	res := make([]string, len(orderedAttrs.attrs))
	for i, attr := range orderedAttrs.attrs {
		param := orderedAttrs.params[i]
		value, err := g.writeElementComponentArgNewVar(indentLevel, attr, param)
		if err != nil {
			return nil, err
		}
		res[i] = value
	}

	if orderedAttrs.restParam.Name != "" {
		// spew.Dump(orderedAttrs.restParam, orderedAttrs.restAttrs)
		for _, attr := range orderedAttrs.restAttrs {
			_ = g.writeElementComponentArgRestVar(indentLevel, restVarName, attr)
		}
		res = append(res, restVarName)
	}
	return res, nil
}

func (g *generator) reorderElementComponentAttributes(sig *ComponentSignature, n *parser.ElementComponent) (elementComponentAttributes, error) {
	rest := make([]parser.Attribute, 0)
	attrMap := make(map[string]parser.Attribute)
	keyMap := make(map[string]parser.ConstantAttributeKey)
	for _, attr := range n.Attributes {
		keyed, ok := attr.(parser.KeyedAttribute)
		if ok {
			key, ok := keyed.AttributeKey().(parser.ConstantAttributeKey)
			if ok {
				if slices.ContainsFunc(sig.Parameters, func(p ParameterInfo) bool { return p.Name == key.Name }) {
					// Element component should only works with const key element.
					attrMap[key.Name] = attr
					keyMap[key.Name] = key
					continue
				}
			}
		}
		rest = append(rest, attr)
	}

	params := sig.Parameters
	// We support an optional last parameter that is of type templ.Attributer.
	var attrParam ParameterInfo
	if len(params) > 0 && isTemplAttributer(params[len(params)-1].Type) {
		attrParam = params[len(params)-1]
		params = params[:len(params)-1]
	}
	ordered := make([]parser.Attribute, len(params))
	keys := make([]parser.ConstantAttributeKey, len(params))
	for i, param := range params {
		var ok bool
		ordered[i], ok = attrMap[param.Name]
		if !ok {
			return elementComponentAttributes{}, fmt.Errorf("missing required attribute %s for component %s", param.Name, n.Name)
		}
		keys[i], ok = keyMap[param.Name]
		if !ok {
			return elementComponentAttributes{}, fmt.Errorf("missing required key for attribute %s in component %s", param.Name, n.Name)
		}
	}
	return elementComponentAttributes{
		params:    sig.Parameters,
		attrs:     ordered,
		keys:      keys,
		restAttrs: rest,
		restParam: attrParam,
	}, nil
}

func (g *generator) writeElementComponentAttrComponent(indentLevel int, attr parser.Attribute, param ParameterInfo) (varName string, err error) {
	switch attr := attr.(type) {
	case *parser.InlineComponentAttribute:
		return g.writeChildrenComponent(indentLevel, attr.Children)
	case *parser.ExpressionAttribute:
		varName = g.createVariableName()
		var r parser.Range
		if _, err = g.w.WriteIndent(indentLevel, varName+", templ_7745c5c3_Err := templ.JoinAnyErrs("); err != nil {
			return "", err
		}
		if r, err = g.w.Write(attr.Expression.Value); err != nil {
			return "", err
		}
		g.sourceMap.Add(attr.Expression, r)
		if _, err = g.w.Write(")\n"); err != nil {
			return "", err
		}
		if err = g.writeExpressionErrorHandler(indentLevel, attr.Expression); err != nil {
			return "", err
		}
		return fmt.Sprintf("templ.Stringable(%s)", varName), nil
	case *parser.ConstantAttribute:
		value := `"` + attr.Value + `"`
		if attr.SingleQuote {
			value = `'` + attr.Value + `'`
		}
		varName = g.createVariableName()
		if _, err = g.w.WriteIndent(indentLevel, varName+" := templ.Stringable("+value+")\n"); err != nil {
			return "", err
		}
		return varName, nil
	default:
		return "", fmt.Errorf("unsupported attribute type %T for templ.Component parameter", attr)
	}
}

func (g *generator) writeElementComponentArgNewVar(indentLevel int, attr parser.Attribute, param ParameterInfo) (string, error) {
	if isTemplComponent(param.Type) {
		return g.writeElementComponentAttrComponent(indentLevel, attr, param)
	}

	switch attr := attr.(type) {
	case *parser.ConstantAttribute:
		quote := `"`
		if attr.SingleQuote {
			quote = `'`
		}
		value := quote + attr.Value + quote
		return value, nil
	case *parser.ExpressionAttribute:
		// TODO: support URL, Script and Style attribute
		// check writeExpressionAttribute
		var r parser.Range
		var err error
		vn := g.createVariableName()
		// vn, templ_7745c5c3_Err := templ.JoinAnyErrs(
		if _, err = g.w.WriteIndent(indentLevel, vn+", templ_7745c5c3_Err := templ.JoinAnyErrs("); err != nil {
			return "", err
		}
		// p.Name()
		if r, err = g.w.Write(attr.Expression.Value); err != nil {
			return "", err
		}
		g.sourceMap.Add(attr.Expression, r)
		if _, err = g.w.Write(")\n"); err != nil {
			return "", err
		}
		// Error handler
		if err = g.writeExpressionErrorHandler(indentLevel, attr.Expression); err != nil {
			return "", err
		}
		return vn, nil
	case *parser.BoolConstantAttribute:
		return "true", nil
	case *parser.BoolExpressionAttribute:
		// For boolean expressions that might return errors, use JoinAnyErrs
		vn := g.createVariableName()
		var err error
		// vn, templ_7745c5c3_Err := templ.JoinAnyErrs(expression)
		if _, err = g.w.WriteIndent(indentLevel, vn+", templ_7745c5c3_Err := templ.JoinAnyErrs("); err != nil {
			return "", err
		}
		var r parser.Range
		if r, err = g.w.Write(attr.Expression.Value); err != nil {
			return "", err
		}
		g.sourceMap.Add(attr.Expression, r)
		if _, err = g.w.Write(")\n"); err != nil {
			return "", err
		}
		// Error handler
		if err = g.writeExpressionErrorHandler(indentLevel, attr.Expression); err != nil {
			return "", err
		}
		return vn, nil
	default:
		return "", fmt.Errorf("unsupported attribute type %T in Element component argument", attr)
	}
}

func (g *generator) writeElementComponentArgRestVar(indentLevel int, restVarName string, attr parser.Attribute) error {
	var err error
	switch attr := attr.(type) {
	case *parser.BoolConstantAttribute:
		if err = g.writeRestAppend(indentLevel, restVarName, attr.Key.String(), "true"); err != nil {
			return err
		}
	case *parser.ConstantAttribute:
		value := `"` + attr.Value + `"`
		if attr.SingleQuote {
			value = `'` + attr.Value + `'`
		}
		if err = g.writeRestAppend(indentLevel, restVarName, attr.Key.String(), value); err != nil {
			return err
		}
	case *parser.BoolExpressionAttribute:
		if _, err = g.w.WriteIndent(indentLevel, `if `); err != nil {
			return err
		}
		if _, err = g.w.Write(attr.Expression.Value); err != nil {
			return err
		}
		if _, err = g.w.Write(" {\n"); err != nil {
			return err
		}
		{
			indentLevel++
			if err := g.writeRestAppend(indentLevel, restVarName, attr.Key.String(), "true"); err != nil {
				return err
			}
			indentLevel--
		}
		if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
			return err
		}
	case *parser.ExpressionAttribute:
		attrKey := attr.Key.String()
		if isScriptAttribute(attrKey) {
			vn := g.createVariableName()
			if _, err = g.w.WriteIndent(indentLevel, "var "+vn+" templ.ComponentScript = "); err != nil {
				return err
			}
			var r parser.Range
			if r, err = g.w.Write(attr.Expression.Value); err != nil {
				return err
			}
			g.sourceMap.Add(attr.Expression, r)
			if _, err = g.w.Write("\n"); err != nil {
				return err
			}
			if err := g.writeRestAppend(indentLevel, restVarName, attrKey, vn+".Call"); err != nil {
				return err
			}
		} else if attrKey == "style" {
			var r parser.Range
			vn := g.createVariableName()
			// var vn string
			if _, err = g.w.WriteIndent(indentLevel, "var "+vn+" string\n"); err != nil {
				return err
			}
			// vn, templ_7745c5c3_Err = templruntime.SanitizeStyleAttributeValues(
			if _, err = g.w.WriteIndent(indentLevel, vn+", templ_7745c5c3_Err = templruntime.SanitizeStyleAttributeValues("); err != nil {
				return err
			}
			// value
			if r, err = g.w.Write(attr.Expression.Value); err != nil {
				return err
			}
			g.sourceMap.Add(attr.Expression, r)
			// )
			if _, err = g.w.Write(")\n"); err != nil {
				return err
			}
			if err = g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
			if err = g.writeRestAppend(indentLevel, restVarName, attrKey, vn); err != nil {
				return err
			}
		} else {
			vn := g.createVariableName()
			var r parser.Range
			if r, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("%s, templ_7745c5c3_Err := templ.JoinAnyErrs(%s)\n", vn, attr.Expression.Value)); err != nil {
				return err
			}
			g.sourceMap.Add(attr.Expression, r)
			if err := g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
			if err = g.writeRestAppend(indentLevel, restVarName, attr.Key.String(), vn); err != nil {
				return err
			}
		}
	case *parser.ConditionalAttribute:
		if _, err = g.w.WriteIndent(indentLevel, `if `); err != nil {
			return err
		}
		var r parser.Range
		if r, err = g.w.Write(attr.Expression.Value); err != nil {
			return err
		}
		g.sourceMap.Add(attr.Expression, r)
		if _, err = g.w.Write(" {\n"); err != nil {
			return err
		}
		{
			indentLevel++
			for _, attr := range attr.Then {
				if err := g.writeElementComponentArgRestVar(indentLevel, restVarName, attr); err != nil {
					return err
				}
			}
			indentLevel--
		}
		if len(attr.Else) > 0 {
			if _, err = g.w.WriteIndent(indentLevel, "} else {\n"); err != nil {
				return err
			}
			{
				indentLevel++
				for _, attr := range attr.Else {
					if err := g.writeElementComponentArgRestVar(indentLevel, restVarName, attr); err != nil {
						return err
					}
				}
			}
		}
		if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
			return err
		}
	case *parser.SpreadAttributes:
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("%s = append(%s, %s.Items()...)\n", restVarName, restVarName, attr.Expression.Value)); err != nil {
			return err
		}
	case *parser.AttributeComment:
		return nil
	default:
		return fmt.Errorf("TODO: support attribute type %T in Element component argument", attr)
	}
	return err
}

func (g *generator) writeRestAppend(indentLevel int, restVarName string, key string, val string) error {
	_, err := g.w.WriteIndent(indentLevel,
		fmt.Sprintf("%s = append(%s, templ.KeyValue[string, any]{Key: \"%s\", Value: %s})\n",
			restVarName, restVarName, key, val))
	return err
}

func (g *generator) writeElementComponentFunctionCall(indentLevel int, n *parser.ElementComponent) (err error) {
	sigs, ok := g.componentSigs.Get(n.Name)
	if !ok {
		return fmt.Errorf("component %s signature not found at %s:%d:%d", n.Name, g.options.FileName, n.Range.From.Line, n.Range.From.Col)
	}

	var vars []string
	if vars, err = g.writeElementComponentAttrVars(indentLevel, &sigs, n); err != nil {
		return err
	}

	if _, err = g.w.WriteIndent(indentLevel, `templ_7745c5c3_Err = `); err != nil {
		return err
	}

	var r parser.Range

	// For types that implement Component, use appropriate struct literal syntax
	if sigs.IsStruct {
		// (ComponentType{Field1: value1, Field2: value2}) or (&ComponentType{...})
		if _, err = g.w.Write("("); err != nil {
			return err
		}

		if sigs.IsPointerRecv {
			if _, err = g.w.Write("&"); err != nil {
				return err
			}
		}
		if r, err = g.w.Write(n.Name); err != nil {
			return err
		}
		g.sourceMap.Add(parser.Expression{Value: n.Name, Range: n.NameRange}, r)

		if _, err = g.w.Write("{"); err != nil {
			return err
		}

		// Write field assignments for struct literal
		for i, arg := range vars {
			if i > 0 {
				if _, err = g.w.Write(", "); err != nil {
					return err
				}
			}
			// Write field name: value
			if i < len(sigs.Parameters) {
				if _, err = g.w.Write(sigs.Parameters[i].Name); err != nil {
					return err
				}
				if _, err = g.w.Write(": "); err != nil {
					return err
				}
			}
			if _, err = g.w.Write(arg); err != nil {
				return err
			}
		}

		if _, err = g.w.Write("})"); err != nil {
			return err
		}
	} else {
		// For functions, use function call syntax
		if r, err = g.w.Write(n.Name); err != nil {
			return err
		}
		g.sourceMap.Add(parser.Expression{Value: n.Name, Range: n.NameRange}, r)

		if _, err = g.w.Write("("); err != nil {
			return err
		}

		for i, arg := range vars {
			if i > 0 {
				if _, err = g.w.Write(", "); err != nil {
					return err
				}
			}
			r, err := g.w.Write(arg)
			if err != nil {
				return err
			}
			_ = r // TODO: Add source map for the key
		}

		if _, err = g.w.Write(")"); err != nil {
			return err
		}
	}

	return nil
}

// tryResolveStructMethod attempts to resolve struct method components like structComp.Page to StructComponent.Page
func (g *generator) tryResolveStructMethod(componentName string) bool {
	parts := strings.Split(componentName, ".")
	if len(parts) < 2 {
		return false
	}

	varName := parts[0]
	methodName := strings.Join(parts[1:], ".")

	// Look through the template file for variable declarations
	for _, node := range g.tf.Nodes {
		if goExpr, ok := node.(*parser.TemplateFileGoExpression); ok {
			// Check if this contains variable declarations
			if g.containsVariableDeclaration(goExpr.Expression.Value, varName) {
				typeName := g.extractVariableType(goExpr.Expression.Value, varName)
				if typeName != "" {
					// Look for signature with TypeName.MethodName
					candidateSig := typeName + "." + methodName
					if _, ok := g.templResolver.Get(candidateSig); ok {
						// Add alias mapping for future lookups
						g.templResolver.AddAlias(componentName, candidateSig)
						return true
					}
				}
			}
		}
	}

	return false
}

// containsVariableDeclaration checks if the Go code contains a variable declaration
func (g *generator) containsVariableDeclaration(goCode, varName string) bool {
	// Simple pattern matching for "var varName" or "varName :="
	patterns := []string{
		"var " + varName + " ",
		varName + " :=",
		varName + " =",
	}

	for _, pattern := range patterns {
		if strings.Contains(goCode, pattern) {
			return true
		}
	}
	return false
}

// extractVariableType extracts the type from a variable declaration
func (g *generator) extractVariableType(goCode, varName string) string {
	lines := strings.Split(goCode, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Handle "var varName TypeName"
		if strings.HasPrefix(line, "var "+varName+" ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[2]
			}
		}

		// Handle "varName := TypeName{}" or "varName = TypeName{}"
		if strings.Contains(line, varName+" :=") || strings.Contains(line, varName+" =") {
			// Extract type from constructor call like "StructComponent{}"
			if idx := strings.Index(line, "{"); idx != -1 {
				beforeBrace := line[:idx]
				parts := strings.Fields(beforeBrace)
				if len(parts) >= 3 {
					return parts[len(parts)-1]
				}
			}
		}
	}

	return ""
}

func (g *generator) resolveImportPath(packageAlias string) string {
	fset := token.NewFileSet()

	// Look through the template file's imports to find the import path for this alias
	for _, node := range g.tf.Nodes {
		if importNode, ok := node.(*parser.TemplateFileGoExpression); ok {
			// Check if this contains import statements
			if strings.Contains(importNode.Expression.Value, "import") {
				path := g.parseImportPathWithAST(importNode.Expression.Value, packageAlias, fset)
				if path != "" {
					return path
				}
			}
		}
	}
	return ""
}

// parseImportPathWithAST extracts the import path for a specific alias using Go AST parser
func (g *generator) parseImportPathWithAST(goCode, packageAlias string, fset *token.FileSet) string {
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
			return g.parseImportPathFallback(goCode, packageAlias)
		}
	}

	// Extract import path for the specific alias from AST
	for _, imp := range astFile.Imports {
		if imp.Path != nil {
			pkgPath := strings.Trim(imp.Path.Value, `"`)
			var alias string

			if imp.Name != nil {
				// Explicit alias: import alias "path"
				alias = imp.Name.Name
			} else {
				// No explicit alias: import "path" -> derive alias from path
				if lastSlash := strings.LastIndex(pkgPath, "/"); lastSlash != -1 {
					alias = pkgPath[lastSlash+1:]
				}
			}

			if alias == packageAlias {
				return pkgPath
			}
		}
	}

	return ""
}

// parseImportPathFallback provides fallback parsing for edge cases
func (g *generator) parseImportPathFallback(goCode, packageAlias string) string {
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
					if alias == packageAlias {
						return pkgPath
					}
				}
			} else {
				// Handle import with explicit alias
				// alias "package" or . "package"
				parts := strings.Fields(importPart)
				if len(parts) >= 2 {
					alias := parts[0]
					if alias == packageAlias {
						return strings.Trim(parts[1], `"`)
					}
				}
			}
		}
	}
	return ""
}

func (g *generator) collectAndResolveComponents() error {
	collector := NewElementComponentCollector()
	_ = collector.Collect(g.tf)

	uniqueComponents := collector.GetUniqueComponents()

	if len(uniqueComponents) == 0 {
		return nil
	}

	for _, comp := range uniqueComponents {
		var sig ComponentSignature
		var err error
		var found bool

		if comp.PackageName == "" {
			// Local component - check both simple name and full name for receiver methods
			if templSig, ok := g.templResolver.Get(comp.Name); ok {
				sig = templSig
				found = true
			} else {
				// If this looks like a struct method (var.Method), try to resolve it
				if strings.Contains(comp.Name, ".") {
					found = g.tryResolveStructMethod(comp.Name)
					if found {
						// Find the resolved signature
						if templSig, ok := g.templResolver.Get(comp.Name); ok {
							sig = templSig
						}
					}
				}

				if !found && g.symbolResolverEnabled {
					// Try Go function resolution
					sig, err = (&g.symbolResolver).ResolveLocalComponent(comp.Name, comp.Position, g.options.FileName)
					if err == nil {
						found = true
					}
				}
			}
		} else {
			// Package import - use Go function resolution with resolved import path
			if g.symbolResolverEnabled {
				importPath := g.resolveImportPath(comp.PackageName)
				if importPath != "" {
					sig, err = (&g.symbolResolver).ResolveComponentWithPosition(importPath, comp.Name, comp.Position, g.options.FileName)
					if err == nil {
						found = true
					}
				}
			}
		}

		if !found {
			var message string
			if err != nil {
				if comp.PackageName != "" {
					message = fmt.Sprintf("Component %s.%s: %v", comp.PackageName, comp.Name, err)
				} else {
					message = fmt.Sprintf("Component %s: %v", comp.Name, err)
				}
			} else {
				if comp.PackageName != "" {
					message = fmt.Sprintf("Component %s.%s not found in templ templates or Go functions", comp.PackageName, comp.Name)
				} else {
					message = fmt.Sprintf("Component %s not found in templ templates or Go functions", comp.Name)
				}
			}

			// Add diagnostic for this missing component
			g.addComponentDiagnostic(comp, message)
			continue
		}

		// Store the signature for use during code generation
		key := comp.Name
		if comp.PackageName != "" {
			key = comp.PackageName + "." + comp.Name
		}
		sig.QualifiedName = key
		g.componentSigs.Add(sig)
	}

	// Don't return an error if no component signatures are resolved
	// Instead, let the diagnostics be reported to the LSP
	// The code generation will handle missing components gracefully
	// if len(g.componentSigs) == 0 && len(uniqueComponents) > 0 {
	// }

	return nil
}

// addComponentDiagnostic adds a diagnostic for component resolution issues
func (g *generator) addComponentDiagnostic(comp ComponentReference, message string) {
	// Create a Range from the component's position
	// ComponentReference.Position is the start position of the component name
	nameStart := comp.Position
	nameLength := int64(len(comp.Name))
	nameEnd := parser.Position{
		Index: nameStart.Index + nameLength,
		Line:  nameStart.Line,
		Col:   nameStart.Col + uint32(len(comp.Name)),
	}

	g.diagnostics = append(g.diagnostics, parser.Diagnostic{
		Message: message,
		Range: parser.Range{
			From: nameStart,
			To:   nameEnd,
		},
	})
}

func isTemplAttributer(typ string) bool {
	// TODO: better handling of the types:
	// when the type comes from templ file, it will be "templ.Attributer"
	// when it comes from a Go file which uses the x/tool/packages parser the moment, it will be "github.com/a-h/templ.Attributer"
	// This is not ideal but a ok compromise. when the symbols is from templ files, it may not have been resolved, only parsed. And resolving takes time and may nto be available.
	return typ == "templ.Attributer" || typ == "github.com/a-h/templ.Attributer"
}

func isTemplComponent(typ string) bool {
	return typ == "templ.Component" || typ == "github.com/a-h/templ.Component"
}
