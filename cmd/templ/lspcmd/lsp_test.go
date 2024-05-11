package lspcmd

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/a-h/protocol"
	"github.com/a-h/templ/cmd/templ/generatecmd/modcheck"
	"github.com/a-h/templ/cmd/templ/lspcmd/lspdiff"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/uri"
	"go.uber.org/zap"
)

//go:embed testdata/*
var testdata embed.FS

func createTestProject(moduleRoot string) (dir string, err error) {
	dir, err = os.MkdirTemp("", "templ_watch_test_*")
	if err != nil {
		return dir, fmt.Errorf("failed to make test dir: %w", err)
	}
	files, err := testdata.ReadDir("testdata")
	if err != nil {
		return dir, fmt.Errorf("failed to read embedded dir: %w", err)
	}
	for _, file := range files {
		src := filepath.Join("testdata", file.Name())
		data, err := testdata.ReadFile(src)
		if err != nil {
			return dir, fmt.Errorf("failed to read file: %w", err)
		}

		target := filepath.Join(dir, file.Name())
		if file.Name() == "go.mod.embed" {
			data = bytes.ReplaceAll(data, []byte("{moduleRoot}"), []byte(moduleRoot))
			target = filepath.Join(dir, "go.mod")
		}
		err = os.WriteFile(target, data, 0660)
		if err != nil {
			return dir, fmt.Errorf("failed to copy file: %w", err)
		}
	}
	return dir, nil
}

func mustReplaceLine(file string, line int, replacement string) string {
	lines := strings.Split(file, "\n")
	lines[line-1] = replacement
	return strings.Join(lines, "\n")
}

func TestCompletion(t *testing.T) {
	if testing.Short() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	log, _ := zap.NewProduction()

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
			updated := mustReplaceLine(string(templFile), test.line, test.replacement)
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
			for i := 0; i < 3; i++ {
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
	log := zap.NewNop()

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
						Value: "```go\npackage fmt\n```\n\n[`fmt` on pkg.go.dev](https://pkg.go.dev/fmt)",
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
						Value: "```go\nfunc fmt.Sprintf(format string, a ...any) string\n```\n\nSprintf formats according to a format specifier and returns the resulting string.\n\n\n[`fmt.Sprintf` on pkg.go.dev](https://pkg.go.dev/fmt#Sprintf)",
					},
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
			for i := 0; i < 3; i++ {
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

func NewTestClient(log *zap.Logger) TestClient {
	return TestClient{
		log: log,
	}
}

type TestClient struct {
	log *zap.Logger
}

func (tc TestClient) Progress(ctx context.Context, params *protocol.ProgressParams) (err error) {
	tc.log.Info("client: Received Progress", zap.Any("params", params))
	return nil
}

func (tc TestClient) WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) (err error) {
	tc.log.Info("client: Received WorkDoneProgressCreate", zap.Any("params", params))
	return nil
}

func (tc TestClient) LogMessage(ctx context.Context, params *protocol.LogMessageParams) (err error) {
	tc.log.Info("client: Received LogMessage", zap.Any("params", params))
	return nil
}

func (tc TestClient) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) (err error) {
	tc.log.Info("client: Received PublishDiagnostics", zap.Any("params", params))
	return nil
}

func (tc TestClient) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) (err error) {
	tc.log.Info("client: Received ShowMessage", zap.Any("params", params))
	return nil
}

func (tc TestClient) ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (result *protocol.MessageActionItem, err error) {
	return nil, nil
}

func (tc TestClient) Telemetry(ctx context.Context, params interface{}) (err error) {
	tc.log.Info("client: Received Telemetry", zap.Any("params", params))
	return nil
}

func (tc TestClient) RegisterCapability(ctx context.Context, params *protocol.RegistrationParams,
) (err error) {
	tc.log.Info("client: Received RegisterCapability", zap.Any("params", params))
	return nil
}

func (tc TestClient) UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) (err error) {
	tc.log.Info("client: Received UnregisterCapability", zap.Any("params", params))
	return nil
}

func (tc TestClient) ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (result *protocol.ApplyWorkspaceEditResponse, err error) {
	tc.log.Info("client: Received ApplyEdit", zap.Any("params", params))
	return nil, nil
}

func (tc TestClient) Configuration(ctx context.Context, params *protocol.ConfigurationParams) (result []interface{}, err error) {
	tc.log.Info("client: Received Configuration", zap.Any("params", params))
	return nil, nil
}

func (tc TestClient) WorkspaceFolders(ctx context.Context) (result []protocol.WorkspaceFolder, err error) {
	tc.log.Info("client: Received WorkspaceFolders")
	return nil, nil
}

func Setup(ctx context.Context, log *zap.Logger) (clientCtx context.Context, appDir string, client protocol.Client, server protocol.Server, teardown func(t *testing.T), err error) {
	wd, err := os.Getwd()
	if err != nil {
		return ctx, appDir, client, server, teardown, fmt.Errorf("could not find working dir: %w", err)
	}
	moduleRoot, err := modcheck.WalkUp(wd)
	if err != nil {
		return ctx, appDir, client, server, teardown, fmt.Errorf("could not find local templ go.mod file: %v", err)
	}

	appDir, err = createTestProject(moduleRoot)
	if err != nil {
		return ctx, appDir, client, server, teardown, fmt.Errorf("failed to create test project: %v", err)
	}

	var wg sync.WaitGroup
	var cmdErr error

	// Copy from the LSP to the CLient, and vice versa.
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
			log.Error("Failed to run", zap.Error(cmdErr))
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
		log.Error("Failed to init", zap.Error(err))
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
