package parser

import (
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

func TestDocTypeParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected DocType
	}{
		{
			name:  "HTML 5 doctype - uppercase",
			input: `<!DOCTYPE html>`,
			expected: DocType{
				Value: "html",
			},
		},
		{
			name:  "HTML 5 doctype - lowercase",
			input: `<!doctype html>`,
			expected: DocType{
				Value: "html",
			},
		},
		{
			name:  "HTML 4.01 doctype",
			input: `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">`,
			expected: DocType{
				Value: `HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd"`,
			},
		},
		{
			name:  "XHTML 1.1",
			input: `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">`,
			expected: DocType{
				Value: `html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd"`,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, ok, err := docType.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if !ok {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestDocTypeParserErrors(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:  "doctype unclosed",
			input: `<!DOCTYPE html`,
			expected: parse.Error("unclosed DOCTYPE",
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				}),
		},
		{
			name: "doctype new tag started",
			input: `<!DOCTYPE html
		<div>`,
			expected: parse.Error("unclosed DOCTYPE",
				parse.Position{
					Index: 0,
					Line:  0,
					Col:   0,
				}),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, _, err := docType.Parse(input)
			if diff := cmp.Diff(tt.expected, err); diff != "" {
				t.Error(diff)
			}
		})
	}
}
