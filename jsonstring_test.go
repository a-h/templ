package templ_test

import (
	"testing"

	"github.com/a-h/templ"
)

func TestJSONString(t *testing.T) {
	t.Run("renders input data as a JSON string", func(t *testing.T) {
		data := map[string]any{"foo": "bar"}
		actual, err := templ.JSONString(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "{\"foo\":\"bar\"}"
		if actual != expected {
			t.Fatalf("unexpected output: want %q, got %q", expected, actual)
		}
	})
	t.Run("returns an error if the data cannot be marshalled", func(t *testing.T) {
		data := make(chan int)
		_, err := templ.JSONString(data)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
	})
}
