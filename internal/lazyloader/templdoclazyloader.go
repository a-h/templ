package lazyloader

import (
	"context"
	"errors"
	"fmt"

	lsp "github.com/a-h/templ/lsp/protocol"
	"golang.org/x/tools/go/packages"
)

// TemplDocLazyLoader lazily loads templ documents as necessary.
type TemplDocLazyLoader interface {
	// Load loads a templ document and its dependencies.
	Load(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error

	// Sync syncs the dependencies of a templ document using the changes made to the document.
	Sync(ctx context.Context, params *lsp.DidChangeTextDocumentParams) error

	// Unload unloads a templ document and its dependencies.
	Unload(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error

	// HasLoaded reports whether a templ document and its dependencies have been loaded.
	HasLoaded(doc lsp.TextDocumentIdentifier) bool
}

// templDocLazyLoader is a loader that uses the go/packages API to lazily load templ documents in the dependency graph.
type templDocLazyLoader struct {
	loadedPkgs      map[string]*packages.Package
	openDocHeaders  map[string]docHeader
	docsPendingLoad map[string]struct{}
	pkgLoader       pkgLoader
	pkgTraverser    pkgTraverser
	docHeaderParser docHeaderParser
	docHandler      TemplDocHandler
}

// NewParams specifies the parameters necessary to create a new lazy loader.
type NewParams struct {
	TemplDocHandler TemplDocHandler
	OpenDocSources  map[string]string
}

// New creates a new lazy loader using the provided arguments.
func New(params NewParams) TemplDocLazyLoader {
	return &templDocLazyLoader{
		loadedPkgs:      make(map[string]*packages.Package),
		openDocHeaders:  make(map[string]docHeader),
		docsPendingLoad: make(map[string]struct{}),
		pkgLoader: &goPkgLoader{
			openDocSources: params.OpenDocSources,
			loadPackages:   packages.Load,
		},
		pkgTraverser: &goPkgTraverser{
			templDocHandler: params.TemplDocHandler,
			pkgsRefCount:    make(map[string]int),
			fileReader:      templFileReader{},
		},
		docHeaderParser: &goDocHeaderParser{
			openDocSources: params.OpenDocSources,
			fileParser:     goFileParser{},
		},
		docHandler: params.TemplDocHandler,
	}
}

// Load loads all templ documents in the dependency graph topologically (dependencies are loaded before dependents).
func (l *templDocLazyLoader) Load(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error {
	filename := params.TextDocument.URI.Filename()

	pkg, err := l.pkgLoader.load(filename)
	if err != nil {
		if errors.Is(err, errNoPkgsLoaded) {
			l.docsPendingLoad[filename] = struct{}{}
			return l.docHandler.HandleDidOpen(ctx, params)
		}
		return fmt.Errorf("load package for file %q: %w", filename, err)
	}

	if err := l.pkgTraverser.openTopologically(ctx, pkg); err != nil {
		return fmt.Errorf("open topologically %q: %w", pkg.PkgPath, err)
	}

	l.loadedPkgs[pkg.PkgPath] = pkg
	l.openDocHeaders[filename] = l.docHeaderParser.parse(filename)

	return nil
}

// Sync loads templ documents in newly added dependencies and unloads those that are no longer necessary.
func (l *templDocLazyLoader) Sync(ctx context.Context, params *lsp.DidChangeTextDocumentParams) error {
	filename := params.TextDocument.URI.Filename()

	header := l.openDocHeaders[filename]
	l.openDocHeaders[filename] = l.docHeaderParser.parse(filename)
	if l.openDocHeaders[filename].equal(header) {
		return nil
	}

	pkg, err := l.pkgLoader.load(filename)
	if err != nil {
		return fmt.Errorf("load package for file %q: %w", filename, err)
	}
	delete(l.docsPendingLoad, filename)

	if _, ok := l.loadedPkgs[pkg.PkgPath]; !ok {
		if err := l.pkgTraverser.openTopologically(ctx, pkg); err != nil {
			return fmt.Errorf("open topologically %q: %w", pkg.PkgPath, err)
		}
		l.loadedPkgs[pkg.PkgPath] = pkg

		return nil
	}

	for _, imp := range pkg.Imports {
		if err := l.pkgTraverser.openTopologically(ctx, imp); err != nil {
			return fmt.Errorf("open topologically %q: %w", imp.PkgPath, err)
		}
	}

	for _, imp := range l.loadedPkgs[pkg.PkgPath].Imports {
		if err := l.pkgTraverser.closeTopologically(ctx, imp); err != nil {
			return fmt.Errorf("close topologically %q: %w", imp.PkgPath, err)
		}
	}
	l.loadedPkgs[pkg.PkgPath] = pkg

	return nil
}

// Unload unloads all templ documents in the dependency graph topologically (dependents are unloaded before dependencies).
func (l *templDocLazyLoader) Unload(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error {
	filename := params.TextDocument.URI.Filename()

	pkg, err := l.pkgLoader.load(filename)
	if err != nil {
		if errors.Is(err, errNoPkgsLoaded) {
			delete(l.docsPendingLoad, filename)
			return l.docHandler.HandleDidClose(ctx, params)
		}
		return fmt.Errorf("load package for file %q: %w", filename, err)
	}

	if err := l.pkgTraverser.closeTopologically(ctx, pkg); err != nil {
		return fmt.Errorf("close topologically %q: %w", pkg.PkgPath, err)
	}

	delete(l.loadedPkgs, pkg.PkgPath)
	delete(l.openDocHeaders, filename)

	return nil
}

func (l *templDocLazyLoader) HasLoaded(doc lsp.TextDocumentIdentifier) bool {
	if _, ok := l.docsPendingLoad[doc.URI.Filename()]; ok {
		return false
	}

	_, ok := l.openDocHeaders[doc.URI.Filename()]
	return ok
}
