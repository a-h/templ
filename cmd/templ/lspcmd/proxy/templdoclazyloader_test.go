package proxy

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	lsp "github.com/a-h/templ/lsp/protocol"
	"github.com/a-h/templ/lsp/uri"
	"golang.org/x/tools/go/packages"
)

type mockPackageLoader struct {
	packages map[string][]*packages.Package
	err      map[string]error
}

type mockFileReader struct {
	data map[string][]byte
	err  map[string]error
}

func (m mockPackageLoader) load(file string) ([]*packages.Package, error) {
	return m.packages[file], m.err[file]
}

func (m mockFileReader) read(file string) ([]byte, error) {
	return m.data[file], m.err[file]
}

func TestNewTemplDocLazyLoader(t *testing.T) {
	hooks := templDocHooks{
		didOpen: func(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error {
			return nil
		},
		didClose: func(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error {
			return nil
		},
	}
	loader := newTemplDocLazyLoader(hooks)

	if loader.templDocHooks.didOpen == nil {
		t.Errorf("expected didOpen to be set")
	}
	if loader.templDocHooks.didClose == nil {
		t.Errorf("expected didClose to be set")
	}
}

func TestLoad(t *testing.T) {
	var order []string
	tests := []struct {
		name              string
		loader            templDocLazyLoader
		params            *lsp.DidOpenTextDocumentParams
		wantErrStr        string
		wantOpenOrder     []string
		wantDocsOpenCount map[string]int
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
			wantErrStr: "load packages for file \"/a.templ\": error",
		},
		{
			name: "err read file",
			loader: templDocLazyLoader{
				packageLoader: mockPackageLoader{
					packages: map[string][]*packages.Package{
						"/a.templ": {
							{
								PkgPath:    "a",
								OtherFiles: []string{"/a.templ"},
							},
						},
					},
				},
				fileReader: mockFileReader{
					err: map[string]error{
						"/a.templ": errors.New("error"),
					},
				},
				docsOpenCount: map[string]int{},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantErrStr:        "open topologically \"a\": read file \"/a.templ\": error",
			wantDocsOpenCount: map[string]int{},
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
					packages: map[string][]*packages.Package{
						"/a.templ": {
							{
								PkgPath:    "a",
								OtherFiles: []string{"/a.templ"},
							},
						},
					},
				},
				fileReader: mockFileReader{
					data: map[string][]byte{
						"/a.templ": []byte("data"),
					},
				},
				docsOpenCount: map[string]int{},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantErrStr:        "open topologically \"a\": did open file \"/a.templ\": error",
			wantDocsOpenCount: map[string]int{},
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
					packages: map[string][]*packages.Package{
						"/a.templ": {
							{
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
				docsOpenCount: map[string]int{},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantOpenOrder: []string{"/c.templ", "/c_other.templ", "/b.templ", "/a.templ", "/a_other.templ"},
			wantDocsOpenCount: map[string]int{
				"/c.templ":       1,
				"/c_other.templ": 1,
				"/b.templ":       1,
				"/a.templ":       1,
				"/a_other.templ": 1,
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
					packages: map[string][]*packages.Package{
						"/a.templ": {
							{
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
				docsOpenCount: map[string]int{},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantOpenOrder: []string{"/c.templ", "/c_other.templ", "/b.templ", "/a.templ", "/a_other.templ"},
			wantDocsOpenCount: map[string]int{
				"/c.templ":       1,
				"/c_other.templ": 1,
				"/b.templ":       1,
				"/a.templ":       1,
				"/a_other.templ": 1,
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
					packages: map[string][]*packages.Package{
						"/a.templ": {
							{
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
				docsOpenCount: map[string]int{
					"/c.templ":       1,
					"/c_other.templ": 1,
				},
			},
			params: &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantOpenOrder: []string{"/b.templ", "/a.templ", "/a_other.templ"},
			wantDocsOpenCount: map[string]int{
				"/c.templ":       2,
				"/c_other.templ": 2,
				"/b.templ":       1,
				"/a.templ":       1,
				"/a_other.templ": 1,
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

			if !reflect.DeepEqual(tt.wantDocsOpenCount, tt.loader.docsOpenCount) {
				t.Errorf("expected count \"%v\", got \"%v\"", tt.wantDocsOpenCount, tt.loader.docsOpenCount)
			}
		})
	}
}

func TestUnload(t *testing.T) {
	var order []string
	tests := []struct {
		name              string
		loader            templDocLazyLoader
		params            *lsp.DidCloseTextDocumentParams
		wantErrStr        string
		wantCloseOrder    []string
		wantDocsOpenCount map[string]int
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
			wantErrStr: "load packages for file \"/a.templ\": error",
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
					packages: map[string][]*packages.Package{
						"/a.templ": {
							{
								PkgPath:    "a",
								OtherFiles: []string{"/a.templ"},
							},
						},
					},
				},
				fileReader: mockFileReader{
					data: map[string][]byte{
						"/a.templ": []byte("data"),
					},
				},
				docsOpenCount: map[string]int{
					"/a.templ": 1,
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantErrStr: "close topologically \"a\": did close file \"/a.templ\": error",
			wantDocsOpenCount: map[string]int{
				"/a.templ": 1,
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
					packages: map[string][]*packages.Package{
						"/a.templ": {
							{
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
				docsOpenCount: map[string]int{
					"/a.templ":       1,
					"/a_other.templ": 1,
					"/b.templ":       1,
					"/c.templ":       1,
					"/c_other.templ": 1,
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantCloseOrder:    []string{"/a.templ", "/a_other.templ", "/b.templ", "/c.templ", "/c_other.templ"},
			wantDocsOpenCount: map[string]int{},
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
					packages: map[string][]*packages.Package{
						"/a.templ": {
							{
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
				docsOpenCount: map[string]int{
					"/a.templ":       1,
					"/a_other.templ": 1,
					"/b.templ":       1,
					"/c.templ":       1,
					"/c_other.templ": 1,
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: uri.URI("file:///a.templ"),
				},
			},
			wantCloseOrder:    []string{"/a.templ", "/a_other.templ", "/b.templ", "/c.templ", "/c_other.templ"},
			wantDocsOpenCount: map[string]int{},
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
					packages: map[string][]*packages.Package{
						"/b.templ": {
							{
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
				fileReader: mockFileReader{
					data: map[string][]byte{
						"/a.templ":       []byte("data"),
						"/a_other.templ": []byte("data"),
						"/b.templ":       []byte("data"),
						"/c.templ":       []byte("data"),
						"/c_other.templ": []byte("data"),
					},
				},
				docsOpenCount: map[string]int{
					"/a.templ":       1,
					"/a_other.templ": 1,
					"/b.templ":       2,
					"/c.templ":       2,
					"/c_other.templ": 2,
				},
			},
			params: &lsp.DidCloseTextDocumentParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: uri.URI("file:///b.templ"),
				},
			},
			wantCloseOrder: []string{},
			wantDocsOpenCount: map[string]int{
				"/a.templ":       1,
				"/a_other.templ": 1,
				"/b.templ":       1,
				"/c.templ":       1,
				"/c_other.templ": 1,
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

		if !reflect.DeepEqual(tt.wantDocsOpenCount, tt.loader.docsOpenCount) {
			t.Errorf("expected count \"%v\", got \"%v\"", tt.wantDocsOpenCount, tt.loader.docsOpenCount)
		}
	}
}
