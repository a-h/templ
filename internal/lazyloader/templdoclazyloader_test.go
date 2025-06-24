package lazyloader

import (
	"context"
	"errors"
	"testing"

	lsp "github.com/a-h/templ/lsp/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestTemplDocLazyLoaderLoad(t *testing.T) {
	tests := []struct {
		name                string
		loader              *templDocLazyLoader
		params              *lsp.DidOpenTextDocumentParams
		wantLoadedPkgs      map[string]*packages.Package
		wantOpenDocHeaders  map[string]docHeader
		wantDocsPendingLoad map[string]struct{}
		wantErrContains     string
	}{
		{
			name: "load package failed",
			loader: &templDocLazyLoader{
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return nil, assert.AnError
					},
				},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: lsp.DocumentURI("file:///foo.go"),
				},
			},
			wantErrContains: "load package for file \"/foo.go\"",
		},
		{
			name: "load package failed with no packages loaded",
			loader: &templDocLazyLoader{
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return nil, errNoPkgsLoaded
					},
				},
				docHandler: &mockTemplDocHandler{
					openedDocs: map[string]int{},
					error:      errors.New("mock error"),
				},
				docsPendingLoad: map[string]struct{}{},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: lsp.DocumentURI("file:///foo.go"),
				},
			},
			wantDocsPendingLoad: map[string]struct{}{
				"/foo.go": {},
			},
			wantErrContains: "mock error",
		},
		{
			name: "open topologically failed",
			loader: &templDocLazyLoader{
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return []*packages.Package{{PkgPath: "foo_pkg"}}, nil
					},
				},
				pkgTraverser: &mockPkgTraverser{
					openErrors: map[string]error{"foo_pkg": assert.AnError},
				},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: lsp.DocumentURI("file:///foo.go"),
				},
			},
			wantErrContains: "open topologically \"foo_pkg\"",
		},
		{
			name: "loaded successfully",
			loader: &templDocLazyLoader{
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return []*packages.Package{{PkgPath: "foo_pkg"}}, nil
					},
				},
				pkgTraverser: &mockPkgTraverser{},
				docHeaderParser: &mockDocHeaderParser{
					headers: map[string]docHeader{
						"/foo.go": &goDocHeader{pkgName: "foo_pkg"},
					},
				},
				loadedPkgs:     map[string]*packages.Package{},
				openDocHeaders: map[string]docHeader{},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: lsp.DocumentURI("file:///foo.go"),
				},
			},
			wantLoadedPkgs: map[string]*packages.Package{
				"foo_pkg": {PkgPath: "foo_pkg"},
			},
			wantOpenDocHeaders: map[string]docHeader{
				"/foo.go": &goDocHeader{pkgName: "foo_pkg"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loader.Load(context.Background(), tt.params)

			if tt.wantErrContains != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrContains)
				assert.Equal(t, tt.wantDocsPendingLoad, tt.loader.docsPendingLoad)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLoadedPkgs, tt.loader.loadedPkgs)
				assert.Equal(t, tt.wantOpenDocHeaders, tt.loader.openDocHeaders)
			}
		})
	}
}

func TestTemplDocLazyLoaderSync(t *testing.T) {
	tests := []struct {
		name                string
		loader              *templDocLazyLoader
		params              *lsp.DidChangeTextDocumentParams
		wantLoadedPkgs      map[string]*packages.Package
		wantOpenDocHeaders  map[string]docHeader
		wantOpenedPkgs      []string
		wantClosedPkgs      []string
		wantDocsPendingLoad map[string]struct{}
		wantErrContains     string
	}{
		{
			name: "same header",
			loader: &templDocLazyLoader{
				openDocHeaders: map[string]docHeader{
					"/foo.go": &goDocHeader{
						pkgName: "foo_pkg",
						imports: map[string]struct{}{
							"fmt": {},
							"os":  {},
						},
					},
				},
				docHeaderParser: &mockDocHeaderParser{
					headers: map[string]docHeader{
						"/foo.go": &goDocHeader{
							pkgName: "foo_pkg",
							imports: map[string]struct{}{
								"fmt": {},
								"os":  {},
							},
						},
					},
				},
				pkgTraverser: &mockPkgTraverser{},
			},
			params: &lsp.DidChangeTextDocumentParams{
				TextDocument: lsp.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: lsp.DocumentURI("file:///foo.go"),
					},
				},
				ContentChanges: []lsp.TextDocumentContentChangeEvent{
					{Range: &lsp.Range{Start: lsp.Position{Line: 11}, End: lsp.Position{Line: 22}}},
					{Range: &lsp.Range{Start: lsp.Position{Line: 3}, End: lsp.Position{Line: 9}}},
				},
			},
			wantOpenDocHeaders: map[string]docHeader{
				"/foo.go": &goDocHeader{
					pkgName: "foo_pkg",
					imports: map[string]struct{}{
						"fmt": {},
						"os":  {},
					},
				},
			},
		},
		{
			name: "load package failed",
			loader: &templDocLazyLoader{
				openDocHeaders: map[string]docHeader{},
				docHeaderParser: &mockDocHeaderParser{
					headers: map[string]docHeader{
						"/foo.go": &goDocHeader{
							pkgName: "foo_pkg",
							imports: map[string]struct{}{
								"fmt": {},
								"os":  {},
							},
						},
					},
				},
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return nil, assert.AnError
					},
				},
				pkgTraverser: &mockPkgTraverser{},
			},
			params: &lsp.DidChangeTextDocumentParams{
				TextDocument: lsp.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: lsp.DocumentURI("file:///foo.go"),
					},
				},
				ContentChanges: []lsp.TextDocumentContentChangeEvent{
					{Range: &lsp.Range{Start: lsp.Position{Line: 11}, End: lsp.Position{Line: 22}}},
				},
			},
			wantOpenDocHeaders: map[string]docHeader{
				"/foo.go": &goDocHeader{
					pkgName: "foo_pkg",
					imports: map[string]struct{}{
						"fmt": {},
						"os":  {},
					},
				},
			},
			wantErrContains: "load package for file \"/foo.go\"",
		},
		{
			name: "open topologically failed when package never loaded",
			loader: &templDocLazyLoader{
				openDocHeaders: map[string]docHeader{},
				docHeaderParser: &mockDocHeaderParser{
					headers: map[string]docHeader{
						"/foo.go": &goDocHeader{
							pkgName: "foo_pkg",
							imports: map[string]struct{}{
								"fmt": {},
								"os":  {},
							},
						},
					},
				},
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return []*packages.Package{
							{
								PkgPath: "foo_pkg",
								Imports: map[string]*packages.Package{
									"bar_pkg": {PkgPath: "bar_pkg"},
								},
							},
						}, nil
					},
				},
				pkgTraverser: &mockPkgTraverser{
					openErrors: map[string]error{"foo_pkg": assert.AnError},
				},
				loadedPkgs: map[string]*packages.Package{},
			},
			params: &lsp.DidChangeTextDocumentParams{
				TextDocument: lsp.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: lsp.DocumentURI("file:///foo.go"),
					},
				},
				ContentChanges: []lsp.TextDocumentContentChangeEvent{
					{Range: &lsp.Range{Start: lsp.Position{Line: 11}, End: lsp.Position{Line: 22}}},
					{Range: &lsp.Range{Start: lsp.Position{Line: 3}, End: lsp.Position{Line: 9}}},
				},
			},
			wantOpenDocHeaders: map[string]docHeader{
				"/foo.go": &goDocHeader{
					pkgName: "foo_pkg",
					imports: map[string]struct{}{
						"fmt": {},
						"os":  {},
					},
				},
			},
			wantErrContains: "open topologically \"foo_pkg\"",
		},
		{
			name: "successfully loaded package when never loaded",
			loader: &templDocLazyLoader{
				openDocHeaders: map[string]docHeader{},
				docHeaderParser: &mockDocHeaderParser{
					headers: map[string]docHeader{
						"/foo.go": &goDocHeader{
							pkgName: "foo_pkg",
							imports: map[string]struct{}{
								"fmt": {},
								"os":  {},
							},
						},
					},
				},
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return []*packages.Package{
							{
								PkgPath: "foo_pkg",
								Imports: map[string]*packages.Package{
									"bar_pkg": {PkgPath: "bar_pkg"},
								},
							},
						}, nil
					},
				},
				pkgTraverser: &mockPkgTraverser{
					openErrors: map[string]error{},
				},
				loadedPkgs: map[string]*packages.Package{},
				docsPendingLoad: map[string]struct{}{
					"/foo.go": {},
				},
			},
			params: &lsp.DidChangeTextDocumentParams{
				TextDocument: lsp.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: lsp.DocumentURI("file:///foo.go"),
					},
				},
				ContentChanges: []lsp.TextDocumentContentChangeEvent{
					{Range: &lsp.Range{Start: lsp.Position{Line: 11}, End: lsp.Position{Line: 22}}},
					{Range: &lsp.Range{Start: lsp.Position{Line: 3}, End: lsp.Position{Line: 9}}},
				},
			},
			wantOpenDocHeaders: map[string]docHeader{
				"/foo.go": &goDocHeader{
					pkgName: "foo_pkg",
					imports: map[string]struct{}{
						"fmt": {},
						"os":  {},
					},
				},
			},
			wantLoadedPkgs: map[string]*packages.Package{
				"foo_pkg": {
					PkgPath: "foo_pkg",
					Imports: map[string]*packages.Package{
						"bar_pkg": {PkgPath: "bar_pkg"},
					},
				},
			},
			wantOpenedPkgs:      []string{"foo_pkg"},
			wantDocsPendingLoad: map[string]struct{}{},
		},
		{
			name: "open topologically failed",
			loader: &templDocLazyLoader{
				openDocHeaders: map[string]docHeader{},
				docHeaderParser: &mockDocHeaderParser{
					headers: map[string]docHeader{
						"/foo.go": &goDocHeader{
							pkgName: "foo_pkg",
							imports: map[string]struct{}{
								"fmt": {},
								"os":  {},
							},
						},
					},
				},
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return []*packages.Package{
							{
								PkgPath: "foo_pkg",
								Imports: map[string]*packages.Package{
									"bar_pkg": {PkgPath: "bar_pkg"},
								},
							},
						}, nil
					},
				},
				pkgTraverser: &mockPkgTraverser{
					openErrors: map[string]error{"bar_pkg": assert.AnError},
				},
				loadedPkgs: map[string]*packages.Package{
					"foo_pkg": {
						PkgPath: "foo_pkg",
						Imports: map[string]*packages.Package{
							"bar_pkg": {PkgPath: "bar_pkg"},
						},
					},
				},
			},
			params: &lsp.DidChangeTextDocumentParams{
				TextDocument: lsp.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: lsp.DocumentURI("file:///foo.go"),
					},
				},
				ContentChanges: []lsp.TextDocumentContentChangeEvent{
					{Range: &lsp.Range{Start: lsp.Position{Line: 11}, End: lsp.Position{Line: 22}}},
					{Range: &lsp.Range{Start: lsp.Position{Line: 3}, End: lsp.Position{Line: 9}}},
				},
			},
			wantOpenDocHeaders: map[string]docHeader{
				"/foo.go": &goDocHeader{
					pkgName: "foo_pkg",
					imports: map[string]struct{}{
						"fmt": {},
						"os":  {},
					},
				},
			},
			wantErrContains: "open topologically \"bar_pkg\"",
		},
		{
			name: "close topologically failed",
			loader: &templDocLazyLoader{
				openDocHeaders: map[string]docHeader{},
				docHeaderParser: &mockDocHeaderParser{
					headers: map[string]docHeader{
						"/foo.go": &goDocHeader{
							pkgName: "foo_pkg",
							imports: map[string]struct{}{
								"fmt": {},
								"os":  {},
							},
						},
					},
				},
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return []*packages.Package{
							{
								PkgPath: "foo_pkg",
								Imports: map[string]*packages.Package{},
							},
						}, nil
					},
				},
				pkgTraverser: &mockPkgTraverser{
					closeErrors: map[string]error{"bar_pkg": assert.AnError},
				},
				loadedPkgs: map[string]*packages.Package{
					"foo_pkg": {
						PkgPath: "foo_pkg",
						Imports: map[string]*packages.Package{
							"bar_pkg": {PkgPath: "bar_pkg"},
						},
					},
				},
			},
			params: &lsp.DidChangeTextDocumentParams{
				TextDocument: lsp.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: lsp.DocumentURI("file:///foo.go"),
					},
				},
				ContentChanges: []lsp.TextDocumentContentChangeEvent{
					{Range: &lsp.Range{Start: lsp.Position{Line: 11}, End: lsp.Position{Line: 22}}},
					{Range: &lsp.Range{Start: lsp.Position{Line: 3}, End: lsp.Position{Line: 9}}},
				},
			},
			wantOpenDocHeaders: map[string]docHeader{
				"/foo.go": &goDocHeader{
					pkgName: "foo_pkg",
					imports: map[string]struct{}{
						"fmt": {},
						"os":  {},
					},
				},
			},
			wantErrContains: "close topologically \"bar_pkg\"",
		},
		{
			name: "synced successfully",
			loader: &templDocLazyLoader{
				openDocHeaders: map[string]docHeader{},
				docHeaderParser: &mockDocHeaderParser{
					headers: map[string]docHeader{
						"/foo.go": &goDocHeader{
							pkgName: "foo_pkg",
							imports: map[string]struct{}{
								"fmt": {},
								"os":  {},
							},
						},
					},
				},
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return []*packages.Package{
							{
								PkgPath: "foo_pkg",
								Imports: map[string]*packages.Package{
									"bar_pkg":    {PkgPath: "bar_pkg"},
									"foobar_pkg": {PkgPath: "foobar_pkg"},
								},
							},
						}, nil
					},
				},
				pkgTraverser: &mockPkgTraverser{},
				loadedPkgs: map[string]*packages.Package{
					"foo_pkg": {
						PkgPath: "foo_pkg",
						Imports: map[string]*packages.Package{
							"bar_pkg":    {PkgPath: "bar_pkg"},
							"barfoo_pkg": {PkgPath: "barfoo_pkg"},
						},
					},
				},
			},
			params: &lsp.DidChangeTextDocumentParams{
				TextDocument: lsp.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: lsp.DocumentURI("file:///foo.go"),
					},
				},
				ContentChanges: []lsp.TextDocumentContentChangeEvent{
					{Range: &lsp.Range{Start: lsp.Position{Line: 11}, End: lsp.Position{Line: 22}}},
					{Range: &lsp.Range{Start: lsp.Position{Line: 3}, End: lsp.Position{Line: 9}}},
				},
			},
			wantLoadedPkgs: map[string]*packages.Package{
				"foo_pkg": {
					PkgPath: "foo_pkg",
					Imports: map[string]*packages.Package{
						"bar_pkg":    {PkgPath: "bar_pkg"},
						"foobar_pkg": {PkgPath: "foobar_pkg"},
					},
				},
			},
			wantOpenDocHeaders: map[string]docHeader{
				"/foo.go": &goDocHeader{
					pkgName: "foo_pkg",
					imports: map[string]struct{}{
						"fmt": {},
						"os":  {},
					},
				},
			},
			wantOpenedPkgs: []string{"bar_pkg", "foobar_pkg"},
			wantClosedPkgs: []string{"bar_pkg", "barfoo_pkg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loader.Sync(context.Background(), tt.params)

			if tt.wantErrContains != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLoadedPkgs, tt.loader.loadedPkgs)
				assert.Equal(t, tt.wantOpenDocHeaders, tt.loader.openDocHeaders)
				assert.Equal(t, tt.wantDocsPendingLoad, tt.loader.docsPendingLoad)

				traverser, ok := tt.loader.pkgTraverser.(*mockPkgTraverser)
				require.True(t, ok)
				assert.ElementsMatch(t, tt.wantOpenedPkgs, traverser.openedPkgs)
				assert.ElementsMatch(t, tt.wantClosedPkgs, traverser.closedPkgs)
			}
		})
	}
}

func TestTemplDocLazyLoaderUnload(t *testing.T) {
	tests := []struct {
		name                string
		loader              *templDocLazyLoader
		params              *lsp.DidCloseTextDocumentParams
		wantLoadedPkgs      map[string]*packages.Package
		wantOpenDocHeaders  map[string]docHeader
		wantDocsPendingLoad map[string]struct{}
		wantErrContains     string
	}{
		{
			name: "load package failed",
			loader: &templDocLazyLoader{
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return nil, assert.AnError
					},
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: lsp.DocumentURI("file:///foo.go"),
				},
			},
			wantErrContains: "load package for file \"/foo.go\"",
		},
		{
			name: "load package failed with no packages loaded",
			loader: &templDocLazyLoader{
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return nil, errNoPkgsLoaded
					},
				},
				docHandler: &mockTemplDocHandler{
					closedDocs: map[string]int{},
					error:      errors.New("mock error"),
				},
				docsPendingLoad: map[string]struct{}{
					"/foo.go": {},
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: lsp.DocumentURI("file:///foo.go"),
				},
			},
			wantDocsPendingLoad: map[string]struct{}{},
			wantErrContains:     "mock error",
		},
		{
			name: "close topologically failed",
			loader: &templDocLazyLoader{
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return []*packages.Package{{PkgPath: "foo_pkg"}}, nil
					},
				},
				pkgTraverser: &mockPkgTraverser{
					closeErrors: map[string]error{"foo_pkg": assert.AnError},
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: lsp.DocumentURI("file:///foo.go"),
				},
			},
			wantErrContains: "close topologically \"foo_pkg\"",
		},
		{
			name: "unloaded successfully",
			loader: &templDocLazyLoader{
				pkgLoader: &goPkgLoader{
					loadPackages: func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
						return []*packages.Package{{PkgPath: "foo_pkg"}}, nil
					},
				},
				pkgTraverser: &mockPkgTraverser{},
				docHeaderParser: &mockDocHeaderParser{
					headers: map[string]docHeader{
						"/foo.go": &goDocHeader{pkgName: "foo_pkg"},
					},
				},
				loadedPkgs: map[string]*packages.Package{
					"foo_pkg": {PkgPath: "foo_pkg"},
				},
				openDocHeaders: map[string]docHeader{
					"/foo.go": &goDocHeader{pkgName: "foo_pkg"},
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: lsp.DocumentURI("file:///foo.go"),
				},
			},
			wantLoadedPkgs:     map[string]*packages.Package{},
			wantOpenDocHeaders: map[string]docHeader{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loader.Unload(context.Background(), tt.params)

			if tt.wantErrContains != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrContains)
				assert.Equal(t, tt.wantDocsPendingLoad, tt.loader.docsPendingLoad)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLoadedPkgs, tt.loader.loadedPkgs)
				assert.Equal(t, tt.wantOpenDocHeaders, tt.loader.openDocHeaders)
			}
		})
	}
}

func TestTemplDocLazyLoaderHasLoaded(t *testing.T) {
	tests := []struct {
		name     string
		loader   *templDocLazyLoader
		doc      lsp.TextDocumentIdentifier
		expected bool
	}{
		{
			name: "doc pending load",
			loader: &templDocLazyLoader{
				docsPendingLoad: map[string]struct{}{
					"/foo.go": {},
				},
			},
			doc: lsp.TextDocumentIdentifier{
				URI: lsp.DocumentURI("file:///foo.go"),
			},
		},
		{
			name: "doc loaded",
			loader: &templDocLazyLoader{
				openDocHeaders: map[string]docHeader{
					"/foo.go": &goDocHeader{pkgName: "foo_pkg"},
				},
			},
			doc: lsp.TextDocumentIdentifier{
				URI: lsp.DocumentURI("file:///foo.go"),
			},
			expected: true,
		},
		{
			name: "doc not loaded",
			loader: &templDocLazyLoader{
				openDocHeaders: map[string]docHeader{},
			},
			doc: lsp.TextDocumentIdentifier{
				URI: lsp.DocumentURI("file:///foo.go"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.loader.HasLoaded(tt.doc))
		})
	}
}

type mockPkgTraverser struct {
	openedPkgs  []string
	closedPkgs  []string
	openErrors  map[string]error
	closeErrors map[string]error
}

func (t *mockPkgTraverser) openTopologically(_ context.Context, pkg *packages.Package) error {
	err, ok := t.openErrors[pkg.PkgPath]
	if !ok {
		t.openedPkgs = append(t.openedPkgs, pkg.PkgPath)
	}
	return err
}

func (t *mockPkgTraverser) closeTopologically(_ context.Context, pkg *packages.Package) error {
	err, ok := t.closeErrors[pkg.PkgPath]
	if !ok {
		t.closedPkgs = append(t.closedPkgs, pkg.PkgPath)
	}
	return err
}

type mockDocHeaderParser struct {
	headers map[string]docHeader
}

func (p *mockDocHeaderParser) parse(filename string) docHeader {
	return p.headers[filename]
}
