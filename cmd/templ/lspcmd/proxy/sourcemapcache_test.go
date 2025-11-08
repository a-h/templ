package proxy

import (
	"testing"

	"github.com/a-h/templ/parser/v2"
)

func TestSourceMapCacheNilRobustness(t *testing.T) {
	t.Run("if source map is nil it should not be cached", func(t *testing.T) {
		smc := NewSourceMapCache()

		var sm *parser.SourceMap

		smc.Set("test", sm)

		if _, ok := smc.Get("test"); ok {
			t.Error("expected nil source map to not be cached")
		}
	})

	t.Run("if source map is nil it should clear existing cache entry", func(t *testing.T) {
		smc := NewSourceMapCache()

		sm := parser.NewSourceMap()
		smc.Set("test", sm)
		if _, ok := smc.Get("test"); !ok {
			t.Error("expected non-nil source map to be cached")
		}

		sm = nil
		smc.Set("test", sm)
		if existing, ok := smc.Get("test"); ok || existing != nil {
			t.Error("expected nil source map set to clear existing cache entry")
		}
	})
}
