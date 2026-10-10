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

func TestWriteConditionalAttributeElseIf(t *testing.T) {
	w := new(bytes.Buffer)
	g := generator{
		w:         NewRangeWriter(w),
		sourceMap: parser.NewSourceMap(),
	}
	attr := &parser.ConditionalAttribute{
		Expression: parser.Expression{Value: "a"},
		Then: []parser.Attribute{
			&parser.ConstantAttribute{
				Key:   parser.ConstantAttributeKey{Name: "class"},
				Value: "then",
			},
		},
		ElseIfs: []parser.ConditionalElseIfAttribute{
			{
				Expression: parser.Expression{Value: "b"},
				Then: []parser.Attribute{
					&parser.ConstantAttribute{
						Key:   parser.ConstantAttributeKey{Name: "class"},
						Value: "else-if",
					},
				},
			},
		},
		Else: []parser.Attribute{
			&parser.ConstantAttribute{
				Key:   parser.ConstantAttributeKey{Name: "class"},
				Value: "else",
			},
		},
	}

	if err := g.writeConditionalAttribute(0, "div", attr); err != nil {
		t.Fatalf("failed to write conditional attribute: %v", err)
	}

	output := w.String()
	for _, expected := range []string{
		"if a {\n",
		"} else if b {\n",
		"} else {\n",
		`" class=\"then\""`,
		`" class=\"else-if\""`,
		`" class=\"else\""`,
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("expected generated output to contain %q, got:\n%s", expected, output)
		}
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
