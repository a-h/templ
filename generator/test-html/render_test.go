package testhtml

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const expected = `<div>` +
	`<h1>Luiz Bonfa</h1>` +
	`<div style="font-family: &#39;sans-serif&#39;" id="test" data-contents="something with &#34;quotes&#34; and a &lt;tag&gt;">` +
	`<div>email:<a href="mailto: luiz@example.com">luiz@example.com</a>` +
	`</div>` +
	`</div>` +
	`</div>` +
	`<hr noshade>` +
	`<hr optionA optionB optionC="other">` +
	`<hr noshade>`

func TestHTML(t *testing.T) {
	w := new(strings.Builder)
	err := render(person{
		name:  "Luiz Bonfa",
		email: "luiz@example.com",
	}).Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
