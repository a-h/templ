package testtemplelement

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const expected = `<div id="1">children1 <div id="2">children2 <div id="3">children3 <div id="4"></div></div></div></div>`

func TestFor(t *testing.T) {
	w := new(strings.Builder)
	err := template().Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
