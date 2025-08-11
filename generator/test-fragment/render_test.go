package testfragment

import (
	_ "embed"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"os"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator/htmldiff"
)

//go:embed complete.html
var complete string

func Test(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.Handler
		expected string
	}{
		{
			name:     "complete pages can be rendered",
			handler:  templ.Handler(Page()),
			expected: complete,
		},
		{
			name:     "single fragments can be rendered",
			handler:  templ.Handler(Page(), templ.WithFragments("content-a")),
			expected: `<div>Fragment Content A</div>`,
		},
		{
			name:     "multiple fragments can be rendered",
			handler:  templ.Handler(Page(), templ.WithFragments("content-a", "content-b")),
			expected: `<div>Fragment Content A</div><div>Fragment Content B</div>`,
		},
		{
			name:     "outer fragments render their contents, even if inner fragments are not requested",
			handler:  templ.Handler(Page(), templ.WithFragments("outer")),
			expected: `<div>Outer Fragment Start</div><div>Inner Fragment Content</div><div>Outer Fragment End</div>`,
		},
		{
			name:     "inner fragments can be rendered without the outer fragment",
			handler:  templ.Handler(Page(), templ.WithFragments("inner")),
			expected: `<div>Inner Fragment Content</div>`,
		},
		{
			name:     "fragments that don't exist return an empty string",
			handler:  templ.Handler(Page(), templ.WithFragments("non-existent")),
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			tt.handler.ServeHTTP(w, r)

			if actualStatusCode := w.Result().StatusCode; http.StatusOK != actualStatusCode {
				t.Errorf("expected status %d, got %d", http.StatusOK, actualStatusCode)
			}

			body, err := io.ReadAll(w.Result().Body)
			if err != nil {
				t.Errorf("failed to read body: %v", err)
			}

			actual, diff, err := htmldiff.DiffStrings(tt.expected, string(body))
			if err != nil {
				t.Fatalf("failed to diff: %v", err)
			}
			if diff != "" {
				if err := os.WriteFile("actual.html", []byte(actual), 0644); err != nil {
					t.Errorf("failed to write actual.html: %v", err)
				}
				t.Error(diff)
			}
		})
	}
}
