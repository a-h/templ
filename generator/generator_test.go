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
	if len(g.sourceMap.Items) != 1 {
		t.Fatalf("expected that writing an expression adds to the source map, ubut got %d items", len(g.sourceMap.Items))
	}
	expectedTarget := parser.NewRange(
		// The from value is (16, 1, 0) because the generator prefixes the
		// expression with a "// GoExpression" comment.
		parser.NewPositionFromValues(16, 1, 0),
		parser.NewPositionFromValues(int64(16+len(exp.Expression.Value)), 2, 5),
	)
	actual := g.sourceMap.Items[0].Target
	if diff := cmp.Diff(expectedTarget, actual); diff != "" {
		t.Errorf("unexpected target:\n%v", diff)
	}
}
