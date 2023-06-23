package generator

import (
	"bytes"
	"testing"

	"github.com/a-h/templ/parser/v2"
	"github.com/google/go-cmp/cmp"
)

func TestGeneratorSourceMap(t *testing.T) {
	w := new(bytes.Buffer)
	g := generator{
		w:         NewRangeWriter(w),
		sourceMap: parser.NewSourceMap(),
	}
	exp := parser.GoExpression{
		Expression: parser.Expression{
			Value: "line1\nline2",
		},
	}
	err := g.writeGoExpression(exp)
	if err != nil {
		t.Fatalf("failed to write Go expression: %v", err)
	}
	expected := parser.NewPosition(0, 0, 0)

	actual, ok := g.sourceMap.TargetPositionFromSource(0, 0)
	if !ok {
		t.Errorf("failed to get matching target")
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected target:\n%v", diff)
	}
}
