package templ_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func TestJoin(t *testing.T) {
	hello := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "Hello"); err != nil {
			t.Fatalf("failed to write string: %v", err)
		}
		return nil
	})
	world := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "World"); err != nil {
			t.Fatalf("failed to write string: %v", err)
		}
		return nil
	})
	components := []templ.Component{hello, world}
	emptyComponents := []templ.Component{}

	tests := []struct {
		name              string
		input             []templ.Component
		expectedComponent string
	}{
		{
			name:              "render hello world",
			input:             components,
			expectedComponent: "HelloWorld",
		},
		{
			name:              "pass an empty array",
			input:             emptyComponents,
			expectedComponent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := templ.Join(tt.input...)
			b := new(bytes.Buffer)
			err := got.Render(context.Background(), b)
			if err != nil {
				t.Fatalf("failure in rendering %s", err)
			}
			if diff := cmp.Diff(tt.expectedComponent, b.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
