package testcssusage

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const expected = `<style type="text/css">.className_f179{background-color:#ffffff;color:#ff0000;}</style><button class="className_f179 other" type="button">A</button><button class="className_f179 other" type="button">B</button>`

func TestHTML(t *testing.T) {
	w := new(strings.Builder)
	err := TwoButtons().Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
