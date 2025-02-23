package htmldiff

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/a-h/htmlformat"
	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func DiffStrings(expected, actual string) (diff string, err error) {
	// Format both strings.
	var wg sync.WaitGroup
	wg.Add(2)

	errs := make([]error, 2)

	// Format expected.
	go func() {
		defer wg.Done()
		e := new(strings.Builder)
		err := htmlformat.Fragment(e, strings.NewReader(expected))
		if err != nil {
			errs[0] = fmt.Errorf("expected html formatting error: %w", err)
		}
		expected = e.String()
	}()

	// Format actual.
	go func() {
		defer wg.Done()
		a := new(strings.Builder)
		err := htmlformat.Fragment(a, strings.NewReader(actual))
		if err != nil {
			errs[1] = fmt.Errorf("actual html formatting error: %w", err)
		}
		actual = a.String()
	}()

	// Wait for processing.
	wg.Wait()

	return cmp.Diff(expected, actual), errors.Join(errs...)
}

func Diff(input templ.Component, expected string) (diff string, err error) {
	_, diff, err = DiffCtx(context.Background(), input, expected)
	return diff, err
}

func DiffCtx(ctx context.Context, input templ.Component, expected string) (formattedInput, diff string, err error) {
	var wg sync.WaitGroup
	wg.Add(2)

	errs := make([]error, 3)

	// Format the expected value.
	go func() {
		defer wg.Done()
		e := new(strings.Builder)
		err := htmlformat.Fragment(e, strings.NewReader(expected))
		if err != nil {
			errs[0] = fmt.Errorf("expected html formatting error: %w", err)
		}
		expected = e.String()
	}()

	// Pipe via the HTML formatter.
	actual := new(strings.Builder)
	r, w := io.Pipe()
	go func() {
		defer wg.Done()
		err := htmlformat.Fragment(actual, r)
		if err != nil {
			errs[1] = fmt.Errorf("actual html formatting error: %w", err)
		}
	}()

	// Render the component.
	err = input.Render(ctx, w)
	if err != nil {
		errs[2] = fmt.Errorf("failed to render component: %w", err)
	}
	w.Close()

	// Wait for processing.
	wg.Wait()

	return actual.String(), cmp.Diff(expected, actual.String()), errors.Join(errs...)
}
