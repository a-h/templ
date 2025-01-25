package proxy

import (
	"path"
	"strings"

	lsp "github.com/a-h/templ/lsp/protocol"
)

func convertTemplToGoURI(templURI lsp.DocumentURI) (isTemplFile bool, goURI lsp.DocumentURI) {
	base, fileName := path.Split(string(templURI))
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}
	return true, lsp.DocumentURI(base + (strings.TrimSuffix(fileName, ".templ") + "_templ.go"))
}

func convertTemplGoToTemplURI(goURI lsp.DocumentURI) (isTemplGoFile bool, templURI lsp.DocumentURI) {
	base, fileName := path.Split(string(goURI))
	if !strings.HasSuffix(fileName, "_templ.go") {
		return
	}
	return true, lsp.DocumentURI(base + (strings.TrimSuffix(fileName, "_templ.go") + ".templ"))
}
