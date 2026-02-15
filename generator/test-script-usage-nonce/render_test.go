package testscriptusage

import (
	"context"
	_ "embed"
	"testing"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator/htmldiff"
	"github.com/a-h/templ/internal/prettier"
)

//go:embed expected.html
var expected string

func Test(t *testing.T) {
	if !prettier.IsAvailable(prettier.DefaultCommand) {
		t.Skip("prettier is not available, skipping test")
	}

	component := ThreeButtons()

	ctx := templ.WithNonce(context.Background(), "nonce1")
	_, diff, err := htmldiff.DiffCtx(ctx, component, expected)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		t.Error(diff)
	}
}
