package htmldiff

import (
	"context"
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/a-h/templ/internal/prettier"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/errgroup"
)

func DiffStrings(expected, actual string) (output, diff string, err error) {
	// Format both strings.
	var wg errgroup.Group

	// Format expected.
	wg.Go(func() (err error) {
		expected, err = prettier.Run(expected, "expected.html", prettier.DefaultCommand)
		if err != nil {
			return err
		}
		return nil
	})

	// Format actual.
	wg.Go(func() (err error) {
		actual, err = prettier.Run(actual, "actual.html", prettier.DefaultCommand)
		if err != nil {
			return fmt.Errorf("actual html formatting error: %w", err)
		}
		return nil
	})

	// Wait for processing.
	if err = wg.Wait(); err != nil {
		return "", "", err
	}

	return actual, cmp.Diff(expected, actual), nil
}

func Diff(input templ.Component, expected string) (actual, diff string, err error) {
	return DiffCtx(context.Background(), input, expected)
}

func DiffCtx(ctx context.Context, input templ.Component, expected string) (actual, diff string, err error) {
	var a strings.Builder
	err = input.Render(ctx, &a)
	if err != nil {
		return "", "", fmt.Errorf("failed to render input: %w", err)
	}
	return DiffStrings(expected, a.String())
}
