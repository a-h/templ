package proxy

import (
	"sync"

	"github.com/a-h/templ/parser/v2"
)

// NewSourceMapCache creates a cache of .templ file URIs to the source map.
func NewSourceMapCache() *SourceMapCache {
	return &SourceMapCache{
		m:              new(sync.Mutex),
		uriToSourceMap: make(map[string]*parser.SourceMap),
	}
}

// SourceMapCache is a cache of .templ file URIs to the source map.
type SourceMapCache struct {
	m              *sync.Mutex
	uriToSourceMap map[string]*parser.SourceMap
}

func (fc *SourceMapCache) Set(uri string, m *parser.SourceMap) {
	fc.m.Lock()
	defer fc.m.Unlock()
	fc.uriToSourceMap[uri] = m
}

func (fc *SourceMapCache) Get(uri string) (m *parser.SourceMap, ok bool) {
	fc.m.Lock()
	defer fc.m.Unlock()
	m, ok = fc.uriToSourceMap[uri]
	return
}

func (fc *SourceMapCache) Delete(uri string) {
	fc.m.Lock()
	defer fc.m.Unlock()
	delete(fc.uriToSourceMap, uri)
}

func (fc *SourceMapCache) URIs() (uris []string) {
	fc.m.Lock()
	defer fc.m.Unlock()
	uris = make([]string, len(fc.uriToSourceMap))
	var i int
	for k := range fc.uriToSourceMap {
		uris[i] = k
		i++
	}
	return uris
}
