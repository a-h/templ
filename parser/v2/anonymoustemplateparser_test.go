package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestAnonymousTemplateParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *AnonymousTemplate
	}{
		{
			name: "anonymous template: no parameters",
			input: `templ() {
}`,
			expected: &AnonymousTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 11, Line: 1, Col: 1},
				},
				Expression: Expression{
					Value: "()",
					Range: Range{
						From: Position{Index: 5, Line: 0, Col: 5},
						To:   Position{Index: 7, Line: 0, Col: 7},
					},
				},
				Children: nil,
			},
		},
		{
			name: "anonymous template: single parameter",
			input: `templ(name string) {
}`,
			expected: &AnonymousTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 22, Line: 1, Col: 1},
				},
				Expression: Expression{
					Value: "(name string)",
					Range: Range{
						From: Position{Index: 5, Line: 0, Col: 5},
						To:   Position{Index: 18, Line: 0, Col: 18},
					},
				},
				Children: nil,
			},
		},
		{
			name: "anonymous template: multiple parameters",
			input: `templ(name string, age int) {
}`,
			expected: &AnonymousTemplate{
				Range: Range{
					From: Position{Index: 0, Line: 0, Col: 0},
					To:   Position{Index: 31, Line: 1, Col: 1},
				},
				Expression: Expression{
					Value: "(name string, age int)",
					Range: Range{
						From: Position{Index: 5, Line: 0, Col: 5},
						To:   Position{Index: 27, Line: 0, Col: 27},
					},
				},
				Children: nil,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, matched, err := anonymousTemplate.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !matched {
				t.Fatalf("expected match, got no match")
			}
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Errorf("unexpected difference:\n%s", diff)
			}
		})
	}
}

func TestAnonymousTemplateParserWithChildren(t *testing.T) {
	// Test that children are parsed correctly - we just verify there are children,
	// not the exact structure which is tested by other parser tests
	input := parse.NewInput(`templ() {
	<div>content</div>
}`)
	actual, matched, err := anonymousTemplate.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !matched {
		t.Fatalf("expected match, got no match")
	}
	at := actual.(*AnonymousTemplate)
	if at.Expression.Value != "()" {
		t.Errorf("expected expression '()', got %q", at.Expression.Value)
	}
	if len(at.Children) == 0 {
		t.Errorf("expected children, got none")
	}
}

func TestAnonymousTemplateParserErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "missing opening brace",
			input: "templ() \n<div>content</div>\n}",
		},
		{
			name:  "missing opening brace with content",
			input: "templ(x int) content }",
		},
		{
			name:  "invalid parameter syntax: unclosed paren",
			input: "templ(x int {",
		},
		{
			name:  "invalid parameter syntax: no closing paren",
			input: "templ(x",
		},
		{
			name:  "missing closing brace",
			input: "templ() {\n\t<div>content</div>\n",
		},
		{
			name:  "missing closing brace with nested content",
			input: "templ() {\n\t<div>\n\t\t<span>nested</span>\n\t</div>\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, matched, err := anonymousTemplate.Parse(input)
			if err == nil {
				t.Fatal("expected error for incomplete anonymous template, but got nil")
			}
			if !matched {
				t.Fatal("expected to be detected as an anonymous template, but wasn't")
			}
		})
	}
}

func TestAnonymousTemplateParserWhitespaceVariations(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantParams string
	}{
		{
			name: "no space before brace",
			input: `templ(){
}`,
			wantParams: "()",
		},
		{
			name: "multiple spaces before brace",
			input: `templ()   {
}`,
			wantParams: "()",
		},
		{
			name:       "tab before brace",
			input:      "templ()\t{" + "\n}",
			wantParams: "()",
		},
		{
			name: "params with no space before brace",
			input: `templ(x int){
}`,
			wantParams: "(x int)",
		},
		{
			name: "params with extra internal whitespace",
			input: `templ(  x  int  ,  y  string  ) {
}`,
			wantParams: "(  x  int  ,  y  string  )",
		},
		{
			name:       "brace on same line, no newline before content",
			input:      "templ() {\n}",
			wantParams: "()",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			actual, matched, err := anonymousTemplate.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !matched {
				t.Fatalf("expected match, got no match")
			}
			at := actual.(*AnonymousTemplate)
			if at.Expression.Value != tt.wantParams {
				t.Errorf("expected params %q, got %q", tt.wantParams, at.Expression.Value)
			}
		})
	}
}

func TestAnonymousTemplateParserNoMatch(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "not an anonymous template: named template",
			input: `templ Name() {`,
		},
		{
			name:  "not an anonymous template: text starting with templ",
			input: `template text`,
		},
		{
			name:  "not an anonymous template: just text",
			input: `some regular text`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, matched, err := anonymousTemplate.Parse(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if matched {
				t.Fatalf("expected no match, got match")
			}
		})
	}
}
