package lspcmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/a-h/templ/cmd/templ/generatecmd/modcheck"
	"github.com/a-h/templ/cmd/templ/lspcmd/lspdiff"
	"github.com/a-h/templ/cmd/templ/testproject"
	"github.com/a-h/templ/lsp/jsonrpc2"
	"github.com/a-h/templ/lsp/protocol"
	"github.com/a-h/templ/lsp/uri"
	"github.com/google/go-cmp/cmp"
)

func TestCompletion(t *testing.T) {
	if testing.Short() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx, appDir, _, server, teardown, err := Setup(ctx, log)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)
	defer cancel()

	templFile, err := os.ReadFile(appDir + "/templates.templ")
	if err != nil {
		t.Errorf("failed to read file %q: %v", appDir+"/templates.templ", err)
		return
	}
	err = server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri.URI("file://" + appDir + "/templates.templ"),
			LanguageID: "templ",
			Version:    1,
			Text:       string(templFile),
		},
	})
	if err != nil {
		t.Errorf("failed to register open file: %v", err)
		return
	}
	log.Info("Calling completion")

	globalSnippetsLen := 1

	// Edit the file.
	// Replace:
	// <div data-testid="count">{ fmt.Sprintf("%d", count) }</div>
	// With various tests:
	// <div data-testid="count">{ f
	tests := []struct {
		line        int
		replacement string
		cursor      string
		assert      func(t *testing.T, cl *protocol.CompletionList) (msg string, ok bool)
	}{
		{
			line:        13,
			replacement: ` <div data-testid="count">{  `,
			cursor:      `                            ^`,
			assert: func(t *testing.T, actual *protocol.CompletionList) (msg string, ok bool) {
				if actual != nil && len(actual.Items) != globalSnippetsLen {
					return "expected completion list to be empty", false
				}
				return "", true
			},
		},
		{
			line:        13,
			replacement: ` <div data-testid="count">{ fmt.`,
			cursor:      `                               ^`,
			assert: func(t *testing.T, actual *protocol.CompletionList) (msg string, ok bool) {
				if !lspdiff.CompletionListContainsText(actual, "fmt.Sprintf") {
					return fmt.Sprintf("expected fmt.Sprintf to be in the completion list, but got %#v", actual), false
				}
				return "", true
			},
		},
		{
			line:        13,
			replacement: ` <div data-testid="count">{ fmt.Sprintf("%d",`,
			cursor:      `                                            ^`,
			assert: func(t *testing.T, actual *protocol.CompletionList) (msg string, ok bool) {
				if actual != nil && len(actual.Items) != globalSnippetsLen {
					return "expected completion list to be empty", false
				}
				return "", true
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			// Edit the file.
			updated := testproject.MustReplaceLine(string(templFile), test.line, test.replacement)
			err = server.DidChange(ctx, &protocol.DidChangeTextDocumentParams{
				TextDocument: protocol.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: protocol.TextDocumentIdentifier{
						URI: uri.URI("file://" + appDir + "/templates.templ"),
					},
					Version: int32(i + 2),
				},
				ContentChanges: []protocol.TextDocumentContentChangeEvent{
					{
						Range: nil,
						Text:  updated,
					},
				},
			})
			if err != nil {
				t.Errorf("failed to change file: %v", err)
				return
			}

			// Give CI/CD pipeline executors some time because they're often quite slow.
			var ok bool
			var msg string
			for range 3 {
				actual, err := server.Completion(ctx, &protocol.CompletionParams{
					Context: &protocol.CompletionContext{
						TriggerCharacter: ".",
						TriggerKind:      protocol.CompletionTriggerKindTriggerCharacter,
					},
					TextDocumentPositionParams: protocol.TextDocumentPositionParams{
						TextDocument: protocol.TextDocumentIdentifier{
							URI: uri.URI("file://" + appDir + "/templates.templ"),
						},
						// Positions are zero indexed.
						Position: protocol.Position{
							Line:      uint32(test.line - 1),
							Character: uint32(len(test.cursor) - 1),
						},
					},
				})
				if err != nil {
					t.Errorf("failed to get completion: %v", err)
					return
				}
				msg, ok = test.assert(t, actual)
				if !ok {
					break
				}
				time.Sleep(time.Millisecond * 500)
			}
			if !ok {
				t.Error(msg)
			}
		})
	}
	log.Info("Completed test")
}

func TestHover(t *testing.T) {
	if testing.Short() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx, appDir, _, server, teardown, err := Setup(ctx, log)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)
	defer cancel()

	templFile, err := os.ReadFile(appDir + "/templates.templ")
	if err != nil {
		t.Fatalf("failed to read file %q: %v", appDir+"/templates.templ", err)
	}
	err = server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri.URI("file://" + appDir + "/templates.templ"),
			LanguageID: "templ",
			Version:    1,
			Text:       string(templFile),
		},
	})
	if err != nil {
		t.Errorf("failed to register open file: %v", err)
		return
	}
	log.Info("Calling hover")

	// Edit the file.
	// Replace:
	// <div data-testid="count">{ fmt.Sprintf("%d", count) }</div>
	// With various tests:
	// <div data-testid="count">{ f
	tests := []struct {
		line        int
		replacement string
		cursor      string
		assert      func(t *testing.T, hr *protocol.Hover) (msg string, ok bool)
	}{
		{
			line:        13,
			replacement: `			<div data-testid="count">{ fmt.Sprintf("%d", count) }</div>`,
			cursor:      `                                 ^`,
			assert: func(t *testing.T, actual *protocol.Hover) (msg string, ok bool) {
				expectedHover := protocol.Hover{
					Contents: protocol.MarkupContent{
						Kind:  "markdown",
						Value: "```go\npackage fmt\n```\n\n---\n\n[`fmt` on pkg.go.dev](https://pkg.go.dev/fmt)",
					},
				}
				if diff := lspdiff.Hover(expectedHover, *actual); diff != "" {
					return fmt.Sprintf("unexpected hover: %v\n\n: markdown: %#v", diff, actual.Contents.Value), false
				}
				return "", true
			},
		},
		{
			line:        13,
			replacement: `			<div data-testid="count">{ fmt.Sprintf("%d", count) }</div>`,
			cursor:      `                                     ^`,
			assert: func(t *testing.T, actual *protocol.Hover) (msg string, ok bool) {
				expectedHover := protocol.Hover{
					Contents: protocol.MarkupContent{
						Kind:  "markdown",
						Value: "```go\nfunc fmt.Sprintf(format string, a ...any) string\n```\n\n---\n\nSprintf formats according to a format specifier and returns the resulting string.\n\n\n---\n\n[`fmt.Sprintf` on pkg.go.dev](https://pkg.go.dev/fmt#Sprintf)",
					},
				}
				if actual == nil {
					return "expected hover to be non-nil", false
				}
				if diff := lspdiff.Hover(expectedHover, *actual); diff != "" {
					return fmt.Sprintf("unexpected hover: %v", diff), false
				}
				return "", true
			},
		},
		{
			line:        19,
			replacement: `var nihao = "你好"`,
			cursor:      `             ^`,
			assert: func(t *testing.T, actual *protocol.Hover) (msg string, ok bool) {
				// There's nothing to hover, just want to make sure it doesn't panic.
				return "", true
			},
		},
		{
			line:        19,
			replacement: `var nihao = "你好"`,
			cursor:      `              ^`, // Your text editor might not render this well, but it's the hao.
			assert: func(t *testing.T, actual *protocol.Hover) (msg string, ok bool) {
				// There's nothing to hover, just want to make sure it doesn't panic.
				return "", true
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			// Put the file back to the initial point.
			err = server.DidChange(ctx, &protocol.DidChangeTextDocumentParams{
				TextDocument: protocol.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: protocol.TextDocumentIdentifier{
						URI: uri.URI("file://" + appDir + "/templates.templ"),
					},
					Version: int32(i + 2),
				},
				ContentChanges: []protocol.TextDocumentContentChangeEvent{
					{
						Range: nil,
						Text:  string(templFile),
					},
				},
			})
			if err != nil {
				t.Errorf("failed to change file: %v", err)
				return
			}

			// Give CI/CD pipeline executors some time because they're often quite slow.
			var ok bool
			var msg string
			for range 3 {
				lspCharIndex, err := runeIndexToUTF8ByteIndex(test.replacement, len(test.cursor)-1)
				if err != nil {
					t.Error(err)
				}
				actual, err := server.Hover(ctx, &protocol.HoverParams{
					TextDocumentPositionParams: protocol.TextDocumentPositionParams{
						TextDocument: protocol.TextDocumentIdentifier{
							URI: uri.URI("file://" + appDir + "/templates.templ"),
						},
						// Positions are zero indexed.
						Position: protocol.Position{
							Line:      uint32(test.line - 1),
							Character: lspCharIndex,
						},
					},
				})
				if err != nil {
					t.Errorf("failed to hover: %v", err)
					return
				}
				msg, ok = test.assert(t, actual)
				if !ok {
					break
				}
				time.Sleep(time.Millisecond * 500)
			}
			if !ok {
				t.Error(msg)
			}
		})
	}
}

func TestReferences(t *testing.T) {
	if testing.Short() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx, appDir, _, server, teardown, err := Setup(ctx, log)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
		return
	}
	defer teardown(t)
	defer cancel()

	log.Info("Calling References")

	tests := []struct {
		line      int
		character int
		filename  string
		assert    func(t *testing.T, l []protocol.Location) (msg string, ok bool)
	}{
		{
			// this is the definition of the templ function in the templates.templ file.
			line:      5,
			character: 9,
			filename:  "/templates.templ",
			assert: func(t *testing.T, actual []protocol.Location) (msg string, ok bool) {
				expectedReference := []protocol.Location{
					{
						// This is the usage of the templ function in the main.go file.
						URI: uri.URI("file://" + appDir + "/main.go"),
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(24),
								Character: uint32(7),
							},
							End: protocol.Position{
								Line:      uint32(24),
								Character: uint32(11),
							},
						},
					},
				}
				if diff := lspdiff.References(expectedReference, actual); diff != "" {
					return fmt.Sprintf("Expected: %+v\nActual: %+v", expectedReference, actual), false
				}
				return "", true
			},
		},
		{
			// this is the definition of the struct in the templates.templ file.
			line:      21,
			character: 9,
			filename:  "/templates.templ",
			assert: func(t *testing.T, actual []protocol.Location) (msg string, ok bool) {
				expectedReference := []protocol.Location{
					{
						// This is the usage of the struct in the templates.templ file.
						URI: uri.URI("file://" + appDir + "/templates.templ"),
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(24),
								Character: uint32(8),
							},
							End: protocol.Position{
								Line:      uint32(24),
								Character: uint32(14),
							},
						},
					},
				}
				if diff := lspdiff.References(expectedReference, actual); diff != "" {
					return fmt.Sprintf("Expected: %+v\nActual: %+v", expectedReference, actual), false
				}
				return "", true
			},
		},
		{
			// this test is for inclusions from a remote file that has not been explicitly called with didOpen
			line:      3,
			character: 9,
			filename:  "/remotechild.templ",
			assert: func(t *testing.T, actual []protocol.Location) (msg string, ok bool) {
				expectedReference := []protocol.Location{
					{
						URI: uri.URI("file://" + appDir + "/remoteparent.templ"),
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(3),
								Character: uint32(2),
							},
							End: protocol.Position{
								Line:      uint32(3),
								Character: uint32(8),
							},
						},
					},
					{
						URI: uri.URI("file://" + appDir + "/remoteparent.templ"),
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(7),
								Character: uint32(2),
							},
							End: protocol.Position{
								Line:      uint32(7),
								Character: uint32(8),
							},
						},
					},
				}
				if diff := lspdiff.References(expectedReference, actual); diff != "" {
					return fmt.Sprintf("Expected: %+v\nActual: %+v", expectedReference, actual), false
				}
				return "", true
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			// Give CI/CD pipeline executors some time because they're often quite slow.
			var ok bool
			var msg string
			for range 3 {
				if err != nil {
					t.Error(err)
					return
				}
				actual, err := server.References(ctx, &protocol.ReferenceParams{
					TextDocumentPositionParams: protocol.TextDocumentPositionParams{
						TextDocument: protocol.TextDocumentIdentifier{
							URI: uri.URI("file://" + appDir + test.filename),
						},
						// Positions are zero indexed.
						Position: protocol.Position{
							Line:      uint32(test.line - 1),
							Character: uint32(test.character - 1),
						},
					},
				})
				if err != nil {
					t.Errorf("failed to get references: %v", err)
					return
				}
				msg, ok = test.assert(t, actual)
				if !ok {
					break
				}
				time.Sleep(time.Millisecond * 500)
			}
			if !ok {
				t.Error(msg)
			}
		})
	}
}

func TestCodeAction(t *testing.T) {
	if testing.Short() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx, appDir, _, server, teardown, err := Setup(ctx, log)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)
	defer cancel()

	templFile, err := os.ReadFile(appDir + "/templates.templ")
	if err != nil {
		t.Fatalf("failed to read file %q: %v", appDir+"/templates.templ", err)
	}
	err = server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri.URI("file://" + appDir + "/templates.templ"),
			LanguageID: "templ",
			Version:    1,
			Text:       string(templFile),
		},
	})
	if err != nil {
		t.Errorf("failed to register open file: %v", err)
		return
	}
	log.Info("Calling codeAction")

	tests := []struct {
		line        int
		replacement string
		cursor      string
		assert      func(t *testing.T, hr []protocol.CodeAction) (msg string, ok bool)
	}{
		{
			line:        25,
			replacement: `var s = Struct{}`,
			cursor:      `              ^`,
			assert: func(t *testing.T, actual []protocol.CodeAction) (msg string, ok bool) {
				var expected []protocol.CodeAction
				// To support code actions, update cmd/templ/lspcmd/proxy/server.go and add the
				// Title (e.g. Organize Imports, or Fill Struct) to the supportedCodeActions map.

				// Some Code Actions are simple edits, so all that is needed is for the server
				// to remap the source code positions.

				// However, other Code Actions are commands, where the arguments must be rewritten
				// and will need to be handled individually.
				if diff := lspdiff.CodeAction(expected, actual); diff != "" {
					return fmt.Sprintf("unexpected codeAction: %v", diff), false
				}
				return "", true
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			// Put the file back to the initial point.
			err = server.DidChange(ctx, &protocol.DidChangeTextDocumentParams{
				TextDocument: protocol.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: protocol.TextDocumentIdentifier{
						URI: uri.URI("file://" + appDir + "/templates.templ"),
					},
					Version: int32(i + 2),
				},
				ContentChanges: []protocol.TextDocumentContentChangeEvent{
					{
						Range: nil,
						Text:  string(templFile),
					},
				},
			})
			if err != nil {
				t.Errorf("failed to change file: %v", err)
				return
			}

			// Give CI/CD pipeline executors some time because they're often quite slow.
			var ok bool
			var msg string
			for range 3 {
				lspCharIndex, err := runeIndexToUTF8ByteIndex(test.replacement, len(test.cursor)-1)
				if err != nil {
					t.Error(err)
				}
				actual, err := server.CodeAction(ctx, &protocol.CodeActionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: uri.URI("file://" + appDir + "/templates.templ"),
					},
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      uint32(test.line - 1),
							Character: lspCharIndex,
						},
						End: protocol.Position{
							Line:      uint32(test.line - 1),
							Character: lspCharIndex + 1,
						},
					},
				})
				if err != nil {
					t.Errorf("failed code action: %v", err)
					return
				}
				msg, ok = test.assert(t, actual)
				if !ok {
					break
				}
				time.Sleep(time.Millisecond * 500)
			}
			if !ok {
				t.Error(msg)
			}
		})
	}
}

func TestDocumentSymbol(t *testing.T) {
	if testing.Short() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx, appDir, _, server, teardown, err := Setup(ctx, log)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)
	defer cancel()

	tests := []struct {
		uri    string
		expect []protocol.SymbolInformationOrDocumentSymbol
	}{
		{
			uri: "file://" + appDir + "/templates.templ",
			expect: []protocol.SymbolInformationOrDocumentSymbol{
				{
					SymbolInformation: &protocol.SymbolInformation{
						Name: "Page",
						Kind: protocol.SymbolKindFunction,
						Location: protocol.Location{
							Range: protocol.Range{
								Start: protocol.Position{Line: 11, Character: 0},
								End:   protocol.Position{Line: 50, Character: 1},
							},
						},
					},
				},
				{
					SymbolInformation: &protocol.SymbolInformation{
						Name: "nihao",
						Kind: protocol.SymbolKindVariable,
						Location: protocol.Location{
							Range: protocol.Range{
								Start: protocol.Position{Line: 18, Character: 4},
								End:   protocol.Position{Line: 18, Character: 16},
							},
						},
					},
				},
				{
					SymbolInformation: &protocol.SymbolInformation{
						Name: "Struct",
						Kind: protocol.SymbolKindStruct,
						Location: protocol.Location{
							Range: protocol.Range{
								Start: protocol.Position{Line: 20, Character: 5},
								End:   protocol.Position{Line: 22, Character: 1},
							},
						},
					},
				},
				{
					SymbolInformation: &protocol.SymbolInformation{
						Name: "s",
						Kind: protocol.SymbolKindVariable,
						Location: protocol.Location{
							Range: protocol.Range{
								Start: protocol.Position{Line: 24, Character: 4},
								End:   protocol.Position{Line: 24, Character: 16},
							},
						},
					},
				},
			},
		},
		{
			uri: "file://" + appDir + "/remoteparent.templ",
			expect: []protocol.SymbolInformationOrDocumentSymbol{
				{
					SymbolInformation: &protocol.SymbolInformation{
						Name: "RemoteInclusionTest",
						Kind: protocol.SymbolKindFunction,
						Location: protocol.Location{
							Range: protocol.Range{
								Start: protocol.Position{Line: 9, Character: 0},
								End:   protocol.Position{Line: 35, Character: 1},
							},
						},
					},
				},
				{
					SymbolInformation: &protocol.SymbolInformation{
						Name: "Remote2",
						Kind: protocol.SymbolKindFunction,
						Location: protocol.Location{
							Range: protocol.Range{
								Start: protocol.Position{Line: 37, Character: 0},
								End:   protocol.Position{Line: 63, Character: 1},
							},
						},
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			actual, err := server.DocumentSymbol(ctx, &protocol.DocumentSymbolParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: uri.URI(test.uri),
				},
			})
			if err != nil {
				t.Errorf("failed to get document symbol: %v", err)
			}

			// Set expected URI.
			for i, v := range test.expect {
				if v.SymbolInformation != nil {
					v.SymbolInformation.Location.URI = uri.URI(test.uri)
					test.expect[i] = v
				}
			}

			if err != nil {
				t.Errorf("failed to convert expect to any slice: %v", err)
			}
			diff := cmp.Diff(test.expect, actual)
			if diff != "" {
				t.Errorf("unexpected document symbol: %v", diff)
			}
		})
	}
}

func runeIndexToUTF8ByteIndex(s string, runeIndex int) (lspChar uint32, err error) {
	for i, r := range []rune(s) {
		if i == runeIndex {
			break
		}
		l := utf8.RuneLen(r)
		if l < 0 {
			return 0, fmt.Errorf("invalid rune in string at index %d", runeIndex)
		}
		lspChar += uint32(l)
	}
	return lspChar, nil
}

func NewTestClient(log *slog.Logger) TestClient {
	return TestClient{
		log: log,
	}
}

type TestClient struct {
	log *slog.Logger
}

func (tc TestClient) Progress(ctx context.Context, params *protocol.ProgressParams) (err error) {
	tc.log.Info("client: Received Progress", slog.Any("params", params))
	return nil
}

func (tc TestClient) WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) (err error) {
	tc.log.Info("client: Received WorkDoneProgressCreate", slog.Any("params", params))
	return nil
}

func (tc TestClient) LogMessage(ctx context.Context, params *protocol.LogMessageParams) (err error) {
	tc.log.Info("client: Received LogMessage", slog.Any("params", params))
	return nil
}

func (tc TestClient) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) (err error) {
	tc.log.Info("client: Received PublishDiagnostics", slog.Any("params", params))
	return nil
}

func (tc TestClient) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) (err error) {
	tc.log.Info("client: Received ShowMessage", slog.Any("params", params))
	return nil
}

func (tc TestClient) ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (result *protocol.MessageActionItem, err error) {
	return nil, nil
}

func (tc TestClient) Telemetry(ctx context.Context, params any) (err error) {
	tc.log.Info("client: Received Telemetry", slog.Any("params", params))
	return nil
}

func (tc TestClient) RegisterCapability(ctx context.Context, params *protocol.RegistrationParams,
) (err error) {
	tc.log.Info("client: Received RegisterCapability", slog.Any("params", params))
	return nil
}

func (tc TestClient) UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) (err error) {
	tc.log.Info("client: Received UnregisterCapability", slog.Any("params", params))
	return nil
}

func (tc TestClient) ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (result *protocol.ApplyWorkspaceEditResponse, err error) {
	tc.log.Info("client: Received ApplyEdit", slog.Any("params", params))
	return nil, nil
}

func (tc TestClient) Configuration(ctx context.Context, params *protocol.ConfigurationParams) (result []any, err error) {
	tc.log.Info("client: Received Configuration", slog.Any("params", params))
	return nil, nil
}

func (tc TestClient) WorkspaceFolders(ctx context.Context) (result []protocol.WorkspaceFolder, err error) {
	tc.log.Info("client: Received WorkspaceFolders")
	return nil, nil
}

func Setup(ctx context.Context, log *slog.Logger) (clientCtx context.Context, appDir string, client protocol.Client, server protocol.Server, teardown func(t *testing.T), err error) {
	wd, err := os.Getwd()
	if err != nil {
		return ctx, appDir, client, server, teardown, fmt.Errorf("could not find working dir: %w", err)
	}
	moduleRoot, err := modcheck.WalkUp(wd)
	if err != nil {
		return ctx, appDir, client, server, teardown, fmt.Errorf("could not find local templ go.mod file: %v", err)
	}

	appDir, err = testproject.Create(moduleRoot)
	if err != nil {
		return ctx, appDir, client, server, teardown, fmt.Errorf("failed to create test project: %v", err)
	}

	var wg sync.WaitGroup
	var cmdErr error

	// Copy from the LSP to the Client, and vice versa.
	fromClient, toLSP := io.Pipe()
	fromLSP, toClient := io.Pipe()
	clientStream := jsonrpc2.NewStream(newStdRwc(log, "clientStream", toLSP, fromLSP))
	serverStream := jsonrpc2.NewStream(newStdRwc(log, "serverStream", toClient, fromClient))

	// Create the client that the server needs.
	client = NewTestClient(log)
	ctx, _, server = protocol.NewClient(ctx, client, clientStream, log)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("Running")
		// Create the server that the client needs.
		cmdErr = run(ctx, log, serverStream, Arguments{})
		if cmdErr != nil {
			log.Error("Failed to run", slog.Any("error", cmdErr))
		}
		log.Info("Stopped")
	}()

	// Initialize.
	ir, err := server.Initialize(ctx, &protocol.InitializeParams{
		ClientInfo: &protocol.ClientInfo{},
		Capabilities: protocol.ClientCapabilities{
			Workspace: &protocol.WorkspaceClientCapabilities{
				ApplyEdit: true,
				WorkspaceEdit: &protocol.WorkspaceClientCapabilitiesWorkspaceEdit{
					DocumentChanges: true,
				},
				WorkspaceFolders: true,
				FileOperations: &protocol.WorkspaceClientCapabilitiesFileOperations{
					DidCreate:  true,
					WillCreate: true,
					DidRename:  true,
					WillRename: true,
					DidDelete:  true,
					WillDelete: true,
				},
			},
			TextDocument: &protocol.TextDocumentClientCapabilities{
				Synchronization: &protocol.TextDocumentSyncClientCapabilities{
					DidSave: true,
				},
				Completion: &protocol.CompletionTextDocumentClientCapabilities{
					CompletionItem: &protocol.CompletionTextDocumentClientCapabilitiesItem{
						SnippetSupport:       true,
						DeprecatedSupport:    true,
						InsertReplaceSupport: true,
					},
				},
				Hover:              &protocol.HoverTextDocumentClientCapabilities{},
				SignatureHelp:      &protocol.SignatureHelpTextDocumentClientCapabilities{},
				Declaration:        &protocol.DeclarationTextDocumentClientCapabilities{},
				Definition:         &protocol.DefinitionTextDocumentClientCapabilities{},
				TypeDefinition:     &protocol.TypeDefinitionTextDocumentClientCapabilities{},
				Implementation:     &protocol.ImplementationTextDocumentClientCapabilities{},
				References:         &protocol.ReferencesTextDocumentClientCapabilities{},
				DocumentHighlight:  &protocol.DocumentHighlightClientCapabilities{},
				DocumentSymbol:     &protocol.DocumentSymbolClientCapabilities{},
				CodeAction:         &protocol.CodeActionClientCapabilities{},
				CodeLens:           &protocol.CodeLensClientCapabilities{},
				Formatting:         &protocol.DocumentFormattingClientCapabilities{},
				RangeFormatting:    &protocol.DocumentRangeFormattingClientCapabilities{},
				OnTypeFormatting:   &protocol.DocumentOnTypeFormattingClientCapabilities{},
				PublishDiagnostics: &protocol.PublishDiagnosticsClientCapabilities{},
				Rename:             &protocol.RenameClientCapabilities{},
				FoldingRange:       &protocol.FoldingRangeClientCapabilities{},
				SelectionRange:     &protocol.SelectionRangeClientCapabilities{},
				CallHierarchy:      &protocol.CallHierarchyClientCapabilities{},
				SemanticTokens:     &protocol.SemanticTokensClientCapabilities{},
				LinkedEditingRange: &protocol.LinkedEditingRangeClientCapabilities{},
			},
			Window:       &protocol.WindowClientCapabilities{},
			General:      &protocol.GeneralClientCapabilities{},
			Experimental: nil,
		},
		WorkspaceFolders: []protocol.WorkspaceFolder{
			{
				URI:  "file://" + appDir,
				Name: "templ-test",
			},
		},
	})
	if err != nil {
		log.Error("Failed to init", slog.Any("error", err))
	}
	if ir.ServerInfo.Name != "templ-lsp" {
		return ctx, appDir, client, server, teardown, fmt.Errorf("expected server name to be templ-lsp, got %q", ir.ServerInfo.Name)
	}

	// Confirm initialization.
	log.Info("Confirming initialization...")
	if err = server.Initialized(ctx, &protocol.InitializedParams{}); err != nil {
		return ctx, appDir, client, server, teardown, fmt.Errorf("failed to confirm initialization: %v", err)
	}
	log.Info("Initialized")

	// Wait for exit.
	teardown = func(t *testing.T) {
		log.Info("Tearing down LSP")
		wg.Wait()
		if cmdErr != nil {
			t.Errorf("failed to run lsp cmd: %v", err)
		}

		if err = os.RemoveAll(appDir); err != nil {
			t.Errorf("failed to remove test dir %q: %v", appDir, err)
		}
	}
	return ctx, appDir, client, server, teardown, err
}
