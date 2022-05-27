package proxy

import (
	"path"
	"strings"

	lsp "go.lsp.dev/protocol"
)

func convertTemplToGoURI(templURI lsp.DocumentURI) (isTemplFile bool, goURI lsp.DocumentURI) {
	base, fileName := path.Split(string(templURI))
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}
	return true, lsp.DocumentURI(base + (strings.TrimSuffix(fileName, ".templ") + "_templ.go"))
}
