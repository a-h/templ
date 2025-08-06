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
	invalidExp := &parser.TemplateFileGoExpression{
		Expression: parser.Expression{
			Value: "line1\nline2",
		},
	}
	if err := g.writeGoExpression(invalidExp); err != nil {
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

	withCommentExp := &parser.TemplateFileGoExpression{
		Expression: parser.Expression{
			Value: `package main

// A comment.
templ h1() {
	<h1></h1>
}
			`,
		},
	}
	if err := g.writeGoExpression(withCommentExp); err != nil {
		t.Fatalf("failed to write Go expression: %v", err)
	}
}

func TestGeneratorForLSP(t *testing.T) {
	input := `package main

templ Hello(name string) {
  if nam`
	tf, err := parser.ParseString(input)
	if err == nil {
		t.Fatalf("expected error, because the file is not valid, got nil")
	}

	w := new(bytes.Buffer)
	op, err := Generate(tf, w)
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}
	if op.SourceMap == nil {
		t.Fatal("expected source map for if expression, got nil")
	}
	if len(op.SourceMap.Expressions) != 3 {
		t.Errorf("expected an expression for the package name, template signature (Hello) and for the if (nam), got %#v", op.SourceMap.Expressions)
	}
}

func TestIsExpressionAttributeValueURL(t *testing.T) {
	testCases := []struct {
		elementName    string
		attrName       string
		expectedOutput bool
	}{
		{
			elementName:    "a",
			attrName:       "href",
			expectedOutput: true,
		},
		{
			elementName:    "a",
			attrName:       "class",
			expectedOutput: false,
		},
		{
			elementName:    "div",
			attrName:       "class",
			expectedOutput: false,
		},
		{
			elementName:    "p",
			attrName:       "href",
			expectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		if output := isExpressionAttributeValueURL(testCase.elementName, testCase.attrName); output != testCase.expectedOutput {
			t.Errorf("expected %t got %t", testCase.expectedOutput, output)
		}
	}
}
