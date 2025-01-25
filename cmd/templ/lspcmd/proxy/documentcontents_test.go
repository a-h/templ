package proxy

import (
	"log/slog"
	"os"
	"testing"

	lsp "github.com/a-h/templ/lsp/protocol"
	"github.com/google/go-cmp/cmp"
)

func TestDocument(t *testing.T) {
	tests := []struct {
		name       string
		start      string
		operations []func(d *Document)
		expected   string
	}{
		{
			name:  "Replace all content if the range is nil",
			start: "0\n1\n2",
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(nil, "replaced")
				},
			},
			expected: "replaced",
		},
		{
			name:  "If the range matches the length of the file, all of it is replaced",
			start: "0\n1\n2",
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      2,
							Character: 1,
						},
					}, "replaced")
				},
			},
			expected: "replaced",
		},
		{
			name:  "Can insert new text",
			start: ``,
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      0,
							Character: 0,
						},
					}, "abc")
				},
			},
			expected: "abc",
		},
		{
			name:  "Can insert new text that ends with a newline",
			start: ``,
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      0,
							Character: 0,
						},
					}, "abc\n")
				},
			},
			expected: `abc
`,
		},
		{
			name: "Can insert a new line at the end of existing text",
			start: `abc
`,
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 3,
						},
						End: lsp.Position{
							Line:      0,
							Character: 3,
						},
					}, "\n")
				},
			},
			expected: `abc

`,
		},
		{
			name:  "Can insert a word at the start of existing text",
			start: `bc`,
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      0,
							Character: 0,
						},
					}, "a")
				},
			},
			expected: `abc`,
		},
		{
			name:  "Can remove whole line",
			start: "0\n1\n2",
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      1,
							Character: 0,
						},
						End: lsp.Position{
							Line:      2,
							Character: 0,
						},
					}, "")
				},
			},
			expected: "0\n2",
		},
		{
			name:  "Can remove line prefix",
			start: "abcdef",
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      0,
							Character: 3,
						},
					}, "")
				},
			},
			expected: "def",
		},
		{
			name:  "Can remove line substring",
			start: "abcdef",
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 2,
						},
						End: lsp.Position{
							Line:      0,
							Character: 3,
						},
					}, "")
				},
			},
			expected: "abdef",
		},
		{
			name:  "Can remove line suffix",
			start: "abcdef",
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 4,
						},
						End: lsp.Position{
							Line:      0,
							Character: 6,
						},
					}, "")
				},
			},
			expected: "abcd",
		},
		{
			name:  "Can remove across lines",
			start: "0\n1\n22",
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      1,
							Character: 0,
						},
						End: lsp.Position{
							Line:      2,
							Character: 1,
						},
					}, "")
				},
			},
			expected: "0\n2",
		},
		{
			name:  "Can remove part of two lines",
			start: "Line one\nLine two\nLine three",
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 4,
						},
						End: lsp.Position{
							Line:      2,
							Character: 4,
						},
					}, "")
				},
			},
			expected: "Line three",
		},
		{
			name:  "Can remove all lines",
			start: "0\n1\n2",
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      2,
							Character: 1,
						},
					}, "")
				},
			},
			expected: "",
		},
		{
			name:  "Can replace line prefix",
			start: "012345",
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      0,
							Character: 3,
						},
					}, "ABCDEFG")
				},
			},
			expected: "ABCDEFG345",
		},
		{
			name:  "Can replace text across line boundaries",
			start: "Line one\nLine two\nLine three",
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 4,
						},
						End: lsp.Position{
							Line:      2,
							Character: 4,
						},
					}, " one test\nNew Line 2\nNew line")
				},
			},
			expected: "Line one test\nNew Line 2\nNew line three",
		},
		{
			name:  "Can add new line to end of single line",
			start: `a`,
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 1,
						},
						End: lsp.Position{
							Line:      0,
							Character: 1,
						},
					}, "\nb")
				},
			},
			expected: "a\nb",
		},
		{
			name:  "Exceeding the col and line count rounds down to the end of the file",
			start: `a`,
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      200,
							Character: 600,
						},
						End: lsp.Position{
							Line:      300,
							Character: 1200,
						},
					}, "\nb")
				},
			},
			expected: "a\nb",
		},
		{
			name:  "Can remove a line and add it back from the end of the previous line (insert)",
			start: "a\nb\nc",
			operations: []func(d *Document){
				func(d *Document) {
					// Delete.
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      1,
							Character: 0,
						},
						End: lsp.Position{
							Line:      2,
							Character: 0,
						},
					}, "")
					// Put it back.
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 1,
						},
						End: lsp.Position{
							Line:      0,
							Character: 1,
						},
					}, "\nb")
				},
			},
			expected: "a\nb\nc",
		},
		{
			name:  "Can remove a line and add it back from the end of the previous line (overwrite)",
			start: "a\nb\nc",
			operations: []func(d *Document){
				func(d *Document) {
					// Delete.
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      1,
							Character: 0,
						},
						End: lsp.Position{
							Line:      2,
							Character: 0,
						},
					}, "")
					// Put it back.
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 1,
						},
						End: lsp.Position{
							Line:      1,
							Character: 0,
						},
					}, "\nb\n")
				},
			},
			expected: "a\nb\nc",
		},
		{
			name: "Add new line with indent to the end of the line",
			// Based on log entry.
			// {"level":"info","ts":"2022-06-04T20:55:15+01:00","caller":"proxy/server.go:391","msg":"client -> server: DidChange","params":{"textDocument":{"uri":"file:///Users/adrian/github.com/a-h/templ/generator/test-call/template.templ","version":2},"contentChanges":[{"range":{"start":{"line":4,"character":21},"end":{"line":4,"character":21}},"text":"\n\t\t"}]}}
			start: `package testcall

templ personTemplate(p person) {
	<div>
		<h1>{ p.name }</h1>
	</div>
}
`,
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      4,
							Character: 21,
						},
						End: lsp.Position{
							Line:      4,
							Character: 21,
						},
					}, "\n\t\t")
				},
			},
			expected: `package testcall

templ personTemplate(p person) {
	<div>
		<h1>{ p.name }</h1>
		
	</div>
}
`,
		},
		{
			name: "Recreate error smaller",
			// Based on log entry.
			// {"level":"info","ts":"2022-06-04T20:55:15+01:00","caller":"proxy/server.go:391","msg":"client -> server: DidChange","params":{"textDocument":{"uri":"file:///Users/adrian/github.com/a-h/templ/generator/test-call/template.templ","version":2},"contentChanges":[{"range":{"start":{"line":4,"character":21},"end":{"line":4,"character":21}},"text":"\n\t\t"}]}}
			start: "line1\n\t\tline2\nline3",
			operations: []func(d *Document){
				func(d *Document) {
					// Remove \t\tline2
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      1,
							Character: 0,
						},
						End: lsp.Position{
							Line:      2,
							Character: 0,
						},
					}, "")
					// Put it back.
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 5,
						},
						End: lsp.Position{
							Line:      1,
							Character: 0,
						},
					},
						"\n\t\tline2\n")
				},
			},
			expected: "line1\n\t\tline2\nline3",
		},
		{
			name: "Recreate error",
			// Based on log entry.
			// {"level":"info","ts":"2022-06-04T20:55:15+01:00","caller":"proxy/server.go:391","msg":"client -> server: DidChange","params":{"textDocument":{"uri":"file:///Users/adrian/github.com/a-h/templ/generator/test-call/template.templ","version":2},"contentChanges":[{"range":{"start":{"line":4,"character":21},"end":{"line":4,"character":21}},"text":"\n\t\t"}]}}
			start: ` <footer data-testid="footerTemplate">
		<div>&copy; { fmt.Sprintf("%d", time.Now().Year()) }</div>
	</footer>
}
`,
			operations: []func(d *Document){
				func(d *Document) {
					// Remove <div>&copy; { fmt.Sprintf("%d", time.Now().Year()) }</div>
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      1,
							Character: 0,
						},
						End: lsp.Position{
							Line:      2,
							Character: 0,
						},
					}, "")
					// Put it back.
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 38,
						},
						End: lsp.Position{
							Line:      1,
							Character: 0,
						},
					},
						"\n\t\t<div>&copy; { fmt.Sprintf(\"%d\", time.Now().Year()) }</div>\n")
				},
			},
			expected: ` <footer data-testid="footerTemplate">
		<div>&copy; { fmt.Sprintf("%d", time.Now().Year()) }</div>
	</footer>
}
`,
		},
		{
			name: "Insert at start of line",
			// Based on log entry.
			// {"level":"info","ts":"2023-03-25T17:17:38Z","caller":"proxy/server.go:393","msg":"client -> server: DidChange","params":{"textDocument":{"uri":"file:///Users/adrian/github.com/a-h/templ/generator/test-call/template.templ","version":5},"contentChanges":[{"range":{"start":{"line":6,"character":0},"end":{"line":6,"character":0}},"text":"a"}]}}
			start: `b`,
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      0,
							Character: 0,
						},
					}, "a")
				},
			},
			expected: `ab`,
		},
		{
			name: "Insert full new line",
			start: `a
c
d`,
			operations: []func(d *Document){
				func(d *Document) {
					d.Apply(&lsp.Range{
						Start: lsp.Position{
							Line:      1,
							Character: 0,
						},
						End: lsp.Position{
							Line:      1,
							Character: 0,
						},
					}, "b\n")
				},
			},
			expected: `a
b
c
d`,
		},
	}

	for _, tt := range tests {
		logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
		t.Run(tt.name, func(t *testing.T) {
			d := NewDocument(logger, tt.start)
			for _, f := range tt.operations {
				f(d)
			}
			actual := d.String()
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Error(diff)
			}
		})
	}
}
