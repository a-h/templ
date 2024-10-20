package templ_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func TestJoin(t *testing.T) {
	compErr := errors.New("component error")

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
	err := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return compErr
	})

	tests := []struct {
		name           string
		input          []templ.Component
		expectedOutput string
		expectedErr    error
	}{
		{
			name:           "a nil slice of components produces no output",
			input:          nil,
			expectedOutput: "",
		},
		{
			name:           "an empty list of components produces no output",
			input:          []templ.Component{},
			expectedOutput: "",
		},
		{
			name:           "components are rendered in order",
			input:          []templ.Component{hello, world},
			expectedOutput: "HelloWorld",
		},
		{
			name:           "components are rendered in order, and errors returned",
			input:          []templ.Component{hello, err},
			expectedOutput: "Hello",
			expectedErr:    compErr,
		},
		{
			name:           "no further components are rendered after an error",
			input:          []templ.Component{err, hello},
			expectedOutput: "",
			expectedErr:    compErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := templ.Join(tt.input...)
			b := new(bytes.Buffer)
			err := got.Render(context.Background(), b)
			if err != tt.expectedErr {
				t.Fatalf("failed to render component: %v", err)
			}
			if diff := cmp.Diff(tt.expectedOutput, b.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
