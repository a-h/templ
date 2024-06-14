package imports

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"log"
	"path"
	"slices"
	"strings"

	goparser "go/parser"

	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/imports"

	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
)

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
	return gofile.Imports, nil
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
		if _, _, err := generator.Generate(t, gw); err != nil {
			return fmt.Errorf("failed to generate go code: %w", err)
		}
		updatedImports, err = updateImports(fileName, gw.String())
		if err != nil {
			return fmt.Errorf("failed to get imports from generated go code: %w", err)
		}
		return nil
	})

	var gofile *ast.File
	// Update the template with the imports.
	// Ensure that there is a Go expression to add the imports to as the first node.
	eg.Go(func() (err error) {
		gofile, err = goparser.ParseFile(fset, fileName, t.Package.Expression.Value+"\n"+importsNode.Expression.Value, goparser.AllErrors)
		if err != nil {
			log.Printf("failed to parse go code: %v", importsNode.Expression.Value)
			return fmt.Errorf("failed to parse imports section: %w", err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return t, err
	}
	slices.SortFunc(updatedImports, func(a, b *ast.ImportSpec) int {
		return strings.Compare(a.Path.Value, b.Path.Value)
	})
	gofile.Imports = updatedImports
	newImportDecl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: convertSlice(updatedImports),
	}
	// Delete all the existing imports.
	var indicesToDelete []int
	for i, decl := range gofile.Decls {
		if decl, ok := decl.(*ast.GenDecl); ok && decl.Tok == token.IMPORT {
			indicesToDelete = append(indicesToDelete, i)
		}
	}
	for i := len(indicesToDelete) - 1; i >= 0; i-- {
		gofile.Decls = append(gofile.Decls[:indicesToDelete[i]], gofile.Decls[indicesToDelete[i]+1:]...)
	}
	gofile.Decls = append([]ast.Decl{newImportDecl}, gofile.Decls...)
	// Write out the Go code with the imports.
	updatedGoCode := new(strings.Builder)
	err := printer.Fprint(updatedGoCode, fset, gofile)
	if err != nil {
		return t, fmt.Errorf("failed to write updated go code: %w", err)
	}
	importsNode.Expression.Value = strings.TrimSpace(strings.SplitN(updatedGoCode.String(), "\n", 2)[1])
	t.Nodes[0] = importsNode

	return t, nil
}

func convertSlice(slice []*ast.ImportSpec) []ast.Spec {
	result := make([]ast.Spec, len(slice))
	for i, v := range slice {
		result[i] = ast.Spec(v)
	}
	return result
}
