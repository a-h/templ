package testcssmiddleware

import (
	_ "embed"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/htmlformat"
	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/errgroup"
)

//go:embed expected.html
var expected string

var expectedCSS = `.red_050e5e03{color:red;}
`

func Test(t *testing.T) {
	var wg errgroup.Group

	// Format the expected value.
	wg.Go(func() error {
		e := new(strings.Builder)
		err := htmlformat.Fragment(e, strings.NewReader(expected))
		if err != nil {
			return fmt.Errorf("expected html formatting error: %w", err)
		}
		expected = e.String()
		return nil
	})

	component := render("Red text")
	h := templ.Handler(component)
	cssmw := templ.NewCSSMiddleware(h, red())

	// Create the actual value.
	var actual string
	wg.Go(func() error {
		w := httptest.NewRecorder()
		cssmw.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

		a := new(strings.Builder)
		err := htmlformat.Fragment(a, w.Body)
		if err != nil {
			return fmt.Errorf("actual html formatting error: %w", err)
		}
		actual = a.String()
		return nil
	})

	var actualCSS string
	wg.Go(func() error {
		w := httptest.NewRecorder()
		cssmw.ServeHTTP(w, httptest.NewRequest("GET", "/styles/templ.css", nil))

		a := new(strings.Builder)
		err := htmlformat.Fragment(a, w.Body)
		if err != nil {
			return fmt.Errorf("actual html formatting error: %w", err)
		}
		actualCSS = a.String()
		return nil
	})

	if err := wg.Wait(); err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Error(diff)
	}
	if diff := cmp.Diff(expectedCSS, actualCSS); diff != "" {
		t.Error(diff)
	}
}
