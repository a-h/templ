package generator

import (
	"bytes"
	"strings"
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

func TestIsTrailingSpaceNeeded(t *testing.T) {
	inlineText := &parser.Text{Value: "hello", TrailingSpace: parser.SpaceHorizontal}
	newlineText := &parser.Text{Value: "hello", TrailingSpace: parser.SpaceVertical}
	inlineStringExpr := &parser.StringExpression{TrailingSpace: parser.SpaceHorizontal}
	selfClosingTempl := &parser.TemplElementExpression{
		Expression:    parser.Expression{Value: "icon()"},
		TrailingSpace: parser.SpaceHorizontal,
	}
	selfClosingTemplNewline := &parser.TemplElementExpression{
		Expression:    parser.Expression{Value: "icon()"},
		TrailingSpace: parser.SpaceVertical,
	}
	blockTempl := &parser.TemplElementExpression{
		Expression: parser.Expression{Value: "wrapper()"},
		Children:   []parser.Node{inlineText},
	}
	inlineElement := &parser.Element{Name: "span"}
	blockElement := &parser.Element{Name: "div"}
	ifExpr := &parser.IfExpression{}

	tests := []struct {
		name     string
		current  parser.Node
		next     parser.Node
		expected bool
	}{
		{
			name:     "inline text needs space before self-closing templ expression",
			current:  inlineText,
			next:     selfClosingTempl,
			expected: true,
		},
		{
			name:     "self-closing templ expression needs space before text",
			current:  selfClosingTempl,
			next:     inlineText,
			expected: true,
		},
		{
			name:     "newline-separated text does not need space before self-closing templ expression",
			current:  newlineText,
			next:     selfClosingTempl,
			expected: false,
		},
		{
			name:     "newline-trailing self-closing templ expression needs space before text",
			current:  selfClosingTemplNewline,
			next:     inlineText,
			expected: true,
		},
		{
			name:     "inline string expression needs space before self-closing templ expression",
			current:  inlineStringExpr,
			next:     selfClosingTempl,
			expected: true,
		},
		{
			name:     "self-closing templ expression needs space before string expression",
			current:  selfClosingTempl,
			next:     inlineStringExpr,
			expected: true,
		},
		{
			name:     "inline text does not need space before block templ expression",
			current:  inlineText,
			next:     blockTempl,
			expected: false,
		},
		{
			name:     "block templ expression does not need space before text",
			current:  blockTempl,
			next:     inlineText,
			expected: false,
		},
		{
			name:     "self-closing templ expression does not need space before inline element",
			current:  selfClosingTempl,
			next:     inlineElement,
			expected: false,
		},
		{
			name:     "self-closing templ expression does not need space before block element",
			current:  selfClosingTempl,
			next:     blockElement,
			expected: false,
		},
		{
			name:     "text needs space before text",
			current:  inlineText,
			next:     inlineText,
			expected: true,
		},
		{
			name:     "text needs space before inline element",
			current:  inlineText,
			next:     inlineElement,
			expected: true,
		},
		{
			name:     "text does not need space before block element",
			current:  inlineText,
			next:     blockElement,
			expected: false,
		},
		{
			name:     "text needs space before if expression",
			current:  inlineText,
			next:     ifExpr,
			expected: true,
		},
		{
			name:     "adjacent self-closing templ expressions do not need space",
			current:  selfClosingTempl,
			next:     selfClosingTempl,
			expected: false,
		},
		{
			name:     "nil current does not need space",
			current:  nil,
			next:     inlineText,
			expected: false,
		},
		{
			name:     "nil next does not need space",
			current:  inlineText,
			next:     nil,
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTrailingSpaceNeeded(tt.current, tt.next)
			if got != tt.expected {
				t.Errorf("got %t, expected %t", got, tt.expected)
			}
		})
	}
}

// TestGenerateEventHandlerAttribute covers GHSA-94c2-xg24-pwhv, where only
// lowercase on* and hx-on: attributes required a templ.ComponentScript, so
// <div OnClick={ s }> rendered a string as an event handler. A string passed to
// an event handler must fail to compile, which a render test can't show, so
// this test asserts on the type of the generated variable.
func TestGenerateEventHandlerAttribute(t *testing.T) {
	tests := []struct {
		name    string
		element string
	}{
		{name: "onclick requires a templ.ComponentScript value", element: `<div onclick={ s }></div>`},
		{name: "OnClick requires a templ.ComponentScript value", element: `<div OnClick={ s }></div>`},
		{name: "ONCLICK requires a templ.ComponentScript value", element: `<div ONCLICK={ s }></div>`},
		{name: "hx-on requires a templ.ComponentScript value", element: `<div hx-on={ s }></div>`},
		{name: "hx-on:click requires a templ.ComponentScript value", element: `<div hx-on:click={ s }></div>`},
		{name: "HX-ON:click requires a templ.ComponentScript value", element: `<div HX-ON:click={ s }></div>`},
		{name: "hx-on-click requires a templ.ComponentScript value", element: `<div hx-on-click={ s }></div>`},
		{name: "hx-on--before-request requires a templ.ComponentScript value", element: `<div hx-on--before-request={ s }></div>`},
		{name: "data-hx-on:click requires a templ.ComponentScript value", element: `<div data-hx-on:click={ s }></div>`},
		{name: "data-hx-on-click requires a templ.ComponentScript value", element: `<div data-hx-on-click={ s }></div>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf, err := parser.ParseString("package main\n\ntempl Test(s templ.ComponentScript) {\n\t" + tt.element + "\n}\n")
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			w := new(bytes.Buffer)
			if _, err = Generate(tf, w); err != nil {
				t.Fatalf("failed to generate: %v", err)
			}
			if !strings.Contains(w.String(), " templ.ComponentScript = s\n") {
				t.Errorf("expected the value to be assigned to a templ.ComponentScript, got:\n%s", w.String())
			}
		})
	}
}
