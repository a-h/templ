package testtext

import (
	"net/http/httptest"
	"os"
	"testing"

	_ "embed"

	"github.com/a-h/templ/generator/htmldiff"
)

//go:embed expected.html
var expected string

func Test(t *testing.T) {
	w := httptest.NewRecorder()
	if err := BasicTemplate("Luiz Bonfa").Render(t.Context(), w); err != nil {
		t.Fatal(err)
	}
	if w.Flushed {
		t.Error("rendering to an HTTP writer flushed the response")
	}

	actual, diff, err := htmldiff.DiffStrings(expected, w.Body.String())
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
