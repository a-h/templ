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
		name              string
		input             []templ.Component
		expectedComponent string
		expectedErr       error
	}{
		{
			name:              "render hello world",
			input:             []templ.Component{hello, world},
			expectedComponent: "HelloWorld",
		},
		{
			name:              "pass an empty array",
			input:             []templ.Component{},
			expectedComponent: "",
		},
		{
			name:              "component returns an error",
			input:             []templ.Component{err},
			expectedComponent: "",
      expectedErr: compErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := templ.Join(tt.input...)
			b := new(bytes.Buffer)
			err := got.Render(context.Background(), b)
			if diff := cmp.Diff(tt.expectedComponent, b.String()); diff != "" {
				t.Error(diff)
			}
			if err != tt.expectedErr {
				t.Fatalf("failure in rendering %s", err)
			}
		})
	}
}
