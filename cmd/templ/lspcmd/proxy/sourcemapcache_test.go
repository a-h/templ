package proxy

import (
	"testing"

	"github.com/a-h/templ/parser/v2"
	"github.com/google/go-cmp/cmp"
)

func TestSourceMapCache(t *testing.T) {
	t.Run("can list URIs", func(t *testing.T) {
		smc := NewSourceMapCache()
		smc.Set("d", parser.NewSourceMap())
		smc.Set("c", parser.NewSourceMap())

		actual := smc.URIs()
		expected := []string{"c", "d"}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("can delete entries", func(t *testing.T) {
		smc := NewSourceMapCache()
		smc.Set("a", parser.NewSourceMap())
		smc.Set("b", parser.NewSourceMap())

		smc.Delete("a")
		actual := smc.URIs()
		expected := []string{"b"}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("nil source maps", func(t *testing.T) {
		t.Run("are not cached", func(t *testing.T) {
			smc := NewSourceMapCache()

			var sm *parser.SourceMap
			smc.Set("test", sm)

			if _, ok := smc.Get("test"); ok {
				t.Error("expected nil source map to not be cached")
			}
		})
		t.Run("delete existing cache entries", func(t *testing.T) {
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
	})
}
