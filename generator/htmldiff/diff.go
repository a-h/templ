package htmldiff

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/htmlformat"
	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/errgroup"
)

func DiffStrings(expected, actual string) (diff string, err error) {
	// Format both strings.
	var wg errgroup.Group

	// Format expected.
	wg.Go(func() error {
		e := new(strings.Builder)
		err := htmlformat.Fragment(e, strings.NewReader(expected))
		if err != nil {
			return fmt.Errorf("expected html formatting error: %w", err)
		}
		expected = e.String()
		return nil
	})

	// Format actual.
	wg.Go(func() error {
		a := new(strings.Builder)
		err := htmlformat.Fragment(a, strings.NewReader(actual))
		if err != nil {
			return fmt.Errorf("actual html formatting error: %w", err)
		}
		actual = a.String()
		return nil
	})

	// Wait for processing.
	if err = wg.Wait(); err != nil {
		return "", err
	}

	return cmp.Diff(expected, actual), nil
}

func Diff(input templ.Component, expected string) (diff string, err error) {
	_, diff, err = DiffCtx(context.Background(), input, expected)
	return diff, err
}

func DiffCtx(ctx context.Context, input templ.Component, expected string) (formattedInput, diff string, err error) {
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

	// Pipe via the HTML formatter.
	actual := new(strings.Builder)
	r, w := io.Pipe()
	wg.Go(func() error {
		err := htmlformat.Fragment(actual, r)
		if err != nil {
			return fmt.Errorf("actual html formatting error: %w", err)
		}
		return nil
	})

	// Render the component.
	renderErr := input.Render(ctx, w)
	closeErr := w.Close()

	// Wait for processing.
	processingErr := wg.Wait()

	return actual.String(), cmp.Diff(expected, actual.String()), errors.Join(renderErr, closeErr, processingErr)
}
