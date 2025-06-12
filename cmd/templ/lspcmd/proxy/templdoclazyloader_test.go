package proxy

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"reflect"
	"slices"
	"testing"

	lsp "github.com/a-h/templ/lsp/protocol"
	"github.com/a-h/templ/lsp/uri"
	"golang.org/x/tools/go/packages"
)

type mockPackageLoader struct {
	packages map[string]*packages.Package
	err      map[string]error
}

type mockFileReader struct {
	data map[string][]byte
	err  map[string]error
}

type mockFileParser struct {
	source map[string]string
}

func (m mockPackageLoader) load(_ *packages.Config, file string) (*packages.Package, error) {
	return m.packages[file], m.err[file]
}

func (m mockFileReader) read(file string) ([]byte, error) {
	return m.data[file], m.err[file]
}

func (m mockFileParser) parseFile(fset *token.FileSet, file string, _ any, mode parser.Mode) (*ast.File, error) {
	return parser.ParseFile(fset, file, m.source[file], mode)
}

func TestGoDocHeaderEquals(t *testing.T) {
	tests := []struct {
		name  string
		this  *goDocHeader
		other *goDocHeader
		want  bool
	}{
		{
			name: "nil",
			this: &goDocHeader{},
		},
		{
			name: "different package names",
			this: &goDocHeader{
				pkgName: "a",
			},
			other: &goDocHeader{
				pkgName: "b",
			},
		},
		{
			name: "different imports lengths",
			this: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{
					"b": {},
				},
			},
			other: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{},
			},
		},
		{
			name: "different imports",
			this: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{
					"b": {},
				},
			},
			other: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{
					"c": {},
				},
			},
		},
		{
			name: "same",
			this: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{
					"b": {},
				},
			},
			other: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{
					"b": {},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.equals(tt.other); got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestNewTemplDocLazyLoader(t *testing.T) {
	loader := newTemplDocLazyLoader(newTemplDocLazyLoaderParams{
		templDocHooks: templDocHooks{
			didOpen: func(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error {
				return nil
			},
			didClose: func(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error {
				return nil
			},
		},
		openDocSources: map[string]string{},
	})

	if loader.templDocHooks.didOpen == nil {
		t.Errorf("expected didOpen to be set")
	}
	if loader.templDocHooks.didClose == nil {
		t.Errorf("expected didClose to be set")
	}
	if loader.openDocSources == nil {
		t.Errorf("expected openTemplDocSources to be set")
	}
}

func TestTemplDocLazyLoaderLoad(t *testing.T) {
	var order []string
	tests := []struct {
		name                    string
		loader                  templDocLazyLoader
		params                  *lsp.DidOpenTextDocumentParams
		wantErrStr              string
		wantOpenOrder           []string
		wantLoadedPkgs          map[string]*packages.Package
		wantPkgsRefCount        map[string]int
		wantOpenTemplDocHeaders map[string]*goDocHeader
	}{
		{
			name: "err load package",
			loader: templDocLazyLoader{
				packageLoader: mockPackageLoader{
					err: map[string]error{
						"/a.templ": errors.New("error"),
					},
				},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantErrStr: "load package for file \"/a.templ\": error",
		},
		{
			name: "err read file",
			loader: templDocLazyLoader{
				packageLoader: mockPackageLoader{
					packages: map[string]*packages.Package{
						"/a.templ": {
							PkgPath:    "a",
							OtherFiles: []string{"/a.templ"},
						},
					},
				},
				fileReader: mockFileReader{
					err: map[string]error{
						"/a.templ": errors.New("error"),
					},
				},
				pkgsRefCount: map[string]int{},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantErrStr:       "open topologically \"a\": read file \"/a.templ\": error",
			wantPkgsRefCount: map[string]int{},
		},
		{
			name: "err did open",
			loader: templDocLazyLoader{
				templDocHooks: templDocHooks{
					didOpen: func(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error {
						return errors.New("error")
					},
				},
				packageLoader: mockPackageLoader{
					packages: map[string]*packages.Package{
						"/a.templ": {
							PkgPath:    "a",
							OtherFiles: []string{"/a.templ"},
						},
					},
				},
				fileReader: mockFileReader{
					data: map[string][]byte{
						"/a.templ": []byte("data"),
					},
				},
				pkgsRefCount: map[string]int{},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantErrStr:       "open topologically \"a\": did open file \"/a.templ\": error",
			wantPkgsRefCount: map[string]int{},
		},
		{
			name: "success a->b->c",
			loader: templDocLazyLoader{
				templDocHooks: templDocHooks{
					didOpen: func(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error {
						order = append(order, params.TextDocument.URI.Filename())
						return nil
					},
				},
				packageLoader: mockPackageLoader{
					packages: map[string]*packages.Package{
						"/a.templ": {
							PkgPath:    "a",
							OtherFiles: []string{"/a.templ", "/a_other.templ"},
							Imports: map[string]*packages.Package{
								"b": {
									PkgPath:    "b",
									OtherFiles: []string{"/b.templ"},
									Imports: map[string]*packages.Package{
										"c": {
											PkgPath:    "c",
											OtherFiles: []string{"/c.templ", "/c_other.templ"},
										},
									},
								},
							},
						},
					},
				},
				fileReader: mockFileReader{
					data: map[string][]byte{
						"/a.templ":       []byte("data"),
						"/a_other.templ": []byte("data"),
						"/b.templ":       []byte("data"),
						"/c.templ":       []byte("data"),
						"/c_other.templ": []byte("data"),
					},
				},
				fileParser: mockFileParser{
					source: map[string]string{
						"/a.templ": "}[gibberish\n",
					},
				},
				loadedPkgs:     map[string]*packages.Package{},
				pkgsRefCount:   map[string]int{},
				openDocHeaders: map[string]*goDocHeader{},
				openDocSources: map[string]string{},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantOpenOrder: []string{"/c.templ", "/c_other.templ", "/b.templ", "/a.templ", "/a_other.templ"},
			wantLoadedPkgs: map[string]*packages.Package{
				"a": {
					PkgPath:    "a",
					OtherFiles: []string{"/a.templ", "/a_other.templ"},
					Imports: map[string]*packages.Package{
						"b": {
							PkgPath:    "b",
							OtherFiles: []string{"/b.templ"},
							Imports: map[string]*packages.Package{
								"c": {
									PkgPath:    "c",
									OtherFiles: []string{"/c.templ", "/c_other.templ"},
								},
							},
						},
					},
				},
			},
			wantPkgsRefCount: map[string]int{
				"a": 1,
				"b": 1,
				"c": 1,
			},
			wantOpenTemplDocHeaders: map[string]*goDocHeader{
				"/a.templ": {
					textRange: &lsp.Range{
						End: lsp.Position{
							Line: math.MaxUint32,
						},
					},
				},
			},
		},
		{
			name: "success a->b->c->a",
			loader: templDocLazyLoader{
				templDocHooks: templDocHooks{
					didOpen: func(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error {
						order = append(order, params.TextDocument.URI.Filename())
						return nil
					},
				},
				packageLoader: mockPackageLoader{
					packages: map[string]*packages.Package{
						"/a.templ": {
							PkgPath:    "a",
							OtherFiles: []string{"/a.templ", "/a_other.templ"},
							Imports: map[string]*packages.Package{
								"b": {
									PkgPath:    "b",
									OtherFiles: []string{"/b.templ"},
									Imports: map[string]*packages.Package{
										"c": {
											PkgPath:    "c",
											OtherFiles: []string{"/c.templ", "/c_other.templ"},
											Imports: map[string]*packages.Package{
												"a": {
													PkgPath:    "a",
													OtherFiles: []string{"/a_loop.templ", "/a_other_loop.templ"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				fileReader: mockFileReader{
					data: map[string][]byte{
						"/a.templ":       []byte("data"),
						"/a_other.templ": []byte("data"),
						"/b.templ":       []byte("data"),
						"/c.templ":       []byte("data"),
						"/c_other.templ": []byte("data"),
					},
				},
				fileParser: mockFileParser{
					source: map[string]string{
						"/a.templ": "package a\n",
					},
				},
				loadedPkgs:     map[string]*packages.Package{},
				pkgsRefCount:   map[string]int{},
				openDocHeaders: map[string]*goDocHeader{},
				openDocSources: map[string]string{},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantOpenOrder: []string{"/c.templ", "/c_other.templ", "/b.templ", "/a.templ", "/a_other.templ"},
			wantLoadedPkgs: map[string]*packages.Package{
				"a": {
					PkgPath:    "a",
					OtherFiles: []string{"/a.templ", "/a_other.templ"},
					Imports: map[string]*packages.Package{
						"b": {
							PkgPath:    "b",
							OtherFiles: []string{"/b.templ"},
							Imports: map[string]*packages.Package{
								"c": {
									PkgPath:    "c",
									OtherFiles: []string{"/c.templ", "/c_other.templ"},
									Imports: map[string]*packages.Package{
										"a": {
											PkgPath:    "a",
											OtherFiles: []string{"/a_loop.templ", "/a_other_loop.templ"},
										},
									},
								},
							},
						},
					},
				},
			},
			wantPkgsRefCount: map[string]int{
				"a": 1,
				"b": 1,
				"c": 1,
			},
			wantOpenTemplDocHeaders: map[string]*goDocHeader{
				"/a.templ": {
					pkgName: "a",
					imports: map[string]struct{}{},
					textRange: &lsp.Range{
						End: lsp.Position{
							Line: 2,
						},
					},
				},
			},
		},
		{
			name: "success a->b->c with c open",
			loader: templDocLazyLoader{
				templDocHooks: templDocHooks{
					didOpen: func(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error {
						order = append(order, params.TextDocument.URI.Filename())
						return nil
					},
				},
				packageLoader: mockPackageLoader{
					packages: map[string]*packages.Package{
						"/a.templ": {
							PkgPath:    "a",
							OtherFiles: []string{"/a.templ", "/a_other.templ"},
							Imports: map[string]*packages.Package{
								"b": {
									PkgPath:    "b",
									OtherFiles: []string{"/b.templ"},
									Imports: map[string]*packages.Package{
										"c": {
											PkgPath:    "c",
											OtherFiles: []string{"/c.templ", "/c_other.templ"},
										},
									},
								},
							},
						},
					},
				},
				fileReader: mockFileReader{
					data: map[string][]byte{
						"/a.templ":       []byte("data"),
						"/a_other.templ": []byte("data"),
						"/b.templ":       []byte("data"),
						"/c.templ":       []byte("data"),
						"/c_other.templ": []byte("data"),
					},
				},
				fileParser: mockFileParser{
					source: map[string]string{
						"/a.templ": "package a\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n}\n",
					},
				},
				loadedPkgs: map[string]*packages.Package{
					"c": {
						PkgPath:    "c",
						OtherFiles: []string{"/c.templ", "/c_other.templ"},
					},
				},
				pkgsRefCount: map[string]int{
					"c": 1,
				},
				openDocHeaders: map[string]*goDocHeader{},
				openDocSources: map[string]string{},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantOpenOrder: []string{"/b.templ", "/a.templ", "/a_other.templ"},
			wantLoadedPkgs: map[string]*packages.Package{
				"a": {
					PkgPath:    "a",
					OtherFiles: []string{"/a.templ", "/a_other.templ"},
					Imports: map[string]*packages.Package{
						"b": {
							PkgPath:    "b",
							OtherFiles: []string{"/b.templ"},
							Imports: map[string]*packages.Package{
								"c": {
									PkgPath:    "c",
									OtherFiles: []string{"/c.templ", "/c_other.templ"},
								},
							},
						},
					},
				},
				"c": {
					PkgPath:    "c",
					OtherFiles: []string{"/c.templ", "/c_other.templ"},
				},
			},
			wantPkgsRefCount: map[string]int{
				"a": 1,
				"b": 1,
				"c": 2,
			},
			wantOpenTemplDocHeaders: map[string]*goDocHeader{
				"/a.templ": {
					pkgName: "a",
					imports: map[string]struct{}{
						"\"fmt\"": {},
						"\"os\"":  {},
					},
					textRange: &lsp.Range{
						End: lsp.Position{
							Line: 6,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		order = nil
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loader.load(context.Background(), tt.params)
			if tt.wantErrStr == "" && err != nil {
				t.Errorf("expected nil error, got \"%s\"", err.Error())
			} else if tt.wantErrStr != "" && err == nil {
				t.Errorf("expected error \"%s\", got nil", tt.wantErrStr)
			} else if tt.wantErrStr != "" && err != nil && err.Error() != tt.wantErrStr {
				t.Errorf("expected error \"%s\", got \"%s\"", tt.wantErrStr, err.Error())
			}

			if !slices.Equal(tt.wantOpenOrder, order) {
				t.Errorf("expected order \"%v\", got \"%v\"", tt.wantOpenOrder, order)
			}

			if !reflect.DeepEqual(tt.wantLoadedPkgs, tt.loader.loadedPkgs) {
				t.Errorf("expected loaded packages \"%v\", got \"%v\"", tt.wantLoadedPkgs, tt.loader.loadedPkgs)
			}

			if !reflect.DeepEqual(tt.wantPkgsRefCount, tt.loader.pkgsRefCount) {
				t.Errorf("expected count \"%v\", got \"%v\"", tt.wantPkgsRefCount, tt.loader.pkgsRefCount)
			}

			if !reflect.DeepEqual(tt.wantOpenTemplDocHeaders, tt.loader.openDocHeaders) {
				t.Errorf("expected headers \"%v\", got \"%v\"", tt.wantOpenTemplDocHeaders, tt.loader.openDocHeaders)
			}
		})
	}
}

func TestTemplDocLazyLoaderSync(t *testing.T) {
	tests := []struct {
		name       string
		loader     templDocLazyLoader
		params     *lsp.DidChangeTextDocumentParams
		wantErrStr string
	}{
		{
			name: "no header change",
			loader: templDocLazyLoader{
				openDocHeaders: map[string]*goDocHeader{
					"/a.templ": {
						pkgName: "a",
						imports: map[string]struct{}{},
						textRange: &lsp.Range{
							End: lsp.Position{
								Line: 2,
							},
						},
					},
				},
			},
			params: &lsp.DidChangeTextDocumentParams{
				TextDocument: lsp.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: uri.URI("file:///a.templ"),
					},
				},
				ContentChanges: []lsp.TextDocumentContentChangeEvent{
					{
						Range: &lsp.Range{
							Start: lsp.Position{
								Line: 2,
							},
						},
					},
					{
						Range: &lsp.Range{
							Start: lsp.Position{
								Line: 7,
							},
						},
					},
					{
						Range: &lsp.Range{
							Start: lsp.Position{
								Line: 9,
							},
						},
					},
				},
			},
		},
		{
			name: "no parsed header change",
			loader: templDocLazyLoader{
				openDocHeaders: map[string]*goDocHeader{
					"/a.templ": {
						pkgName: "a",
						imports: map[string]struct{}{
							"\"fmt\"": {},
							"\"os\"":  {},
						},
						textRange: &lsp.Range{
							End: lsp.Position{
								Line: 6,
							},
						},
					},
				},
				fileParser: mockFileParser{
					source: map[string]string{
						"/a.templ": "package a\n\nimport (\n\t\"os\"\n\t\"fmt\"\n)\n\nfunc main() {\n}\n",
					},
				},
				openDocSources: map[string]string{},
			},
			params: &lsp.DidChangeTextDocumentParams{
				TextDocument: lsp.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: uri.URI("file:///a.templ"),
					},
				},
				ContentChanges: []lsp.TextDocumentContentChangeEvent{
					{
						Range: &lsp.Range{
							Start: lsp.Position{
								Line: 7,
							},
						},
					},
					{
						Range: &lsp.Range{
							Start: lsp.Position{
								Line: 2,
							},
						},
					},
					{
						Range: &lsp.Range{
							Start: lsp.Position{
								Line: 4,
							},
						},
					},
				},
			},
		},
		{
			name: "no parsed header change",
			loader: templDocLazyLoader{
				openDocHeaders: map[string]*goDocHeader{
					"/a.templ": {
						pkgName: "a",
						imports: map[string]struct{}{
							"\"fmt\"": {},
							"\"os\"":  {},
						},
						textRange: &lsp.Range{
							End: lsp.Position{
								Line: 6,
							},
						},
					},
				},
				fileParser: mockFileParser{
					source: map[string]string{
						"/a.templ": "package a\n\nimport (\n\t\"os\"\n\t\"fmt\"\n)\n\nfunc main() {\n}\n",
					},
				},
				openDocSources: map[string]string{},
			},
			params: &lsp.DidChangeTextDocumentParams{
				TextDocument: lsp.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: uri.URI("file:///a.templ"),
					},
				},
				ContentChanges: []lsp.TextDocumentContentChangeEvent{
					{
						Range: &lsp.Range{
							Start: lsp.Position{
								Line: 7,
							},
						},
					},
					{
						Range: &lsp.Range{
							Start: lsp.Position{
								Line: 2,
							},
						},
					},
					{
						Range: &lsp.Range{
							Start: lsp.Position{
								Line: 4,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loader.sync(context.Background(), tt.params)
			if tt.wantErrStr == "" && err != nil {
				t.Errorf("expected nil error, got \"%s\"", err.Error())
			} else if tt.wantErrStr != "" && err == nil {
				t.Errorf("expected error \"%s\", got nil", tt.wantErrStr)
			} else if tt.wantErrStr != "" && err != nil && err.Error() != tt.wantErrStr {
				t.Errorf("expected error \"%s\", got \"%s\"", tt.wantErrStr, err.Error())
			}
		})
	}
}

func TestTemplDocLazyLoaderUnload(t *testing.T) {
	var order []string
	tests := []struct {
		name             string
		loader           templDocLazyLoader
		params           *lsp.DidCloseTextDocumentParams
		wantErrStr       string
		wantCloseOrder   []string
		wantLoadedPkgs   map[string]*packages.Package
		wantPkgsRefCount map[string]int
	}{
		{
			name: "err load package",
			loader: templDocLazyLoader{
				packageLoader: mockPackageLoader{
					err: map[string]error{
						"/a.templ": errors.New("error"),
					},
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantErrStr: "load package for file \"/a.templ\": error",
		},
		{
			name: "err did close",
			loader: templDocLazyLoader{
				templDocHooks: templDocHooks{
					didClose: func(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error {
						return errors.New("error")
					},
				},
				packageLoader: mockPackageLoader{
					packages: map[string]*packages.Package{
						"/a.templ": {
							PkgPath:    "a",
							OtherFiles: []string{"/a.templ"},
						},
					},
				},
				fileReader: mockFileReader{
					data: map[string][]byte{
						"/a.templ": []byte("data"),
					},
				},
				pkgsRefCount: map[string]int{
					"a": 1,
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantErrStr: "close topologically \"a\": did close file \"/a.templ\": error",
			wantPkgsRefCount: map[string]int{
				"a": 1,
			},
		},
		{
			name: "success a->b->c",
			loader: templDocLazyLoader{
				templDocHooks: templDocHooks{
					didClose: func(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error {
						order = append(order, params.TextDocument.URI.Filename())
						return nil
					},
				},
				packageLoader: mockPackageLoader{
					packages: map[string]*packages.Package{
						"/a.templ": {
							PkgPath:    "a",
							OtherFiles: []string{"/a.templ", "/a_other.templ"},
							Imports: map[string]*packages.Package{
								"b": {
									PkgPath:    "b",
									OtherFiles: []string{"/b.templ"},
									Imports: map[string]*packages.Package{
										"c": {
											PkgPath:    "c",
											OtherFiles: []string{"/c.templ", "/c_other.templ"},
										},
									},
								},
							},
						},
					},
				},
				fileReader: mockFileReader{
					data: map[string][]byte{
						"/a.templ":       []byte("data"),
						"/a_other.templ": []byte("data"),
						"/b.templ":       []byte("data"),
						"/c.templ":       []byte("data"),
						"/c_other.templ": []byte("data"),
					},
				},
				loadedPkgs: map[string]*packages.Package{
					"a": {
						PkgPath:    "a",
						OtherFiles: []string{"/a.templ", "/a_other.templ"},
						Imports: map[string]*packages.Package{
							"b": {
								PkgPath:    "b",
								OtherFiles: []string{"/b.templ"},
								Imports: map[string]*packages.Package{
									"c": {
										PkgPath:    "c",
										OtherFiles: []string{"/c.templ", "/c_other.templ"},
									},
								},
							},
						},
					},
				},
				pkgsRefCount: map[string]int{
					"a": 1,
					"b": 1,
					"c": 1,
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantCloseOrder:   []string{"/a.templ", "/a_other.templ", "/b.templ", "/c.templ", "/c_other.templ"},
			wantLoadedPkgs:   map[string]*packages.Package{},
			wantPkgsRefCount: map[string]int{},
		},
		{
			name: "success a->b->c->a",
			loader: templDocLazyLoader{
				templDocHooks: templDocHooks{
					didClose: func(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error {
						order = append(order, params.TextDocument.URI.Filename())
						return nil
					},
				},
				packageLoader: mockPackageLoader{
					packages: map[string]*packages.Package{
						"/a.templ": {
							PkgPath:    "a",
							OtherFiles: []string{"/a.templ", "/a_other.templ"},
							Imports: map[string]*packages.Package{
								"b": {
									PkgPath:    "b",
									OtherFiles: []string{"/b.templ"},
									Imports: map[string]*packages.Package{
										"c": {
											PkgPath:    "c",
											OtherFiles: []string{"/c.templ", "/c_other.templ"},
											Imports: map[string]*packages.Package{
												"a": {
													PkgPath:    "a",
													OtherFiles: []string{"/a_loop.templ", "/a_other_loop.templ"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				fileReader: mockFileReader{
					data: map[string][]byte{
						"/a.templ":       []byte("data"),
						"/a_other.templ": []byte("data"),
						"/b.templ":       []byte("data"),
						"/c.templ":       []byte("data"),
						"/c_other.templ": []byte("data"),
					},
				},
				loadedPkgs: map[string]*packages.Package{
					"a": {
						PkgPath:    "a",
						OtherFiles: []string{"/a.templ", "/a_other.templ"},
						Imports: map[string]*packages.Package{
							"b": {
								PkgPath:    "b",
								OtherFiles: []string{"/b.templ"},
								Imports: map[string]*packages.Package{
									"c": {
										PkgPath:    "c",
										OtherFiles: []string{"/c.templ", "/c_other.templ"},
										Imports: map[string]*packages.Package{
											"a": {
												PkgPath:    "a",
												OtherFiles: []string{"/a_loop.templ", "/a_other_loop.templ"},
											},
										},
									},
								},
							},
						},
					},
				},
				pkgsRefCount: map[string]int{
					"a": 1,
					"b": 1,
					"c": 1,
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantCloseOrder:   []string{"/a.templ", "/a_other.templ", "/b.templ", "/c.templ", "/c_other.templ"},
			wantLoadedPkgs:   map[string]*packages.Package{},
			wantPkgsRefCount: map[string]int{},
		},
		{
			name: "success a->b->c with b close",
			loader: templDocLazyLoader{
				templDocHooks: templDocHooks{
					didClose: func(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error {
						order = append(order, params.TextDocument.URI.Filename())
						return nil
					},
				},
				packageLoader: mockPackageLoader{
					packages: map[string]*packages.Package{
						"/b.templ": {
							PkgPath:    "b",
							OtherFiles: []string{"/b.templ"},
							Imports: map[string]*packages.Package{
								"c": {
									PkgPath:    "c",
									OtherFiles: []string{"/c.templ", "/c_other.templ"},
								},
							},
						},
					},
				},
				fileReader: mockFileReader{
					data: map[string][]byte{
						"/a.templ":       []byte("data"),
						"/a_other.templ": []byte("data"),
						"/b.templ":       []byte("data"),
						"/c.templ":       []byte("data"),
						"/c_other.templ": []byte("data"),
					},
				},
				loadedPkgs: map[string]*packages.Package{
					"b": {
						PkgPath:    "b",
						OtherFiles: []string{"/b.templ"},
						Imports: map[string]*packages.Package{
							"c": {
								PkgPath:    "c",
								OtherFiles: []string{"/c.templ", "/c_other.templ"},
							},
						},
					},
				},
				pkgsRefCount: map[string]int{
					"a": 1,
					"b": 2,
					"c": 1,
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: uri.URI("file:///b.templ"),
				},
			},
			wantCloseOrder: []string{},
			wantLoadedPkgs: map[string]*packages.Package{},
			wantPkgsRefCount: map[string]int{
				"a": 1,
				"b": 1,
				"c": 1,
			},
		},
	}

	for _, tt := range tests {
		order = nil
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loader.unload(context.Background(), tt.params)
			if tt.wantErrStr == "" && err != nil {
				t.Errorf("expected nil error, got \"%s\"", err.Error())
			} else if tt.wantErrStr != "" && err == nil {
				t.Errorf("expected error \"%s\", got nil", tt.wantErrStr)
			} else if tt.wantErrStr != "" && err != nil && err.Error() != tt.wantErrStr {
				t.Errorf("expected error \"%s\", got \"%s\"", tt.wantErrStr, err.Error())
			}
		})

		if !slices.Equal(tt.wantCloseOrder, order) {
			t.Errorf("expected order \"%v\", got \"%v\"", tt.wantCloseOrder, order)
		}

		if !reflect.DeepEqual(tt.wantLoadedPkgs, tt.loader.loadedPkgs) {
			t.Errorf("expected loaded packages \"%v\", got \"%v\"", tt.wantLoadedPkgs, tt.loader.loadedPkgs)
		}

		if !reflect.DeepEqual(tt.wantPkgsRefCount, tt.loader.pkgsRefCount) {
			t.Errorf("expected count \"%v\", got \"%v\"", tt.wantPkgsRefCount, tt.loader.pkgsRefCount)
		}
	}
}
