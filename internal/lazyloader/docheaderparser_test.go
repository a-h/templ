package lazyloader

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoDocHeaderParserParse(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		parser          goDocHeaderParser
		wantGoDocHeader *goDocHeader
	}{
		{
			name:     "returns fallback header on parse error",
			filename: "/bad.go",
			parser: goDocHeaderParser{
				openDocSources: map[string]string{},
				fileParser: mockFileParser{
					err: assert.AnError,
				},
			},
			wantGoDocHeader: &goDocHeader{},
		},
		{
			name:     "returns header with no imports",
			filename: "/no_imports.go",
			parser: goDocHeaderParser{
				openDocSources: map[string]string{
					"/no_imports.go": "package main\n\nfunc main() {}\n",
				},
				fileParser: mockFileParser{
					source: map[string]string{
						"/no_imports.go": "package main\n\nfunc main() {}\n",
					},
				},
			},
			wantGoDocHeader: &goDocHeader{
				pkgName: "main",
				imports: map[string]struct{}{},
			},
		},
		{
			name:     "returns header with imports",
			filename: "/a.templ",
			parser: goDocHeaderParser{
				openDocSources: map[string]string{
					"/a.templ": "package a\n\nimport (\n\t\"strings\"\n\t\"fmt\"\n)\n\nfunc main() {\n}\n",
				},
				fileParser: mockFileParser{
					source: map[string]string{
						"/a.templ": "package a\n\nimport (\n\t\"strings\"\n\t\"fmt\"\n)\n\nfunc main() {\n}\n",
					},
				},
			},
			wantGoDocHeader: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{
					"\"strings\"": {},
					"\"fmt\"":     {},
				},
			},
		},
		{
			name:     "reads overlay source",
			filename: "/overlay.go",
			parser: goDocHeaderParser{
				openDocSources: map[string]string{
					"file:///overlay.go": "package overlay\nfunc main() {}",
				},
				fileParser: mockFileParser{
					source: map[string]string{
						"/overlay.go": "package overlay\nfunc main() {}",
					},
				},
			},
			wantGoDocHeader: &goDocHeader{
				pkgName: "overlay",
				imports: map[string]struct{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.parser.parse(tt.filename)
			assert.IsType(t, &goDocHeader{}, got)
			h := got.(*goDocHeader)
			assert.Equal(t, tt.wantGoDocHeader.pkgName, h.pkgName)
			assert.Equal(t, tt.wantGoDocHeader.imports, h.imports)
		})
	}
}

type mockFileParser struct {
	source map[string]string
	err    error
}

func (m mockFileParser) parseFile(fset *token.FileSet, file string, _ any, mode parser.Mode) (*ast.File, error) {
	if m.err != nil {
		return nil, m.err
	}
	code, ok := m.source[file]
	if !ok {
		return nil, nil
	}
	return parser.ParseFile(fset, file, code, mode)
}
