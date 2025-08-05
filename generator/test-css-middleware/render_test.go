package testcssmiddleware

import (
	_ "embed"
	"net/http/httptest"
	"testing"

	"os"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator/htmldiff"
	"github.com/a-h/templ/internal/prettier"
	"golang.org/x/sync/errgroup"
)

//go:embed expected.html
var expected string

var expectedCSS = `.red_050e5e03 { color: red; }
`

func Test(t *testing.T) {
	var wg errgroup.Group

	// Format the expected value.
	wg.Go(func() (err error) {
		expected, err = prettier.Run(expected, "expected.html", prettier.DefaultCommand)
		if err != nil {
			return err
		}
		return nil
	})

	component := render("Red text")
	h := templ.Handler(component)
	cssmw := templ.NewCSSMiddleware(h, red())

	// Create the actual value.
	var actual string
	wg.Go(func() (err error) {
		w := httptest.NewRecorder()
		cssmw.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		actual, err = prettier.Run(w.Body.String(), "actual.html", prettier.DefaultCommand)
		if err != nil {
			return err
		}
		return nil
	})

	var actualCSS string
	wg.Go(func() (err error) {
		w := httptest.NewRecorder()
		cssmw.ServeHTTP(w, httptest.NewRequest("GET", "/styles/templ.css", nil))
		actualCSS, err = prettier.Run(w.Body.String(), "actual.css", prettier.DefaultCommand)
		if err != nil {
			return err
		}
		return nil
	})

	if err := wg.Wait(); err != nil {
		t.Error(err)
	}
	actualHTML, diffHTML, err := htmldiff.DiffStrings(expected, actual)
	if err != nil {
		t.Error(err)
	}
	if diffHTML != "" {
		if err := os.WriteFile("actual.html", []byte(actualHTML), 0644); err != nil {
			t.Errorf("failed to write actual.html: %v", err)
		}
		t.Error(diffHTML)
	}
	actualCSSOut, diffCSS, err := htmldiff.DiffStrings(expectedCSS, actualCSS)
	if err != nil {
		t.Error(err)
	}
	if diffCSS != "" {
		if err := os.WriteFile("actual_css.html", []byte(actualCSSOut), 0644); err != nil {
			t.Errorf("failed to write actual_css.html: %v", err)
		}
		t.Error(diffCSS)
	}
}
