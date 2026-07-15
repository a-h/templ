package proxy

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"

	lsp "github.com/a-h/templ/lsp/protocol"
)

// mockClient records the parameters it receives.
type mockClient struct {
	applyEditParams         *lsp.ApplyWorkspaceEditParams
	publishDiagnosticParams *lsp.PublishDiagnosticsParams
}

func (m *mockClient) ApplyEdit(ctx context.Context, params *lsp.ApplyWorkspaceEditParams) (*lsp.ApplyWorkspaceEditResponse, error) {
	m.applyEditParams = params
	return &lsp.ApplyWorkspaceEditResponse{Applied: true}, nil
}

func (m *mockClient) PublishDiagnostics(ctx context.Context, params *lsp.PublishDiagnosticsParams) error {
	m.publishDiagnosticParams = params
	return nil
}

func (m *mockClient) Progress(context.Context, *lsp.ProgressParams) error { return nil }
func (m *mockClient) WorkDoneProgressCreate(context.Context, *lsp.WorkDoneProgressCreateParams) error {
	return nil
}
func (m *mockClient) LogMessage(context.Context, *lsp.LogMessageParams) error   { return nil }
func (m *mockClient) ShowMessage(context.Context, *lsp.ShowMessageParams) error { return nil }
func (m *mockClient) ShowMessageRequest(context.Context, *lsp.ShowMessageRequestParams) (*lsp.MessageActionItem, error) {
	return nil, nil
}
func (m *mockClient) Telemetry(context.Context, any) error                              { return nil }
func (m *mockClient) RegisterCapability(context.Context, *lsp.RegistrationParams) error { return nil }
func (m *mockClient) UnregisterCapability(context.Context, *lsp.UnregistrationParams) error {
	return nil
}
func (m *mockClient) Configuration(context.Context, *lsp.ConfigurationParams) ([]any, error) {
	return nil, nil
}
func (m *mockClient) WorkspaceFolders(context.Context) ([]lsp.WorkspaceFolder, error) {
	return nil, nil
}

func TestApplyEditConvertsWorkspaceEdit(t *testing.T) {
	mock := &mockClient{}
	cache := NewSourceMapCache()
	cache.Set("file:///project/component.templ", newTestSourceMap())
	log := slog.Default()
	client := Client{
		Log:             log,
		Target:          mock,
		SourceMapCache:  cache,
		DiagnosticCache: NewDiagnosticCache(),
	}

	_, err := client.ApplyEdit(context.Background(), &lsp.ApplyWorkspaceEditParams{
		Edit: lsp.WorkspaceEdit{
			DocumentChanges: []lsp.TextDocumentEdit{
				{
					TextDocument: lsp.OptionalVersionedTextDocumentIdentifier{
						TextDocumentIdentifier: lsp.TextDocumentIdentifier{URI: "file:///project/component_templ.go"},
					},
					Edits: []lsp.TextEdit{{Range: goRange(), NewText: "newName"}},
				},
				{
					TextDocument: lsp.OptionalVersionedTextDocumentIdentifier{
						TextDocumentIdentifier: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
					},
					Edits: []lsp.TextEdit{{Range: lsp.Range{Start: lsp.Position{Line: 3, Character: 0}, End: lsp.Position{Line: 3, Character: 5}}, NewText: "newName"}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.applyEditParams == nil {
		t.Fatal("expected ApplyEdit to be forwarded")
	}
	// The _templ.go document change should be converted.
	if mock.applyEditParams.Edit.DocumentChanges[0].TextDocument.URI != "file:///project/component.templ" {
		t.Errorf("expected _templ.go URI to be converted, got %q", mock.applyEditParams.Edit.DocumentChanges[0].TextDocument.URI)
	}
	if diff := cmp.Diff(templRange(), mock.applyEditParams.Edit.DocumentChanges[0].Edits[0].Range); diff != "" {
		t.Errorf("expected range to be converted (-want +got):\n%s", diff)
	}
	// The .go document change should be unchanged.
	if mock.applyEditParams.Edit.DocumentChanges[1].TextDocument.URI != "file:///project/main.go" {
		t.Errorf("expected .go URI to be unchanged, got %q", mock.applyEditParams.Edit.DocumentChanges[1].TextDocument.URI)
	}
}

func TestPublishDiagnosticsDropsGoFiles(t *testing.T) {
	mock := &mockClient{}
	cache := NewSourceMapCache()
	log := slog.Default()
	client := Client{
		Log:             log,
		Target:          mock,
		SourceMapCache:  cache,
		DiagnosticCache: NewDiagnosticCache(),
	}

	err := client.PublishDiagnostics(context.Background(), &lsp.PublishDiagnosticsParams{
		URI: "file:///project/main.go",
		Diagnostics: []lsp.Diagnostic{
			{
				Range:   lsp.Range{Start: lsp.Position{Line: 5, Character: 0}, End: lsp.Position{Line: 5, Character: 10}},
				Message: "undefined variable",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Diagnostics for regular .go files should be dropped because the Go
	// extension publishes those directly.
	if mock.publishDiagnosticParams != nil {
		t.Error("expected PublishDiagnostics to be dropped for .go files")
	}
}

func TestPublishDiagnosticsConvertsTemplGoFiles(t *testing.T) {
	mock := &mockClient{}
	cache := NewSourceMapCache()
	cache.Set("file:///project/component.templ", newTestSourceMap())
	log := slog.Default()
	client := Client{
		Log:             log,
		Target:          mock,
		SourceMapCache:  cache,
		DiagnosticCache: NewDiagnosticCache(),
	}

	err := client.PublishDiagnostics(context.Background(), &lsp.PublishDiagnosticsParams{
		URI: "file:///project/component_templ.go",
		Diagnostics: []lsp.Diagnostic{
			{
				Range:   lsp.Range{Start: lsp.Position{Line: 15, Character: 20}, End: lsp.Position{Line: 15, Character: 30}},
				Message: "unused variable",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.publishDiagnosticParams == nil {
		t.Fatal("expected PublishDiagnostics to be forwarded")
	}
	if mock.publishDiagnosticParams.URI != "file:///project/component.templ" {
		t.Errorf("expected URI to be converted to .templ, got %q", mock.publishDiagnosticParams.URI)
	}
	diag := mock.publishDiagnosticParams.Diagnostics[0]
	if diag.Range.Start.Line != 5 || diag.Range.Start.Character != 10 {
		t.Errorf("expected diagnostic start to be converted to templ position, got %v", diag.Range.Start)
	}
}
