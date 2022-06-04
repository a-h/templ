package proxy

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	lsp "go.lsp.dev/protocol"
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
					d.Overwrite(lsp.Range{}, "replaced")
				},
			},
			expected: "replaced",
		},
		{
			name:  "Replace all content if the range matches",
			start: "0\n1\n2",
			operations: []func(d *Document){
				func(d *Document) {
					d.Overwrite(lsp.Range{
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
					d.Overwrite(lsp.Range{
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
					d.Overwrite(lsp.Range{
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
					d.Overwrite(lsp.Range{
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
					d.Overwrite(lsp.Range{
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
			name: "Add new line with indent",
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
					d.Overwrite(lsp.Range{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDocument(tt.start)
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
