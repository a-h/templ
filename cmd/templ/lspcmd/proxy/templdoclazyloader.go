package proxy

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"

	lsp "github.com/a-h/templ/lsp/protocol"
	"github.com/a-h/templ/lsp/uri"
	"golang.org/x/tools/go/packages"
)

// templDocLazyLoader is a loader that uses the packages API to lazily load templ documents in the dependency graph.
type templDocLazyLoader struct {
	templDocHooks  templDocHooks
	packageLoader  packageLoader
	fileReader     fileReader
	fileParser     fileParser
	pkgsRefCount   map[string]int
	openDocHeaders map[string]*goDocHeader
}

type templDocHooks struct {
	didOpen  func(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error
	didClose func(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error
}

type packageLoader interface {
	load(file string) (*packages.Package, error)
}

type goPackageLoader struct{}

type fileReader interface {
	read(file string) ([]byte, error)
}

type templFileReader struct{}

type fileParser interface {
	parseFile(fset *token.FileSet, file string, overlay any, mode parser.Mode) (*ast.File, error)
}

type templFileParser struct{}

type goDocHeader struct {
	pkgName   string
	imports   map[string]struct{}
	textRange *lsp.Range
}

func (goPackageLoader) load(file string) (*packages.Package, error) {
	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps,
		},
		"file="+file,
	)

	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected 1 package, got %d packages", len(pkgs))
	}

	return pkgs[0], nil
}

func (templFileReader) read(file string) ([]byte, error) {
	return os.ReadFile(file)
}

func (templFileParser) parseFile(fset *token.FileSet, file string, overlay any, mode parser.Mode) (*ast.File, error) {
	return parser.ParseFile(fset, file, overlay, mode)
}

func (g *goDocHeader) equals(other *goDocHeader) bool {
	if other == nil {
		return false
	}

	if g.pkgName != other.pkgName {
		return false
	}

	if len(g.imports) != len(other.imports) {
		return false
	}

	for imp := range g.imports {
		if _, ok := other.imports[imp]; !ok {
			return false
		}
	}

	return true
}

func newTemplDocLazyLoader(templDocHooks templDocHooks) templDocLazyLoader {
	return templDocLazyLoader{
		templDocHooks:  templDocHooks,
		packageLoader:  goPackageLoader{},
		fileReader:     templFileReader{},
		fileParser:     templFileParser{},
		pkgsRefCount:   make(map[string]int),
		openDocHeaders: make(map[string]*goDocHeader),
	}
}

// load loads all templ documents in the dependency graph topologically (dependencies are loaded before dependents).
func (l *templDocLazyLoader) load(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error {
	filename := params.TextDocument.URI.Filename()

	pkg, err := l.packageLoader.load(filename)
	if err != nil {
		return fmt.Errorf("load package for file %q: %w", filename, err)
	}

	if err := l.openTopologically(ctx, pkg, make(map[string]bool)); err != nil {
		return fmt.Errorf("open topologically %q: %w", pkg.PkgPath, err)
	}

	l.openDocHeaders[filename] = l.parseHeader(filename, nil)

	return nil
}

// sync opens newly added dependencies and closes those that are no longer necessary.
func (l *templDocLazyLoader) sync(_ context.Context, params *lsp.DidChangeTextDocumentParams, overlay any) error {
	filename := params.TextDocument.URI.Filename()
	header := l.openDocHeaders[filename]

	didChangeHeader := false
	for _, change := range params.ContentChanges {
		if change.Range == nil || change.Range.Start.Line < header.textRange.End.Line {
			didChangeHeader = true
			break
		}
	}

	if !didChangeHeader {
		return nil
	}

	l.openDocHeaders[filename] = l.parseHeader(filename, overlay)
	if l.openDocHeaders[filename].equals(header) {
		return nil
	}

	return fmt.Errorf("not implemented")
}

// openTopologically opens templ files in dependency-first order (topological sort).
func (l *templDocLazyLoader) openTopologically(ctx context.Context, pkg *packages.Package, visited map[string]bool) error {
	if visited[pkg.PkgPath] {
		return nil
	}
	visited[pkg.PkgPath] = true

	if l.pkgsRefCount[pkg.PkgPath] > 0 {
		l.pkgsRefCount[pkg.PkgPath]++
		return nil
	}

	for _, imp := range pkg.Imports {
		if err := l.openTopologically(ctx, imp, visited); err != nil {
			return fmt.Errorf("open topologically %q: %w", imp.PkgPath, err)
		}
	}

	for _, otherFile := range pkg.OtherFiles {
		if filepath.Ext(otherFile) != ".templ" {
			continue
		}

		text, err := l.fileReader.read(otherFile)
		if err != nil {
			return fmt.Errorf("read file %q: %w", otherFile, err)
		}

		if err := l.templDocHooks.didOpen(ctx, &lsp.DidOpenTextDocumentParams{
			TextDocument: lsp.TextDocumentItem{
				URI:        uri.File(otherFile),
				Text:       string(text),
				Version:    1,
				LanguageID: "go",
			},
		}); err != nil {
			return fmt.Errorf("did open file %q: %w", otherFile, err)
		}
	}
	l.pkgsRefCount[pkg.PkgPath]++

	return nil
}

// parseHeader parses the header from a templ file using the provided overlay contents when available.
func (l *templDocLazyLoader) parseHeader(filename string, overlay any) *goDocHeader {
	fset := token.NewFileSet()
	file, err := l.fileParser.parseFile(fset, filename, overlay, parser.ImportsOnly)
	if err != nil {
		return &goDocHeader{
			textRange: &lsp.Range{
				End: lsp.Position{
					Line: math.MaxUint32,
				},
			},
		}
	}

	header := &goDocHeader{
		pkgName: file.Name.Name,
		imports: make(map[string]struct{}),
	}

	if len(file.Imports) == 0 {
		header.textRange = &lsp.Range{
			End: lsp.Position{
				Line: uint32(fset.File(file.Pos()).LineCount()) + 1,
			},
		}
		return header
	}

	for _, imp := range file.Imports {
		header.imports[imp.Path.Value] = struct{}{}
	}

	lastImportPos := file.Imports[len(file.Imports)-1].Pos()
	header.textRange = &lsp.Range{
		End: lsp.Position{
			Line: uint32(fset.Position(lastImportPos).Line) + 1,
		},
	}
	return header
}

// unload unloads all templ documents in the dependency graph topologically (dependents are unloaded before dependencies).
func (l *templDocLazyLoader) unload(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error {
	filename := params.TextDocument.URI.Filename()

	pkg, err := l.packageLoader.load(filename)
	if err != nil {
		return fmt.Errorf("load package for file %q: %w", filename, err)
	}

	if err := l.closeTopologically(ctx, pkg, make(map[string]bool)); err != nil {
		return fmt.Errorf("close topologically %q: %w", pkg.PkgPath, err)
	}

	delete(l.openDocHeaders, filename)

	return nil
}

// closeTopologically closes templ files in dependent-first order (reverse topological sort).
func (l *templDocLazyLoader) closeTopologically(ctx context.Context, pkg *packages.Package, visited map[string]bool) error {
	if visited[pkg.PkgPath] {
		return nil
	}
	visited[pkg.PkgPath] = true

	if l.pkgsRefCount[pkg.PkgPath] > 1 {
		l.pkgsRefCount[pkg.PkgPath]--
		return nil
	}

	for _, otherFile := range pkg.OtherFiles {
		if filepath.Ext(otherFile) != ".templ" {
			continue
		}

		if err := l.templDocHooks.didClose(ctx, &lsp.DidCloseTextDocumentParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: uri.File(otherFile),
			},
		}); err != nil {
			return fmt.Errorf("did close file %q: %w", otherFile, err)
		}
	}
	delete(l.pkgsRefCount, pkg.PkgPath)

	for _, imp := range pkg.Imports {
		if err := l.closeTopologically(ctx, imp, visited); err != nil {
			return fmt.Errorf("close topologically %q: %w", imp.PkgPath, err)
		}
	}

	return nil
}
