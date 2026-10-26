package jstokenizer

import (
	"testing"

	"github.com/a-h/parse"
)

func TestCommentReaderReadsComments(t *testing.T) {
	tests := []struct {
		name       string
		before     []string
		input      string
		expected   string
		expectedOK bool
	}{
		{
			name:       "a single line comment ends at a line feed",
			input:      "// a\nb",
			expected:   "// a\n",
			expectedOK: true,
		},
		{
			name:       "a single line comment ends at a line separator",
			input:      "// a b",
			expected:   "// a ",
			expectedOK: true,
		},
		{
			name:       "a single line comment ends at a paragraph separator",
			input:      "// a b",
			expected:   "// a ",
			expectedOK: true,
		},
		{
			name:       "a multi-line comment ends at its closing delimiter",
			input:      "/* a */b",
			expected:   "/* a */",
			expectedOK: true,
		},
		{
			name:       "an HTML open comment comments out the rest of the line",
			before:     []string{"a"},
			input:      "<!-- a\nb",
			expected:   "<!-- a\n",
			expectedOK: true,
		},
		{
			name:       "a hashbang at the start of the source is a comment",
			input:      "#! a\nb",
			expected:   "#! a\n",
			expectedOK: true,
		},
		{
			name:       "a hashbang after whitespace is not a comment",
			before:     []string{" "},
			input:      "#! a\nb",
			expectedOK: false,
		},
		{
			name:       "an HTML close comment at the start of the source is a comment",
			input:      "--> a\nb",
			expected:   "--> a\n",
			expectedOK: true,
		},
		{
			name:       "an HTML close comment after whitespace at the start of a line is a comment",
			before:     []string{"a", "\n", " "},
			input:      "--> a\nb",
			expected:   "--> a\n",
			expectedOK: true,
		},
		{
			name:       "an HTML close comment after a line separator is a comment",
			before:     []string{"a", " "},
			input:      "--> a\nb",
			expected:   "--> a\n",
			expectedOK: true,
		},
		{
			name:       "an HTML close comment after code on the same line is not a comment",
			before:     []string{"a"},
			input:      "--> a\nb",
			expectedOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reader := newCommentReader()
			for _, token := range tt.before {
				reader.Record(token)
			}
			actual, ok, err := reader.Read(parse.NewInput(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok != tt.expectedOK {
				t.Fatalf("expected ok to be %v, got %v", tt.expectedOK, ok)
			}
			if actual != tt.expected {
				t.Errorf("expected comment %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestCommentReaderReadsAnHTMLCloseCommentAfterAMultiLineCommentAtTheStartOfALine(t *testing.T) {
	reader := newCommentReader()
	pi := parse.NewInput("/* a */ --> b")
	if _, ok, err := reader.Read(pi); err != nil || !ok {
		t.Fatalf("expected a multi-line comment, got ok %v, err %v", ok, err)
	}
	actual, ok, err := reader.Read(pi)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || actual != "--> b" {
		t.Errorf("expected comment %q, got %q, ok %v", "--> b", actual, ok)
	}
}
