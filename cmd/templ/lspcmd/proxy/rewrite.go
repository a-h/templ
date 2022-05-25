package proxy

import (
	"path"
	"strings"

	"github.com/a-h/templ/parser/v2"
	lsp "go.lsp.dev/protocol"
)

func templatePositionToLSPPosition(p parser.Position) lsp.Position {
	return lsp.Position{Line: p.Line - 1, Character: p.Col + 1}
}

func convertTemplToGoURI(templURI lsp.DocumentURI) (isTemplFile bool, goURI lsp.DocumentURI) {
	base, fileName := path.Split(string(templURI))
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}
	return true, lsp.DocumentURI(base + (strings.TrimSuffix(fileName, ".templ") + "_templ.go"))
}
