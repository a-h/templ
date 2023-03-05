package proxy

import (
	"testing"

	lsp "github.com/a-h/protocol"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

func TestDocument(t *testing.T) {
	tests := []struct {
		name       string
		start      string
		operations []func(d *Document)
		expected   string
	}{
		{
			name:  "Replace all content if the range is empty",
			start: "0\n1\n2",
			operations: []func(d *Document){
				func(d *Document) {
					d.Overwrite(&lsp.Range{}, "replaced")
				},
			},
			expected: "replaced",
		},
		{
			name:  "Replace all content if the range is nil",
			start: "0\n1\n2",
			operations: []func(d *Document){
				func(d *Document) {
					d.Overwrite(nil, "replaced")
				},
			},
			expected: "replaced",
		},
		{
			name:  "Replace all content if the range matches",
			start: "0\n1\n2",
			operations: []func(d *Document){
				func(d *Document) {
					d.Overwrite(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      2,
							Character: 0,
						},
					}, "replaced")
				},
			},
			expected: "replaced",
		},
		{
			name:  "Remove whole line",
			start: "0\n1\n2",
			operations: []func(d *Document){
				func(d *Document) {
					d.Overwrite(&lsp.Range{
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
			name:  "Strip prefix",
			start: "012345",
			operations: []func(d *Document){
				func(d *Document) {
					d.Overwrite(&lsp.Range{
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
			expected: "345",
		},
		{
			name:  "Replace line prefix",
			start: "012345",
			operations: []func(d *Document){
				func(d *Document) {
					d.Overwrite(&lsp.Range{
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
			name:  "Overwrite line",
			start: "Line one\nLine two\nLine three",
			operations: []func(d *Document){
				func(d *Document) {
					d.Overwrite(&lsp.Range{
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
			name:  "Add new line to end of single line",
			start: `a`,
			operations: []func(d *Document){
				func(d *Document) {
					d.Overwrite(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 1,
						},
						End: lsp.Position{
							Line:      1,
							Character: 0,
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
					d.Overwrite(&lsp.Range{
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
			name:  "Can remove a line and add it back from the end of the previous line",
			start: "a\nb\nc",
			operations: []func(d *Document){
				func(d *Document) {
					// Delete.
					d.Overwrite(&lsp.Range{
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
					d.Overwrite(&lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 1,
						},
						End: lsp.Position{
							Line:      1,
							Character: 0,
						},
					}, "\nb")
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
					d.Overwrite(&lsp.Range{
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
					d.Overwrite(&lsp.Range{
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
					d.Overwrite(&lsp.Range{
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
					d.Overwrite(&lsp.Range{
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
					d.Overwrite(&lsp.Range{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := zap.NewExample()
			d := NewDocument(log, tt.start)
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
