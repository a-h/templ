package proxy

import (
	"context"
	"log/slog"
	"testing"

	"github.com/a-h/templ/parser/v2"
	"github.com/google/go-cmp/cmp"

	lsp "github.com/a-h/templ/lsp/protocol"
)

// mockServer records the parameters it receives and returns preconfigured results.
// Only methods under test are implemented; the rest panic.
type mockServer struct {
	// Recorded params from the most recent call.
	definitionParams           *lsp.DefinitionParams
	declarationParams          *lsp.DeclarationParams
	referencesParams           *lsp.ReferenceParams
	renameParams               *lsp.RenameParams
	typeDefinitionParams       *lsp.TypeDefinitionParams
	implementationParams       *lsp.ImplementationParams
	hoverParams                *lsp.HoverParams
	documentHighlightParams    *lsp.DocumentHighlightParams
	prepareRenameParams        *lsp.PrepareRenameParams
	completionParams           *lsp.CompletionParams
	didChangeParams            *lsp.DidChangeTextDocumentParams
	didOpenParams              *lsp.DidOpenTextDocumentParams
	didCloseParams             *lsp.DidCloseTextDocumentParams
	willSaveParams             *lsp.WillSaveTextDocumentParams
	foldingRangesParams        *lsp.FoldingRangeParams
	formattingParams           *lsp.DocumentFormattingParams
	rangeFormattingParams      *lsp.DocumentRangeFormattingParams
	semanticTokensFullParams   *lsp.SemanticTokensParams
	documentLinkParams         *lsp.DocumentLinkParams
	prepareCallHierarchyParams *lsp.CallHierarchyPrepareParams
	incomingCallsParams        *lsp.CallHierarchyIncomingCallsParams
	outgoingCallsParams        *lsp.CallHierarchyOutgoingCallsParams
	symbolsParams              *lsp.WorkspaceSymbolParams
	signatureHelpParams        *lsp.SignatureHelpParams
	codeActionParams           *lsp.CodeActionParams

	// Return values.
	definitionResult           []lsp.Location
	declarationResult          []lsp.Location
	referencesResult           []lsp.Location
	renameResult               *lsp.WorkspaceEdit
	typeDefinitionResult       []lsp.Location
	implementationResult       []lsp.Location
	hoverResult                *lsp.Hover
	documentHighlightResult    []lsp.DocumentHighlight
	prepareRenameResult        *lsp.Range
	completionResult           *lsp.CompletionList
	foldingRangesResult        []lsp.FoldingRange
	formattingResult           []lsp.TextEdit
	rangeFormattingResult      []lsp.TextEdit
	semanticTokensFullResult   *lsp.SemanticTokens
	documentLinkResult         []lsp.DocumentLink
	prepareCallHierarchyResult []lsp.CallHierarchyItem
	incomingCallsResult        []lsp.CallHierarchyIncomingCall
	outgoingCallsResult        []lsp.CallHierarchyOutgoingCall
	symbolsResult              []lsp.SymbolInformation
	signatureHelpResult        *lsp.SignatureHelp
	codeActionResult           []lsp.CodeAction
}

func (m *mockServer) Definition(ctx context.Context, params *lsp.DefinitionParams) ([]lsp.Location, error) {
	m.definitionParams = params
	return m.definitionResult, nil
}

func (m *mockServer) Declaration(ctx context.Context, params *lsp.DeclarationParams) ([]lsp.Location, error) {
	m.declarationParams = params
	return m.declarationResult, nil
}

func (m *mockServer) References(ctx context.Context, params *lsp.ReferenceParams) ([]lsp.Location, error) {
	m.referencesParams = params
	return m.referencesResult, nil
}

func (m *mockServer) Rename(ctx context.Context, params *lsp.RenameParams) (*lsp.WorkspaceEdit, error) {
	m.renameParams = params
	return m.renameResult, nil
}

func (m *mockServer) TypeDefinition(ctx context.Context, params *lsp.TypeDefinitionParams) ([]lsp.Location, error) {
	m.typeDefinitionParams = params
	return m.typeDefinitionResult, nil
}

func (m *mockServer) Implementation(ctx context.Context, params *lsp.ImplementationParams) ([]lsp.Location, error) {
	m.implementationParams = params
	return m.implementationResult, nil
}

func (m *mockServer) Hover(ctx context.Context, params *lsp.HoverParams) (*lsp.Hover, error) {
	m.hoverParams = params
	return m.hoverResult, nil
}

func (m *mockServer) DocumentHighlight(ctx context.Context, params *lsp.DocumentHighlightParams) ([]lsp.DocumentHighlight, error) {
	m.documentHighlightParams = params
	return m.documentHighlightResult, nil
}

func (m *mockServer) PrepareRename(ctx context.Context, params *lsp.PrepareRenameParams) (*lsp.Range, error) {
	m.prepareRenameParams = params
	return m.prepareRenameResult, nil
}

func (m *mockServer) Completion(ctx context.Context, params *lsp.CompletionParams) (*lsp.CompletionList, error) {
	m.completionParams = params
	return m.completionResult, nil
}

func (m *mockServer) DidChange(ctx context.Context, params *lsp.DidChangeTextDocumentParams) error {
	m.didChangeParams = params
	return nil
}

func (m *mockServer) DidOpen(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error {
	m.didOpenParams = params
	return nil
}

func (m *mockServer) DidClose(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error {
	m.didCloseParams = params
	return nil
}

func (m *mockServer) WillSave(ctx context.Context, params *lsp.WillSaveTextDocumentParams) error {
	m.willSaveParams = params
	return nil
}

func (m *mockServer) FoldingRanges(ctx context.Context, params *lsp.FoldingRangeParams) ([]lsp.FoldingRange, error) {
	m.foldingRangesParams = params
	return m.foldingRangesResult, nil
}

func (m *mockServer) Formatting(ctx context.Context, params *lsp.DocumentFormattingParams) ([]lsp.TextEdit, error) {
	m.formattingParams = params
	return m.formattingResult, nil
}

func (m *mockServer) RangeFormatting(ctx context.Context, params *lsp.DocumentRangeFormattingParams) ([]lsp.TextEdit, error) {
	m.rangeFormattingParams = params
	return m.rangeFormattingResult, nil
}

func (m *mockServer) SemanticTokensFull(ctx context.Context, params *lsp.SemanticTokensParams) (*lsp.SemanticTokens, error) {
	m.semanticTokensFullParams = params
	return m.semanticTokensFullResult, nil
}

func (m *mockServer) DocumentLink(ctx context.Context, params *lsp.DocumentLinkParams) ([]lsp.DocumentLink, error) {
	m.documentLinkParams = params
	return m.documentLinkResult, nil
}

func (m *mockServer) PrepareCallHierarchy(ctx context.Context, params *lsp.CallHierarchyPrepareParams) ([]lsp.CallHierarchyItem, error) {
	m.prepareCallHierarchyParams = params
	return m.prepareCallHierarchyResult, nil
}

func (m *mockServer) IncomingCalls(ctx context.Context, params *lsp.CallHierarchyIncomingCallsParams) ([]lsp.CallHierarchyIncomingCall, error) {
	m.incomingCallsParams = params
	return m.incomingCallsResult, nil
}

func (m *mockServer) OutgoingCalls(ctx context.Context, params *lsp.CallHierarchyOutgoingCallsParams) ([]lsp.CallHierarchyOutgoingCall, error) {
	m.outgoingCallsParams = params
	return m.outgoingCallsResult, nil
}

func (m *mockServer) Symbols(ctx context.Context, params *lsp.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	m.symbolsParams = params
	return m.symbolsResult, nil
}

func (m *mockServer) SignatureHelp(ctx context.Context, params *lsp.SignatureHelpParams) (*lsp.SignatureHelp, error) {
	m.signatureHelpParams = params
	return m.signatureHelpResult, nil
}

func (m *mockServer) CodeAction(ctx context.Context, params *lsp.CodeActionParams) ([]lsp.CodeAction, error) {
	m.codeActionParams = params
	return m.codeActionResult, nil
}

// Stubs for interface satisfaction.
func (m *mockServer) Initialize(context.Context, *lsp.InitializeParams) (*lsp.InitializeResult, error) {
	return &lsp.InitializeResult{
		ServerInfo: &lsp.ServerInfo{},
	}, nil
}
func (m *mockServer) Initialized(context.Context, *lsp.InitializedParams) error {
	panic("not implemented")
}
func (m *mockServer) Shutdown(context.Context) error { panic("not implemented") }
func (m *mockServer) Exit(context.Context) error     { panic("not implemented") }
func (m *mockServer) WorkDoneProgressCancel(context.Context, *lsp.WorkDoneProgressCancelParams) error {
	return nil
}
func (m *mockServer) LogTrace(context.Context, *lsp.LogTraceParams) error { return nil }
func (m *mockServer) SetTrace(context.Context, *lsp.SetTraceParams) error { return nil }
func (m *mockServer) CodeLens(context.Context, *lsp.CodeLensParams) ([]lsp.CodeLens, error) {
	return nil, nil
}
func (m *mockServer) CodeLensResolve(context.Context, *lsp.CodeLens) (*lsp.CodeLens, error) {
	return nil, nil
}
func (m *mockServer) ColorPresentation(context.Context, *lsp.ColorPresentationParams) ([]lsp.ColorPresentation, error) {
	return nil, nil
}
func (m *mockServer) CompletionResolve(context.Context, *lsp.CompletionItem) (*lsp.CompletionItem, error) {
	return nil, nil
}
func (m *mockServer) DidChangeConfiguration(context.Context, *lsp.DidChangeConfigurationParams) error {
	return nil
}
func (m *mockServer) DidChangeWatchedFiles(context.Context, *lsp.DidChangeWatchedFilesParams) error {
	return nil
}
func (m *mockServer) DidChangeWorkspaceFolders(context.Context, *lsp.DidChangeWorkspaceFoldersParams) error {
	return nil
}
func (m *mockServer) DidSave(context.Context, *lsp.DidSaveTextDocumentParams) error { return nil }
func (m *mockServer) DocumentColor(context.Context, *lsp.DocumentColorParams) ([]lsp.ColorInformation, error) {
	return nil, nil
}
func (m *mockServer) DocumentLinkResolve(context.Context, *lsp.DocumentLink) (*lsp.DocumentLink, error) {
	return nil, nil
}
func (m *mockServer) DocumentSymbol(context.Context, *lsp.DocumentSymbolParams) ([]lsp.SymbolInformationOrDocumentSymbol, error) {
	return nil, nil
}
func (m *mockServer) ExecuteCommand(context.Context, *lsp.ExecuteCommandParams) (any, error) {
	return nil, nil
}
func (m *mockServer) OnTypeFormatting(context.Context, *lsp.DocumentOnTypeFormattingParams) ([]lsp.TextEdit, error) {
	return nil, nil
}
func (m *mockServer) WillSaveWaitUntil(context.Context, *lsp.WillSaveTextDocumentParams) ([]lsp.TextEdit, error) {
	return nil, nil
}
func (m *mockServer) ShowDocument(context.Context, *lsp.ShowDocumentParams) (*lsp.ShowDocumentResult, error) {
	return nil, nil
}
func (m *mockServer) WillCreateFiles(context.Context, *lsp.CreateFilesParams) (*lsp.WorkspaceEdit, error) {
	return nil, nil
}
func (m *mockServer) DidCreateFiles(context.Context, *lsp.CreateFilesParams) error { return nil }
func (m *mockServer) WillRenameFiles(context.Context, *lsp.RenameFilesParams) (*lsp.WorkspaceEdit, error) {
	return nil, nil
}
func (m *mockServer) DidRenameFiles(context.Context, *lsp.RenameFilesParams) error { return nil }
func (m *mockServer) WillDeleteFiles(context.Context, *lsp.DeleteFilesParams) (*lsp.WorkspaceEdit, error) {
	return nil, nil
}
func (m *mockServer) DidDeleteFiles(context.Context, *lsp.DeleteFilesParams) error { return nil }
func (m *mockServer) CodeLensRefresh(context.Context) error                        { return nil }
func (m *mockServer) SemanticTokensFullDelta(context.Context, *lsp.SemanticTokensDeltaParams) (any, error) {
	return nil, nil
}
func (m *mockServer) SemanticTokensRange(context.Context, *lsp.SemanticTokensRangeParams) (*lsp.SemanticTokens, error) {
	return nil, nil
}
func (m *mockServer) SemanticTokensRefresh(context.Context) error { return nil }
func (m *mockServer) LinkedEditingRange(context.Context, *lsp.LinkedEditingRangeParams) (*lsp.LinkedEditingRanges, error) {
	return nil, nil
}
func (m *mockServer) Moniker(context.Context, *lsp.MonikerParams) ([]lsp.Moniker, error) {
	return nil, nil
}
func (m *mockServer) Request(context.Context, string, any) (any, error) { return nil, nil }

// newTestServer creates a Server with a mock target and a pre-populated source map cache.
// The source map maps templ line 5, cols 10..20 to Go line 15, cols 20..30.
func newTestServer(mock *mockServer) *Server {
	log := slog.Default()
	cache := NewSourceMapCache()
	cache.Set("file:///project/component.templ", newTestSourceMap())
	return &Server{
		Log:             log,
		Target:          mock,
		SourceMapCache:  cache,
		DiagnosticCache: NewDiagnosticCache(),
		TemplSource:     newDocumentContents(log),
		GoSource:        make(map[string]string),
	}
}

func goRange() lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: 15, Character: 20},
		End:   lsp.Position{Line: 15, Character: 30},
	}
}

func templRange() lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: 5, Character: 10},
		End:   lsp.Position{Line: 5, Character: 20},
	}
}

func TestDefinitionFromGoFile(t *testing.T) {
	mock := &mockServer{
		definitionResult: []lsp.Location{
			{URI: "file:///project/component_templ.go", Range: goRange()},
			{URI: "file:///project/other.go", Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 0}, End: lsp.Position{Line: 1, Character: 5}}},
		},
	}
	s := newTestServer(mock)
	result, err := s.Definition(context.Background(), &lsp.DefinitionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
			Position:     lsp.Position{Line: 10, Character: 5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The request should be forwarded unchanged.
	if mock.definitionParams.TextDocument.URI != "file:///project/main.go" {
		t.Errorf("expected URI to be forwarded unchanged, got %q", mock.definitionParams.TextDocument.URI)
	}
	if mock.definitionParams.Position.Line != 10 {
		t.Errorf("expected position to be forwarded unchanged, got %v", mock.definitionParams.Position)
	}
	// The _templ.go result should be converted to .templ.
	if result[0].URI != "file:///project/component.templ" {
		t.Errorf("expected _templ.go URI to be converted, got %q", result[0].URI)
	}
	if diff := cmp.Diff(templRange(), result[0].Range); diff != "" {
		t.Errorf("expected range to be converted (-want +got):\n%s", diff)
	}
	// The non-templ result should be unchanged.
	if result[1].URI != "file:///project/other.go" {
		t.Errorf("expected non-templ URI to be unchanged, got %q", result[1].URI)
	}
}

func TestDefinitionFromTemplFile(t *testing.T) {
	mock := &mockServer{
		definitionResult: []lsp.Location{
			{URI: "file:///project/component_templ.go", Range: goRange()},
		},
	}
	s := newTestServer(mock)
	result, err := s.Definition(context.Background(), &lsp.DefinitionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/component.templ"},
			Position:     lsp.Position{Line: 5, Character: 10},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The request should be rewritten to the Go file.
	if mock.definitionParams.TextDocument.URI != "file:///project/component_templ.go" {
		t.Errorf("expected URI to be converted to Go, got %q", mock.definitionParams.TextDocument.URI)
	}
	if mock.definitionParams.Position.Line != 15 || mock.definitionParams.Position.Character != 20 {
		t.Errorf("expected position to be converted to Go, got %v", mock.definitionParams.Position)
	}
	// The response should be converted back.
	if result[0].URI != "file:///project/component.templ" {
		t.Errorf("expected URI to be converted back, got %q", result[0].URI)
	}
}

func TestRenameFromGoFile(t *testing.T) {
	mock := &mockServer{
		renameResult: &lsp.WorkspaceEdit{
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
					Edits: []lsp.TextEdit{{Range: lsp.Range{Start: lsp.Position{Line: 3, Character: 0}, End: lsp.Position{Line: 3, Character: 7}}, NewText: "newName"}},
				},
			},
		},
	}
	s := newTestServer(mock)
	result, err := s.Rename(context.Background(), &lsp.RenameParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
			Position:     lsp.Position{Line: 10, Character: 5},
		},
		NewName: "newName",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The request should be forwarded unchanged.
	if mock.renameParams.TextDocument.URI != "file:///project/main.go" {
		t.Errorf("expected URI to be forwarded unchanged, got %q", mock.renameParams.TextDocument.URI)
	}
	// The _templ.go document change should be converted to .templ.
	if result.DocumentChanges[0].TextDocument.URI != "file:///project/component.templ" {
		t.Errorf("expected _templ.go URI to be converted, got %q", result.DocumentChanges[0].TextDocument.URI)
	}
	if diff := cmp.Diff(templRange(), result.DocumentChanges[0].Edits[0].Range); diff != "" {
		t.Errorf("expected range to be converted (-want +got):\n%s", diff)
	}
	// The .go document change should be unchanged.
	if result.DocumentChanges[1].TextDocument.URI != "file:///project/main.go" {
		t.Errorf("expected .go URI to be unchanged, got %q", result.DocumentChanges[1].TextDocument.URI)
	}
}

func TestRenameFromTemplFile(t *testing.T) {
	mock := &mockServer{
		renameResult: &lsp.WorkspaceEdit{
			Changes: map[lsp.DocumentURI][]lsp.TextEdit{
				"file:///project/component_templ.go": {{Range: goRange(), NewText: "newName"}},
			},
		},
	}
	s := newTestServer(mock)
	result, err := s.Rename(context.Background(), &lsp.RenameParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/component.templ"},
			Position:     lsp.Position{Line: 5, Character: 10},
		},
		NewName: "newName",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The request should be rewritten to the Go file.
	if mock.renameParams.TextDocument.URI != "file:///project/component_templ.go" {
		t.Errorf("expected URI to be converted to Go, got %q", mock.renameParams.TextDocument.URI)
	}
	// The response should be converted.
	if _, exists := result.Changes["file:///project/component_templ.go"]; exists {
		t.Error("expected _templ.go key to be removed from Changes map")
	}
	edits, exists := result.Changes["file:///project/component.templ"]
	if !exists {
		t.Fatal("expected .templ key to be present in Changes map")
	}
	if diff := cmp.Diff(templRange(), edits[0].Range); diff != "" {
		t.Errorf("expected range to be converted (-want +got):\n%s", diff)
	}
}

func TestTypeDefinitionResponseConverted(t *testing.T) {
	mock := &mockServer{
		typeDefinitionResult: []lsp.Location{
			{URI: "file:///project/component_templ.go", Range: goRange()},
		},
	}
	s := newTestServer(mock)
	result, err := s.TypeDefinition(context.Background(), &lsp.TypeDefinitionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/component.templ"},
			Position:     lsp.Position{Line: 5, Character: 10},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].URI != "file:///project/component.templ" {
		t.Errorf("expected response URI to be converted, got %q", result[0].URI)
	}
}

func TestImplementationNotHardcodedToRequestURI(t *testing.T) {
	mock := &mockServer{
		implementationResult: []lsp.Location{
			{URI: "file:///project/component_templ.go", Range: goRange()},
			{URI: "file:///project/other.go", Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 0}, End: lsp.Position{Line: 1, Character: 5}}},
		},
	}
	s := newTestServer(mock)
	result, err := s.Implementation(context.Background(), &lsp.ImplementationParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/component.templ"},
			Position:     lsp.Position{Line: 5, Character: 10},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The _templ.go result should be converted, not hardcoded to the request URI.
	if result[0].URI != "file:///project/component.templ" {
		t.Errorf("expected _templ.go result to be converted, got %q", result[0].URI)
	}
	// Non-templ results should be left alone, not hardcoded.
	if result[1].URI != "file:///project/other.go" {
		t.Errorf("expected non-templ result to be unchanged, got %q", result[1].URI)
	}
}

func TestDocumentHighlightNotStubbed(t *testing.T) {
	mock := &mockServer{
		documentHighlightResult: []lsp.DocumentHighlight{
			{Range: goRange(), Kind: 1},
			{Range: lsp.Range{Start: lsp.Position{Line: 15, Character: 25}, End: lsp.Position{Line: 15, Character: 30}}, Kind: 2},
		},
	}
	s := newTestServer(mock)
	result, err := s.DocumentHighlight(context.Background(), &lsp.DocumentHighlightParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/component.templ"},
			Position:     lsp.Position{Line: 5, Character: 10},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 highlights, got %d", len(result))
	}
	// Ranges should be converted back to templ positions.
	if result[0].Range.Start.Line != 5 {
		t.Errorf("expected highlight range to be converted, got line %d", result[0].Range.Start.Line)
	}
}

func TestDocumentHighlightDropsGoFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	result, err := s.DocumentHighlight(context.Background(), &lsp.DocumentHighlightParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
			Position:     lsp.Position{Line: 3, Character: 2},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.documentHighlightParams != nil {
		t.Error("expected DocumentHighlight to not be forwarded for .go files")
	}
	if result != nil {
		t.Errorf("expected nil result for .go files, got %v", result)
	}
}

func TestDidChangeForwardsGoFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	err := s.DidChange(context.Background(), &lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
		},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{{Text: "package main"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.didChangeParams == nil {
		t.Fatal("expected DidChange to be forwarded to gopls for .go files")
	}
	if mock.didChangeParams.TextDocument.URI != "file:///project/main.go" {
		t.Errorf("expected URI to be forwarded unchanged, got %q", mock.didChangeParams.TextDocument.URI)
	}
}

func TestDidOpenForwardsGoFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	err := s.HandleDidOpen(context.Background(), &lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{URI: "file:///project/main.go", Text: "package main"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.didOpenParams == nil {
		t.Fatal("expected DidOpen to be forwarded to gopls for .go files")
	}
	if mock.didOpenParams.TextDocument.URI != "file:///project/main.go" {
		t.Errorf("expected URI to be forwarded unchanged, got %q", mock.didOpenParams.TextDocument.URI)
	}
}

func TestDidCloseForwardsGoFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	err := s.HandleDidClose(context.Background(), &lsp.DidCloseTextDocumentParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.didCloseParams == nil {
		t.Fatal("expected DidClose to be forwarded to gopls for .go files")
	}
}

func TestWillSaveForwardsGoFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	err := s.WillSave(context.Background(), &lsp.WillSaveTextDocumentParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.willSaveParams == nil {
		t.Fatal("expected WillSave to be forwarded to gopls for .go files")
	}
	if mock.willSaveParams.TextDocument.URI != "file:///project/main.go" {
		t.Errorf("expected URI to be forwarded unchanged, got %q", mock.willSaveParams.TextDocument.URI)
	}
}

func TestFoldingRangesDropsGoFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	result, err := s.FoldingRanges(context.Background(), &lsp.FoldingRangeParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.foldingRangesParams != nil {
		t.Error("expected FoldingRanges to not be forwarded for .go files")
	}
	if result != nil {
		t.Errorf("expected nil result for .go files, got %v", result)
	}
}

func TestFoldingRangesReturnsEmptyForTemplFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	result, err := s.FoldingRanges(context.Background(), &lsp.FoldingRangeParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/component.templ"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.foldingRangesParams != nil {
		t.Error("expected FoldingRanges to not be forwarded for templ files")
	}
	if len(result) != 0 {
		t.Errorf("expected empty result for templ files, got %v", result)
	}
}

func TestSemanticTokensDropsGoFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	result, err := s.SemanticTokensFull(context.Background(), &lsp.SemanticTokensParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.semanticTokensFullParams != nil {
		t.Error("expected SemanticTokensFull to not be forwarded for .go files")
	}
	if result != nil {
		t.Errorf("expected nil result for .go files, got %v", result)
	}
}

func TestDocumentLinkDropsGoFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	result, err := s.DocumentLink(context.Background(), &lsp.DocumentLinkParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.documentLinkParams != nil {
		t.Error("expected DocumentLink to not be forwarded for .go files")
	}
	if result != nil {
		t.Errorf("expected nil result for .go files, got %v", result)
	}
}

func TestCompletionDropsGoFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	result, err := s.Completion(context.Background(), &lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
			Position:     lsp.Position{Line: 10, Character: 5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.completionParams != nil {
		t.Error("expected Completion to not be forwarded for .go files")
	}
	if result != nil {
		t.Errorf("expected nil result for .go files, got %v", result)
	}
}

func TestFormattingDropsGoFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	result, err := s.Formatting(context.Background(), &lsp.DocumentFormattingParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.formattingParams != nil {
		t.Error("expected Formatting to not be forwarded for .go files")
	}
	if result != nil {
		t.Errorf("expected nil result for .go files, got %v", result)
	}
}

func TestRangeFormattingConvertsRange(t *testing.T) {
	mock := &mockServer{
		rangeFormattingResult: []lsp.TextEdit{{Range: goRange(), NewText: "formatted"}},
	}
	s := newTestServer(mock)
	result, err := s.RangeFormatting(context.Background(), &lsp.DocumentRangeFormattingParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/component.templ"},
		Range:        templRange(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The request range should have been converted.
	if mock.rangeFormattingParams.Range.Start.Line != 15 {
		t.Errorf("expected request range to be converted to Go, got %v", mock.rangeFormattingParams.Range)
	}
	// The response range should be converted back.
	if diff := cmp.Diff(templRange(), result[0].Range); diff != "" {
		t.Errorf("expected response range to be converted back (-want +got):\n%s", diff)
	}
}

func TestRangeFormattingDropsGoFiles(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	result, err := s.RangeFormatting(context.Background(), &lsp.DocumentRangeFormattingParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
		Range:        lsp.Range{Start: lsp.Position{Line: 1, Character: 0}, End: lsp.Position{Line: 5, Character: 0}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.rangeFormattingParams != nil {
		t.Error("expected RangeFormatting to not be forwarded for .go files")
	}
	if result != nil {
		t.Errorf("expected nil result for .go files, got %v", result)
	}
}

func TestReferencesFromGoFile(t *testing.T) {
	mock := &mockServer{
		referencesResult: []lsp.Location{
			{URI: "file:///project/component_templ.go", Range: goRange()},
			{URI: "file:///project/main.go", Range: lsp.Range{Start: lsp.Position{Line: 3, Character: 0}, End: lsp.Position{Line: 3, Character: 5}}},
		},
	}
	s := newTestServer(mock)
	result, err := s.References(context.Background(), &lsp.ReferenceParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
			Position:     lsp.Position{Line: 10, Character: 5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Request forwarded unchanged.
	if mock.referencesParams.TextDocument.URI != "file:///project/main.go" {
		t.Errorf("expected URI forwarded unchanged, got %q", mock.referencesParams.TextDocument.URI)
	}
	// _templ.go reference converted to .templ.
	if result[0].URI != "file:///project/component.templ" {
		t.Errorf("expected _templ.go reference to be converted, got %q", result[0].URI)
	}
	// .go reference unchanged.
	if result[1].URI != "file:///project/main.go" {
		t.Errorf("expected .go reference to be unchanged, got %q", result[1].URI)
	}
}

func TestPrepareCallHierarchyConvertsItems(t *testing.T) {
	mock := &mockServer{
		prepareCallHierarchyResult: []lsp.CallHierarchyItem{
			{
				Name:           "MyComponent",
				URI:            "file:///project/component_templ.go",
				Range:          goRange(),
				SelectionRange: goRange(),
			},
		},
	}
	s := newTestServer(mock)
	result, err := s.PrepareCallHierarchy(context.Background(), &lsp.CallHierarchyPrepareParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/component.templ"},
			Position:     lsp.Position{Line: 5, Character: 10},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].URI != "file:///project/component.templ" {
		t.Errorf("expected URI to be converted, got %q", result[0].URI)
	}
	if diff := cmp.Diff(templRange(), result[0].Range); diff != "" {
		t.Errorf("expected Range to be converted (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(templRange(), result[0].SelectionRange); diff != "" {
		t.Errorf("expected SelectionRange to be converted (-want +got):\n%s", diff)
	}
}

func TestIncomingCallsConvertsFromItems(t *testing.T) {
	mock := &mockServer{
		incomingCallsResult: []lsp.CallHierarchyIncomingCall{
			{
				From: lsp.CallHierarchyItem{
					Name:           "Caller",
					URI:            "file:///project/component_templ.go",
					Range:          goRange(),
					SelectionRange: goRange(),
				},
				FromRanges: []lsp.Range{goRange()},
			},
		},
	}
	s := newTestServer(mock)
	result, err := s.IncomingCalls(context.Background(), &lsp.CallHierarchyIncomingCallsParams{
		Item: lsp.CallHierarchyItem{Name: "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].From.URI != "file:///project/component.templ" {
		t.Errorf("expected From URI to be converted, got %q", result[0].From.URI)
	}
	if diff := cmp.Diff(templRange(), result[0].FromRanges[0]); diff != "" {
		t.Errorf("expected FromRanges to be converted (-want +got):\n%s", diff)
	}
}

func TestOutgoingCallsConvertsToItems(t *testing.T) {
	mock := &mockServer{
		outgoingCallsResult: []lsp.CallHierarchyOutgoingCall{
			{
				To: lsp.CallHierarchyItem{
					Name:           "Callee",
					URI:            "file:///project/component_templ.go",
					Range:          goRange(),
					SelectionRange: goRange(),
				},
			},
		},
	}
	s := newTestServer(mock)
	result, err := s.OutgoingCalls(context.Background(), &lsp.CallHierarchyOutgoingCallsParams{
		Item: lsp.CallHierarchyItem{Name: "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].To.URI != "file:///project/component.templ" {
		t.Errorf("expected To URI to be converted, got %q", result[0].To.URI)
	}
}

func TestSymbolsConvertsTemplGoLocations(t *testing.T) {
	mock := &mockServer{
		symbolsResult: []lsp.SymbolInformation{
			{
				Name:     "MyComponent",
				Location: lsp.Location{URI: "file:///project/component_templ.go", Range: goRange()},
			},
			{
				Name:     "main",
				Location: lsp.Location{URI: "file:///project/main.go", Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 0}, End: lsp.Position{Line: 1, Character: 4}}},
			},
		},
	}
	s := newTestServer(mock)
	result, err := s.Symbols(context.Background(), &lsp.WorkspaceSymbolParams{Query: "My"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].Location.URI != "file:///project/component.templ" {
		t.Errorf("expected _templ.go symbol location to be converted, got %q", result[0].Location.URI)
	}
	if result[1].Location.URI != "file:///project/main.go" {
		t.Errorf("expected .go symbol location to be unchanged, got %q", result[1].Location.URI)
	}
}

func TestSignatureHelpDropsGoFiles(t *testing.T) {
	mock := &mockServer{
		signatureHelpResult: &lsp.SignatureHelp{},
	}
	s := newTestServer(mock)
	result, err := s.SignatureHelp(context.Background(), &lsp.SignatureHelpParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/main.go"},
			Position:     lsp.Position{Line: 10, Character: 5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.signatureHelpParams != nil {
		t.Error("expected SignatureHelp to not be forwarded for .go files")
	}
	if result != nil {
		t.Errorf("expected nil result for .go files, got %v", result)
	}
}

// newTestSourceMapMulti creates a source map with two expressions for testing CodeAction.
func newTestSourceMapMulti() *parser.SourceMap {
	sm := parser.NewSourceMap()
	// Map a whole line range: templ line 3, col 0..10 -> Go line 10, col 0..10.
	sm.Add(
		parser.Expression{
			Value: "0123456789",
			Range: parser.Range{
				From: parser.Position{Index: 0, Line: 3, Col: 0},
				To:   parser.Position{Index: 10, Line: 3, Col: 10},
			},
		},
		parser.Range{
			From: parser.Position{Index: 0, Line: 10, Col: 0},
			To:   parser.Position{Index: 10, Line: 10, Col: 10},
		},
	)
	return sm
}

func TestCodeActionConvertsWorkspaceEdit(t *testing.T) {
	mock := &mockServer{
		codeActionResult: []lsp.CodeAction{
			{
				Title: "Organize Imports",
				Edit: &lsp.WorkspaceEdit{
					DocumentChanges: []lsp.TextDocumentEdit{
						{
							TextDocument: lsp.OptionalVersionedTextDocumentIdentifier{
								TextDocumentIdentifier: lsp.TextDocumentIdentifier{URI: "file:///project/component_templ.go"},
							},
							Edits: []lsp.TextEdit{{Range: lsp.Range{Start: lsp.Position{Line: 10, Character: 0}, End: lsp.Position{Line: 10, Character: 10}}, NewText: "import"}},
						},
					},
				},
			},
			{
				Title: "Unsupported Action",
			},
		},
	}
	cache := NewSourceMapCache()
	cache.Set("file:///project/component.templ", newTestSourceMapMulti())
	log := slog.Default()
	s := &Server{
		Log:             log,
		Target:          mock,
		SourceMapCache:  cache,
		DiagnosticCache: NewDiagnosticCache(),
		TemplSource:     newDocumentContents(log),
		GoSource:        make(map[string]string),
	}
	result, err := s.CodeAction(context.Background(), &lsp.CodeActionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///project/component.templ"},
		Range: lsp.Range{
			Start: lsp.Position{Line: 3, Character: 0},
			End:   lsp.Position{Line: 3, Character: 5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only "Organize Imports" should pass through.
	if len(result) != 1 {
		t.Fatalf("expected 1 code action (unsupported filtered), got %d", len(result))
	}
	if result[0].Title != "Organize Imports" {
		t.Errorf("expected Organize Imports, got %q", result[0].Title)
	}
	// Workspace edit should be converted.
	if result[0].Edit.DocumentChanges[0].TextDocument.URI != "file:///project/component.templ" {
		t.Errorf("expected document change URI to be converted, got %q", result[0].Edit.DocumentChanges[0].TextDocument.URI)
	}
}

func TestInitializeAdvertisesGoFileSupport(t *testing.T) {
	mock := &mockServer{}
	s := newTestServer(mock)
	result, err := s.Initialize(context.Background(), &lsp.InitializeParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	experimental, ok := result.Capabilities.Experimental.(map[string]any)
	if !ok {
		t.Fatal("expected Experimental to be map[string]any")
	}
	templ, ok := experimental["templ"].(map[string]any)
	if !ok {
		t.Fatal("expected experimental.templ to be map[string]any")
	}
	goFileSupport, ok := templ["goFileSupport"].(bool)
	if !ok || !goFileSupport {
		t.Errorf("expected experimental.templ.goFileSupport to be true, got %v", templ["goFileSupport"])
	}
}
