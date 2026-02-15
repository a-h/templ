package testcontext

import (
	"context"
	_ "embed"
	"testing"

	"os"

	"github.com/a-h/templ/generator/htmldiff"
	"github.com/a-h/templ/internal/prettier"
)

//go:embed expected.html
var expected string

func Test(t *testing.T) {
	if !prettier.IsAvailable(prettier.DefaultCommand) {
		t.Skip("prettier is not available, skipping test")
	}

	component := render()

	ctx := context.WithValue(context.Background(), contextKeyName, "test")

	actual, diff, err := htmldiff.DiffCtx(ctx, component, expected)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		if err := os.WriteFile("actual.html", []byte(actual), 0644); err != nil {
			t.Errorf("failed to write actual.html: %v", err)
		}
		t.Error(diff)
	}
}
