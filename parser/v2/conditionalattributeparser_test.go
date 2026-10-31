package parser

import (
	"strings"
	"testing"

	"github.com/a-h/parse"
)

func TestConditionalAttributeElseIfErrors(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  string
	}{
		{"empty branch", `if a { id="a" } else if b { }`, "attribute else if: invalid content or no attributes"},
		{"missing opening brace", `if a { id="a" } else if b`, "attribute else if"},
		{"missing closing brace", `if a { id="a" } else if b { id="b"`, "expected '}'"},
		{"unterminated attribute string", `if a { id="a" } else if b { id={ "unterminated } }`, "string expression attribute: missing closing brace"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := conditionalAttribute.Parse(parse.NewInput([]byte(tt.input)))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestConditionalAttributeElseIfWithoutElse(t *testing.T) {
	input := "if a {\n\tid=\"a\"\n} else if b {\n\tid=\"b\"\n}"
	attr, matched, err := conditionalAttribute.Parse(parse.NewInput([]byte(input)))
	if err != nil || !matched {
		t.Fatalf("failed to parse conditional attribute: matched=%v, error=%v", matched, err)
	}
	if len(attr.Else) != 0 {
		t.Fatalf("expected no else branch, got %#v", attr.Else)
	}
	if got := attr.String(); got != input {
		t.Errorf("formatted: got %q, want %q", got, input)
	}

	copy := attr.Copy().(*ConditionalAttribute)
	copy.ElseIfs[0].Expression.Value = "c"
	copy.ElseIfs[0].Then[0].(*ConstantAttribute).Value = "c"
	if got := attr.String(); got != input {
		t.Errorf("changing the copy modified the original: got %q, want %q", got, input)
	}
}
