package urlbuilder

import (
	"testing"

	"github.com/a-h/templ"
)

func BenchmarkURLBuilder(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		New("https", "example.com").
			Path("a").
			Path("b").
			Path("c").
			Query("key1", "value1").
			Query("key2", "value2").
			Query("key with space", "value with slash").
			Query("key/with/slash", "value/with/slash").
			Path("a/b").
			Query("key between paths", "value between paths").
			Path("c d").
			Fragment("fragment").Build()
	}
}

func TestBasicURL(t *testing.T) {
	t.Parallel()

	builder := New("https", "example.com")
	got := builder.Build()
	expected := templ.URL("https://example.com")

	if got != expected {
		t.Fatalf("got %s, want %s", got, expected)
	}
}

func TestURLWithPaths(t *testing.T) {
	t.Parallel()

	c := "c"
	builder := New("https", "example.com").
		Path("a").
		Path("b").
		Path(c).
		Query("key", "value")

	got := builder.Build()
	expected := templ.URL("https://example.com/a/b/c?key=value")

	if got != expected {
		t.Fatalf("got %s, want %s", got, expected)
	}
}

func TestURLWithMultipleQueries(t *testing.T) {
	t.Parallel()

	builder := New("https", "example.com").
		Path("path").
		Query("key1", "value1").
		Query("key2", "value2")

	got := builder.Build()
	expected := templ.URL("https://example.com/path?key1=value1&key2=value2")

	if got != expected {
		t.Fatalf("got %s, want %s", got, expected)
	}
}

func TestURLWithNoPaths(t *testing.T) {
	t.Parallel()

	builder := New("http", "example.org").
		Query("search", "golang")

	got := builder.Build()
	expected := templ.URL("http://example.org?search=golang")

	if got != expected {
		t.Fatalf("got %s, want %s", got, expected)
	}
}

func TestURLEscapingPath(t *testing.T) {
	t.Parallel()

	builder := New("https", "example.com").
		Path("a/b").
		Path("c d")

	got := builder.Build()
	expected := templ.URL("https://example.com/a%2Fb/c%20d")

	if got != expected {
		t.Fatalf("got %s, want %s", got, expected)
	}
}

func TestURLEscapingQuery(t *testing.T) {
	t.Parallel()

	builder := New("https", "example.com").
		Query("key with space", "value/with/slash")

	got := builder.Build()
	expected := templ.URL("https://example.com?key+with+space=value%2Fwith%2Fslash")

	if got != expected {
		t.Fatalf("got %s, want %s", got, expected)
	}
}