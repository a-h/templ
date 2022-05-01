package parser

import (
	"testing"

	"github.com/a-h/lexical/input"
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
			input := input.NewFromString(tt.input)
			result := newDocTypeParser().Parse(input)
			if result.Error != nil {
				t.Fatalf("parser error: %v", result.Error)
			}
			if !result.Success {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result.Item); diff != "" {
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
			expected: newParseError("unclosed DOCTYPE",
				Position{
					Index: 0,
					Line:  1,
					Col:   0,
				},
				Position{
					Index: 15,
					Line:  1,
					Col:   14,
				}),
		},
		{
			name: "doctype new tag started",
			input: `<!DOCTYPE html
		<div>`,
			expected: newParseError("unclosed DOCTYPE",
				Position{
					Index: 17,
					Line:  2,
					Col:   2,
				},
				Position{
					Index: 17,
					Line:  2,
					Col:   2,
				}),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			result := newDocTypeParser().Parse(input)
			if diff := cmp.Diff(tt.expected, result.Error); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
