package parser

import "testing"

func TestTemplateFileParser(t *testing.T) {
	t.Run("does not require a package expression", func(t *testing.T) {
		input := `templ Hello() {
Hello
}`
		tf, err := ParseString(input)
		if err != nil {
			t.Fatalf("failed to parse template, with error: %v", err)
		}
		if len(tf.Nodes) != 1 {
			t.Errorf("expected 1 node, got %+v", tf.Nodes)
		}
	})
	t.Run("but can accept a package expression, if one is provided", func(t *testing.T) {
		input := `package main

templ Hello() {
	Hello
}`
		tf, err := ParseString(input)
		if err != nil {
			t.Fatalf("failed to parse template, with error: %v", err)
		}
		if len(tf.Nodes) != 2 {
			t.Errorf("expected 2 nodes, got %+v", tf.Nodes)
		}
	})
}

func TestDefaultPackageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard filename",
			input:    "/files/on/disk/header.templ",
			expected: "disk",
		},
		{
			name:     "path that starts with numbers",
			input:    "/files/on/123disk/header.templ",
			expected: "main",
		},
		{
			name:     "path that includes hyphens",
			input:    "/files/on/disk-drive/header.templ",
			expected: "main",
		},
		{
			name:     "relative path",
			input:    "header.templ",
			expected: "main",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getDefaultPackageName(tt.input)
			if actual != tt.expected {
				t.Errorf("expected %q got %q", tt.expected, actual)
			}
		})
	}
}
