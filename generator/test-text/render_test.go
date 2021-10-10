package testtext

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const expected = `<div>Name: Luiz Bonfa</div>` +
	`<div>Text ` + "`" + `with backticks` + "`" + `</div>` +
	`<div>Text ` + "`" + `with backtick` + `</div>` +
	`<div>Text ` + "`" + `with backtick alongside variable: ` + `Luiz Bonfa</div>`

func TestHTML(t *testing.T) {
	w := new(strings.Builder)
	err := BasicTemplate("Luiz Bonfa").Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
