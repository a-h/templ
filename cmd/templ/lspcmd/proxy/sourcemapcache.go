package proxy

import (
	"sync"

	"github.com/a-h/templ/parser/v2"
)

// NewSourceMapCache creates a cache of .templ file URIs to the source map.
func NewSourceMapCache() *SourceMapCache {
	return &SourceMapCache{
		m:              new(sync.Mutex),
		uriToSource:    make(map[string]string),
		uriToSourceMap: make(map[string]*parser.SourceMap),
	}
}

// SourceMapCache is a cache of .templ file URIs to the source map.
type SourceMapCache struct {
	m              *sync.Mutex
	uriToSource    map[string]string
	uriToSourceMap map[string]*parser.SourceMap
}

func (fc *SourceMapCache) Set(uri string, m *parser.SourceMap, source string) {
	fc.m.Lock()
	defer fc.m.Unlock()
	fc.uriToSource[uri] = source
	fc.uriToSourceMap[uri] = m
}

func (fc *SourceMapCache) Get(uri string) (m *parser.SourceMap, source string, ok bool) {
	fc.m.Lock()
	defer fc.m.Unlock()
	source, _ = fc.uriToSource[uri]
	m, ok = fc.uriToSourceMap[uri]
	return
}

func (fc *SourceMapCache) Delete(uri string) {
	fc.m.Lock()
	defer fc.m.Unlock()
	delete(fc.uriToSource, uri)
	delete(fc.uriToSourceMap, uri)
}
