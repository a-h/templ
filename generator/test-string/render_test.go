package teststring

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const input = `Strings are HTML escaped. So ampersands (&), greater than (>), and less than symbols (<) are converted.`
const expected = `Strings are HTML escaped. So ampersands (&amp;), greater than (&gt;), and less than symbols (&lt;) are converted.`

func TestRender(t *testing.T) {
	w := new(strings.Builder)
	err := render(context.Background(), w, input)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
