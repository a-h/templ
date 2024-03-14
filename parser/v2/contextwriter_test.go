package parser

import (
	"bytes"
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestContextWriter(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		writeContext WriteContext
		expected     string
	}{
		{
			name:         "Adrian's example",
			input:        `<a href="google.com" class={ css() }>Click</a>`,
			writeContext: WriteContextHTML,
			expected:     `<a href="google.com" class="       ">Click</a>`,
		},
	}

	for _, tt := range tests {
		tt := tt
		w := new(bytes.Buffer)
		cw := NewContextWriter(w, tt.writeContext)
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, ok, err := element.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !ok {
				t.Fatalf("unexpected failure for input %q", tt.input)
			}

			if err := actual.Write(cw, 0); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expected, w.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
