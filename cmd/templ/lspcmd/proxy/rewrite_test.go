package proxy

import (
	"log/slog"
	"testing"

	"github.com/a-h/templ/parser/v2"
	"github.com/google/go-cmp/cmp"

	lsp "github.com/a-h/templ/lsp/protocol"
)

func TestConvertTemplToGoURI(t *testing.T) {
	tests := []struct {
		name        string
		input       lsp.DocumentURI
		wantIsTempl bool
		wantGoURI   lsp.DocumentURI
	}{
		{name: "templ file", input: "file:///tmp/test.templ", wantIsTempl: true, wantGoURI: "file:///tmp/test_templ.go"},
		{name: "go file", input: "file:///tmp/test.go", wantIsTempl: false, wantGoURI: ""},
		{name: "templ go file", input: "file:///tmp/test_templ.go", wantIsTempl: false, wantGoURI: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isTempl, goURI := convertTemplToGoURI(tt.input)
			if isTempl != tt.wantIsTempl {
				t.Errorf("isTemplFile: got %v, expected %v", isTempl, tt.wantIsTempl)
			}
			if isTempl && goURI != tt.wantGoURI {
				t.Errorf("goURI: got %q, expected %q", goURI, tt.wantGoURI)
			}
		})
	}
}

func TestConvertTemplGoToTemplURI(t *testing.T) {
	tests := []struct {
		name          string
		input         lsp.DocumentURI
		wantIsTemplGo bool
		wantTemplURI  lsp.DocumentURI
	}{
		{name: "templ go file", input: "file:///tmp/test_templ.go", wantIsTemplGo: true, wantTemplURI: "file:///tmp/test.templ"},
		{name: "regular go file", input: "file:///tmp/test.go", wantIsTemplGo: false, wantTemplURI: ""},
		{name: "templ file", input: "file:///tmp/test.templ", wantIsTemplGo: false, wantTemplURI: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isTemplGo, templURI := convertTemplGoToTemplURI(tt.input)
			if isTemplGo != tt.wantIsTemplGo {
				t.Errorf("isTemplGoFile: got %v, expected %v", isTemplGo, tt.wantIsTemplGo)
			}
			if isTemplGo && templURI != tt.wantTemplURI {
				t.Errorf("templURI: got %q, expected %q", templURI, tt.wantTemplURI)
			}
		})
	}
}

func newTestSourceMap() *parser.SourceMap {
	sm := parser.NewSourceMap()
	// Map templ line 5, col 10..20 -> Go line 15, col 20..30.
	// The Value must be 10 characters to create per-character mappings.
	sm.Add(
		parser.Expression{
			Value: "0123456789",
			Range: parser.Range{
				From: parser.Position{Index: 0, Line: 5, Col: 10},
				To:   parser.Position{Index: 10, Line: 5, Col: 20},
			},
		},
		parser.Range{
			From: parser.Position{Index: 0, Line: 15, Col: 20},
			To:   parser.Position{Index: 10, Line: 15, Col: 30},
		},
	)
	return sm
}

func TestConvertLocationResults(t *testing.T) {
	cache := NewSourceMapCache()
	cache.Set("file:///tmp/test.templ", newTestSourceMap())
	log := slog.Default()

	tests := []struct {
		name     string
		input    []lsp.Location
		expected []lsp.Location
	}{
		{
			name: "converts templ go URI to templ URI",
			input: []lsp.Location{
				{
					URI: "file:///tmp/test_templ.go",
					Range: lsp.Range{
						Start: lsp.Position{Line: 15, Character: 20},
						End:   lsp.Position{Line: 15, Character: 30},
					},
				},
			},
			expected: []lsp.Location{
				{
					URI: "file:///tmp/test.templ",
					Range: lsp.Range{
						Start: lsp.Position{Line: 5, Character: 10},
						End:   lsp.Position{Line: 5, Character: 20},
					},
				},
			},
		},
		{
			name: "leaves non-templ go URIs unchanged",
			input: []lsp.Location{
				{
					URI: "file:///tmp/other.go",
					Range: lsp.Range{
						Start: lsp.Position{Line: 1, Character: 2},
						End:   lsp.Position{Line: 3, Character: 4},
					},
				},
			},
			expected: []lsp.Location{
				{
					URI: "file:///tmp/other.go",
					Range: lsp.Range{
						Start: lsp.Position{Line: 1, Character: 2},
						End:   lsp.Position{Line: 3, Character: 4},
					},
				},
			},
		},
		{
			name: "mixed results",
			input: []lsp.Location{
				{URI: "file:///tmp/other.go", Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 2}, End: lsp.Position{Line: 3, Character: 4}}},
				{URI: "file:///tmp/test_templ.go", Range: lsp.Range{Start: lsp.Position{Line: 15, Character: 20}, End: lsp.Position{Line: 15, Character: 30}}},
			},
			expected: []lsp.Location{
				{URI: "file:///tmp/other.go", Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 2}, End: lsp.Position{Line: 3, Character: 4}}},
				{URI: "file:///tmp/test.templ", Range: lsp.Range{Start: lsp.Position{Line: 5, Character: 10}, End: lsp.Position{Line: 5, Character: 20}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convertLocationResults(cache, log, tt.input)
			if diff := cmp.Diff(tt.expected, tt.input); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConvertWorkspaceEdit(t *testing.T) {
	cache := NewSourceMapCache()
	cache.Set("file:///tmp/test.templ", newTestSourceMap())
	log := slog.Default()

	t.Run("nil edit", func(t *testing.T) {
		convertWorkspaceEdit(cache, log, nil)
	})
	t.Run("converts DocumentChanges", func(t *testing.T) {
		edit := &lsp.WorkspaceEdit{
			DocumentChanges: []lsp.TextDocumentEdit{
				{
					TextDocument: lsp.OptionalVersionedTextDocumentIdentifier{
						TextDocumentIdentifier: lsp.TextDocumentIdentifier{URI: "file:///tmp/test_templ.go"},
					},
					Edits: []lsp.TextEdit{
						{
							Range: lsp.Range{
								Start: lsp.Position{Line: 15, Character: 20},
								End:   lsp.Position{Line: 15, Character: 30},
							},
							NewText: "renamed",
						},
					},
				},
				{
					TextDocument: lsp.OptionalVersionedTextDocumentIdentifier{
						TextDocumentIdentifier: lsp.TextDocumentIdentifier{URI: "file:///tmp/other.go"},
					},
					Edits: []lsp.TextEdit{
						{
							Range: lsp.Range{
								Start: lsp.Position{Line: 1, Character: 2},
								End:   lsp.Position{Line: 3, Character: 4},
							},
							NewText: "unchanged",
						},
					},
				},
			},
		}
		convertWorkspaceEdit(cache, log, edit)
		if edit.DocumentChanges[0].TextDocument.URI != "file:///tmp/test.templ" {
			t.Errorf("expected URI to be converted, got %q", edit.DocumentChanges[0].TextDocument.URI)
		}
		expectedRange := lsp.Range{
			Start: lsp.Position{Line: 5, Character: 10},
			End:   lsp.Position{Line: 5, Character: 20},
		}
		if diff := cmp.Diff(expectedRange, edit.DocumentChanges[0].Edits[0].Range); diff != "" {
			t.Errorf("unexpected range (-want +got):\n%s", diff)
		}
		if edit.DocumentChanges[1].TextDocument.URI != "file:///tmp/other.go" {
			t.Errorf("expected non-templ URI to be unchanged, got %q", edit.DocumentChanges[1].TextDocument.URI)
		}
	})
	t.Run("converts Changes map", func(t *testing.T) {
		edit := &lsp.WorkspaceEdit{
			Changes: map[lsp.DocumentURI][]lsp.TextEdit{
				"file:///tmp/test_templ.go": {
					{
						Range: lsp.Range{
							Start: lsp.Position{Line: 15, Character: 20},
							End:   lsp.Position{Line: 15, Character: 30},
						},
						NewText: "renamed",
					},
				},
				"file:///tmp/other.go": {
					{
						Range: lsp.Range{
							Start: lsp.Position{Line: 1, Character: 2},
							End:   lsp.Position{Line: 3, Character: 4},
						},
						NewText: "unchanged",
					},
				},
			},
		}
		convertWorkspaceEdit(cache, log, edit)
		if _, exists := edit.Changes["file:///tmp/test_templ.go"]; exists {
			t.Error("expected _templ.go key to be removed from Changes map")
		}
		templEdits, exists := edit.Changes["file:///tmp/test.templ"]
		if !exists {
			t.Fatal("expected .templ key to be present in Changes map")
		}
		expectedRange := lsp.Range{
			Start: lsp.Position{Line: 5, Character: 10},
			End:   lsp.Position{Line: 5, Character: 20},
		}
		if diff := cmp.Diff(expectedRange, templEdits[0].Range); diff != "" {
			t.Errorf("unexpected range (-want +got):\n%s", diff)
		}
		otherEdits, exists := edit.Changes["file:///tmp/other.go"]
		if !exists {
			t.Fatal("expected .go key to be preserved in Changes map")
		}
		if otherEdits[0].NewText != "unchanged" {
			t.Errorf("expected non-templ edit to be unchanged")
		}
	})
}
