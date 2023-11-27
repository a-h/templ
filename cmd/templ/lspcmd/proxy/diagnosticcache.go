package proxy

import (
	lsp "github.com/a-h/protocol"
	"sync"
)

func NewDiagnosticCache() *DiagnosticCache {
	return &DiagnosticCache{
		m:     &sync.Mutex{},
		cache: make(map[string]fileDiagnostic),
	}
}

type fileDiagnostic struct {
	templDiagnostics []lsp.Diagnostic
	goplsDiagnostics []lsp.Diagnostic
}

type DiagnosticCache struct {
	m     *sync.Mutex
	cache map[string]fileDiagnostic
}

func (dc *DiagnosticCache) AddTemplDiagnostics(uri string, goDiagnostics []lsp.Diagnostic) []lsp.Diagnostic {
	dc.m.Lock()
	defer dc.m.Unlock()
	diag := dc.cache[uri]
	diag.goplsDiagnostics = goDiagnostics
	dc.cache[uri] = diag
	return append(diag.templDiagnostics, goDiagnostics...)
}

func (dc *DiagnosticCache) AddGoDiagnostics(uri string, templDiagnostics []lsp.Diagnostic) []lsp.Diagnostic {
	dc.m.Lock()
	defer dc.m.Unlock()
	diag := dc.cache[uri]
	diag.templDiagnostics = templDiagnostics
	dc.cache[uri] = diag
	return append(diag.goplsDiagnostics, templDiagnostics...)
}
