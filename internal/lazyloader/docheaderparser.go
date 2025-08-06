package lazyloader

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/a-h/templ/lsp/uri"
)

type docHeaderParser interface {
	parse(filename string) docHeader
}

type goDocHeaderParser struct {
	openDocSources map[string]string
	fileParser     fileParser
}

type fileParser interface {
	parseFile(fset *token.FileSet, file string, overlay any, mode parser.Mode) (*ast.File, error)
}

type goFileParser struct{}

func (goFileParser) parseFile(fset *token.FileSet, file string, overlay any, mode parser.Mode) (*ast.File, error) {
	return parser.ParseFile(fset, file, overlay, mode)
}

func (p *goDocHeaderParser) parse(filename string) docHeader {
	var overlay any
	fileURI := string(uri.File(filename))
	if source, ok := p.openDocSources[fileURI]; ok {
		overlay = source
	}

	fset := token.NewFileSet()
	file, err := p.fileParser.parseFile(fset, filename, overlay, parser.ImportsOnly)
	if err != nil {
		return &goDocHeader{}
	}

	header := &goDocHeader{
		pkgName: file.Name.Name,
		imports: make(map[string]struct{}),
	}

	for _, imp := range file.Imports {
		header.imports[imp.Path.Value] = struct{}{}
	}

	return header
}
