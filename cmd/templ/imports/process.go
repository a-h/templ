package imports

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"path"
	"slices"
	"strconv"
	"strings"

	goparser "go/parser"

	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"

	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
)

var internalImports = []string{"github.com/a-h/templ", "github.com/a-h/templ/runtime"}

func convertTemplToGoURI(templURI string) (isTemplFile bool, goURI string) {
	base, fileName := path.Split(templURI)
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}
	return true, base + (strings.TrimSuffix(fileName, ".templ") + "_templ.go")
}

var fset = token.NewFileSet()

func updateImports(name, src string) (updated []*ast.ImportSpec, err error) {
	// Apply auto imports.
	updatedGoCode, err := imports.Process(name, []byte(src), nil)
	if err != nil {
		return updated, fmt.Errorf("failed to process go code %q: %w", src, err)
	}
	// Get updated imports.
	gofile, err := goparser.ParseFile(fset, name, updatedGoCode, goparser.ImportsOnly)
	if err != nil {
		return updated, fmt.Errorf("failed to get imports from updated go code: %w", err)
	}
	for _, imp := range gofile.Imports {
		if !slices.Contains(internalImports, strings.Trim(imp.Path.Value, "\"")) {
			updated = append(updated, imp)
		}
	}
	return updated, nil
}

func Process(t parser.TemplateFile) (parser.TemplateFile, error) {
	if t.Filepath == "" {
		return t, nil
	}
	isTemplFile, fileName := convertTemplToGoURI(t.Filepath)
	if !isTemplFile {
		return t, fmt.Errorf("invalid filepath: %s", t.Filepath)
	}

	// The first node always contains existing imports.
	// If there isn't one, create it.
	if len(t.Nodes) == 0 {
		t.Nodes = append(t.Nodes, parser.TemplateFileGoExpression{})
	}
	// If there is one, ensure it is a Go expression.
	if _, ok := t.Nodes[0].(parser.TemplateFileGoExpression); !ok {
		t.Nodes = append([]parser.TemplateFileNode{parser.TemplateFileGoExpression{}}, t.Nodes...)
	}

	// Find all existing imports.
	importsNode := t.Nodes[0].(parser.TemplateFileGoExpression)

	// Generate code.
	gw := bytes.NewBuffer(nil)
	var updatedImports []*ast.ImportSpec
	var eg errgroup.Group
	eg.Go(func() (err error) {
		if _, err := generator.Generate(t, gw); err != nil {
			return fmt.Errorf("failed to generate go code: %w", err)
		}
		updatedImports, err = updateImports(fileName, gw.String())
		if err != nil {
			return fmt.Errorf("failed to get imports from generated go code: %w", err)
		}
		return nil
	})

	var firstGoNodeInTemplate *ast.File
	// Update the template with the imports.
	// Ensure that there is a Go expression to add the imports to as the first node.
	eg.Go(func() (err error) {
		firstGoNodeInTemplate, err = goparser.ParseFile(fset, fileName, t.Package.Expression.Value+"\n"+importsNode.Expression.Value, goparser.AllErrors|goparser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse imports section: %w", err)
		}
		return nil
	})

	// Wait for completion of both parts.
	if err := eg.Wait(); err != nil {
		return t, err
	}
	// Delete unused imports.
	for _, imp := range firstGoNodeInTemplate.Imports {
		if !containsImport(updatedImports, imp) {
			name, path, err := getImportDetails(imp)
			if err != nil {
				return t, err
			}
			astutil.DeleteNamedImport(fset, firstGoNodeInTemplate, name, path)
		}
	}
	// Add imports, if there are any to add.
	for _, imp := range updatedImports {
		if !containsImport(firstGoNodeInTemplate.Imports, imp) {
			name, path, err := getImportDetails(imp)
			if err != nil {
				return t, err
			}
			astutil.AddNamedImport(fset, firstGoNodeInTemplate, name, path)
		}
	}
	// Edge case: reinsert the import to use import syntax without parentheses.
	if len(firstGoNodeInTemplate.Imports) == 1 {
		name, path, err := getImportDetails(firstGoNodeInTemplate.Imports[0])
		if err != nil {
			return t, err
		}
		astutil.DeleteNamedImport(fset, firstGoNodeInTemplate, name, path)
		astutil.AddNamedImport(fset, firstGoNodeInTemplate, name, path)
	}
	// Write out the Go code with the imports.
	updatedGoCode := new(strings.Builder)
	err := format.Node(updatedGoCode, fset, firstGoNodeInTemplate)
	if err != nil {
		return t, fmt.Errorf("failed to write updated go code: %w", err)
	}
	// Remove the package statement from the node, by cutting the first line of the file.
	importsNode.Expression.Value = strings.TrimSpace(strings.SplitN(updatedGoCode.String(), "\n", 2)[1])
	if len(updatedImports) == 0 && importsNode.Expression.Value == "" {
		t.Nodes = t.Nodes[1:]
		return t, nil
	}
	t.Nodes[0] = importsNode
	return t, nil
}

func getImportDetails(imp *ast.ImportSpec) (name, importPath string, err error) {
	if imp.Name != nil {
		name = imp.Name.Name
	}
	if imp.Path != nil {
		importPath, err = strconv.Unquote(imp.Path.Value)
		if err != nil {
			err = fmt.Errorf("failed to unquote package path %s: %w", imp.Path.Value, err)
			return
		}
	}
	return name, importPath, nil
}

func containsImport(imports []*ast.ImportSpec, spec *ast.ImportSpec) bool {
	for _, imp := range imports {
		if imp.Path.Value == spec.Path.Value {
			return true
		}
	}

	return false
}
