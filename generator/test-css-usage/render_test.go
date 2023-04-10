package testcssusage

import (
	"context"
	_ "embed"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yosssi/gohtml"
)

//go:embed expected.html
var expected string

func TestHTML(t *testing.T) {
	w := new(strings.Builder)
	err := ThreeButtons().Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	actual := gohtml.Format(w.String())
	expected = gohtml.Format(expected)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Error(diff)
	}
}
