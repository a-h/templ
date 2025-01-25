package proxy

import (
	"sync"

	lsp "github.com/a-h/templ/lsp/protocol"
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

func zeroLengthSliceIfNil(diags []lsp.Diagnostic) []lsp.Diagnostic {
	if diags == nil {
		return make([]lsp.Diagnostic, 0)
	}
	return diags
}

func (dc *DiagnosticCache) AddTemplDiagnostics(uri string, goDiagnostics []lsp.Diagnostic) []lsp.Diagnostic {
	goDiagnostics = zeroLengthSliceIfNil(goDiagnostics)
	dc.m.Lock()
	defer dc.m.Unlock()
	diag := dc.cache[uri]
	diag.goplsDiagnostics = goDiagnostics
	diag.templDiagnostics = zeroLengthSliceIfNil(diag.templDiagnostics)
	dc.cache[uri] = diag
	return append(diag.templDiagnostics, goDiagnostics...)
}

func (dc *DiagnosticCache) ClearTemplDiagnostics(uri string) {
	dc.m.Lock()
	defer dc.m.Unlock()
	diag := dc.cache[uri]
	diag.templDiagnostics = make([]lsp.Diagnostic, 0)
	dc.cache[uri] = diag
}

func (dc *DiagnosticCache) AddGoDiagnostics(uri string, templDiagnostics []lsp.Diagnostic) []lsp.Diagnostic {
	templDiagnostics = zeroLengthSliceIfNil(templDiagnostics)
	dc.m.Lock()
	defer dc.m.Unlock()
	diag := dc.cache[uri]
	diag.templDiagnostics = templDiagnostics
	diag.goplsDiagnostics = zeroLengthSliceIfNil(diag.goplsDiagnostics)
	dc.cache[uri] = diag
	return append(diag.goplsDiagnostics, templDiagnostics...)
}
