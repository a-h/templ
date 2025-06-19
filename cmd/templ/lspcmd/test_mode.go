package lspcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/a-h/templ/cmd/templ/lspcmd/pls"
	"github.com/a-h/templ/cmd/templ/lspcmd/proxy"
	"github.com/a-h/templ/internal/skipdir"
	"github.com/a-h/templ/lsp/jsonrpc2"
	"github.com/a-h/templ/lsp/protocol"
)

func runTestMode(ctx context.Context, args Arguments) error {
	stderr := os.Stderr
	stdout := os.Stdout
	log := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Convert to absolute path
	absDir, err := filepath.Abs(args.TestDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	log.Info("lsp: running in test mode", slog.String("dir", absDir))

	// Start gopls
	log.Info("lsp: starting gopls...")
	rwc, err := pls.NewGopls(ctx, log, pls.Options{
		Log:      args.GoplsLog,
		RPCTrace: args.GoplsRPCTrace,
		Remote:   args.GoplsRemote,
	})
	if err != nil {
		return fmt.Errorf("failed to start gopls: %w", err)
	}
	defer func() { _ = rwc.Close() }()

	// Create proxy and clients
	cache := proxy.NewSourceMapCache()
	diagnosticCache := proxy.NewDiagnosticCache()

	clientProxy, clientInit := proxy.NewClient(log, cache, diagnosticCache)
	_, goplsConn, goplsServer := protocol.NewClient(ctx, clientProxy, jsonrpc2.NewStream(rwc), log)
	defer func() { _ = goplsConn.Close() }()

	serverProxy := proxy.NewServer(log, goplsServer, cache, diagnosticCache, args.NoPreload)

	// Create pipes for bidirectional communication
	fromClient, toServer := io.Pipe()
	fromServer, toClient := io.Pipe()

	// Create streams
	clientStream := jsonrpc2.NewStream(newStdRwc(log, "testClientStream", toServer, fromServer))
	serverStream := jsonrpc2.NewStream(newStdRwc(log, "testServerStream", toClient, fromClient))

	// Create the client
	testClient := &testModeClient{log: log, stdout: stdout}
	ctx, clientConn, server := protocol.NewClient(ctx, testClient, clientStream, log)
	defer func() { _ = clientConn.Close() }()

	// Create the server connection
	_, serverConn, templClient := protocol.NewServer(ctx, serverProxy, serverStream, log)
	defer func() { _ = serverConn.Close() }()

	clientInit(templClient)

	initializeResult, err := server.Initialize(ctx, initializeParams(absDir))
	if err != nil {
		return fmt.Errorf("failed to send initialize: %w", err)
	}
	resultJSON, err := json.MarshalIndent(initializeResult, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal initialize result: %w", err)
	}
	_, _ = fmt.Fprintf(stdout, "Initialize result:\n%s\n", resultJSON)

	if err := server.Initialized(ctx, &protocol.InitializedParams{}); err != nil {
		return fmt.Errorf("failed to send initialized: %w", err)
	}

	// If a test request is provided, execute it
	if args.TestRequest != "" {

		var requestParams map[string]any
		if err := json.Unmarshal([]byte(args.TestRequest), &requestParams); err != nil {
			return fmt.Errorf("failed to unmarshal test request params: %w", err)
		}
		indented, err := json.MarshalIndent(requestParams, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal test request params: %w", err)
		}
		_, _ = fmt.Fprintf(stdout, "Executing test request: %s\n", string(indented))

		// TODO: Implement support for test requests
		return nil
		// server.Request(ctx, )
		// // For now, just support specific common requests
		// // We can extend this later to support arbitrary JSON requests
		// switch args.TestRequest {
		// case "diagnostics":
		// 	// Wait a bit for diagnostics to be computed
		// 	log.Info("lsp: waiting for diagnostics...")
		// 	select {
		// 	case <-time.After(2 * time.Second):
		// 		fmt.Fprintf(stdout, "\nDiagnostics collection complete.\n")
		// 	case <-ctx.Done():
		// 		return ctx.Err()
		// 	}
		// }
		// return fmt.Errorf("unsupported test request: %s", args.TestRequest)
	}

	_, _ = fmt.Fprintf(stdout, "No test request provided. Loading directory %s...\n", absDir)

	pattern := `(.+\.go$)|(.+\.templ$)`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("failed to compile regex pattern %q: %w", pattern, err)
	}
	fileSystem := os.DirFS(absDir)
	err = fs.WalkDir(fileSystem, ".", func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		absPath, err := filepath.Abs(filepath.Join(absDir, path))
		if err != nil {
			return nil
		}
		if info.IsDir() && skipdir.ShouldSkip(absPath) {
			return filepath.SkipDir
		}
		if !re.MatchString(absPath) {
			return nil
		}
		content, err := os.ReadFile(absPath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", absPath, err)
		}
		// Load the directory to trigger diagnostics and other initializations
		didOpenParams := &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{
				// Type: protocol.FileChangeTypeCreated,
				URI:        protocol.DocumentURI("file://" + absPath),
				LanguageID: "templ",
				Version:    1,
				Text:       string(content),
			},
		}
		if err := server.DidOpen(ctx, didOpenParams); err != nil {
			return fmt.Errorf("failed to load directory: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory %s: %w", absDir, err)
	}

	return nil
}

// testModeClient is a minimal LSP client for test mode
type testModeClient struct {
	log    *slog.Logger
	stdout io.Writer
}

func (c *testModeClient) Progress(ctx context.Context, params *protocol.ProgressParams) error {
	c.log.Debug("client: Progress", slog.Any("params", params))
	return nil
}

func (c *testModeClient) WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) error {
	c.log.Debug("client: WorkDoneProgressCreate", slog.Any("params", params))
	return nil
}

func (c *testModeClient) LogMessage(ctx context.Context, params *protocol.LogMessageParams) error {
	c.log.Info("client: LogMessage", slog.String("message", params.Message))
	return nil
}

func (c *testModeClient) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) error {
	c.log.Info("client: PublishDiagnostics", slog.String("uri", string(params.URI)), slog.Int("count", len(params.Diagnostics)))
	// Print diagnostics to stdout
	if len(params.Diagnostics) > 0 {
		_, _ = fmt.Fprintf(c.stdout, "\nDiagnostics for %s:\n", params.URI)
		for _, diag := range params.Diagnostics {
			_, _ = fmt.Fprintf(c.stdout, "  - [%d:%d] %s\n", diag.Range.Start.Line, diag.Range.Start.Character, diag.Message)
		}
	}
	return nil
}

func (c *testModeClient) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) error {
	c.log.Info("client: ShowMessage", slog.String("message", params.Message))
	return nil
}

func (c *testModeClient) ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (*protocol.MessageActionItem, error) {
	c.log.Info("client: ShowMessageRequest", slog.String("message", params.Message))
	return nil, nil
}

func (c *testModeClient) Telemetry(ctx context.Context, params any) error {
	c.log.Debug("client: Telemetry", slog.Any("params", params))
	return nil
}

func (c *testModeClient) RegisterCapability(ctx context.Context, params *protocol.RegistrationParams) error {
	c.log.Debug("client: RegisterCapability", slog.Any("params", params))
	return nil
}

func (c *testModeClient) UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) error {
	c.log.Debug("client: UnregisterCapability", slog.Any("params", params))
	return nil
}

func (c *testModeClient) ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (*protocol.ApplyWorkspaceEditResponse, error) {
	c.log.Info("client: ApplyEdit", slog.Any("params", params))
	return &protocol.ApplyWorkspaceEditResponse{Applied: true}, nil
}

func (c *testModeClient) Configuration(ctx context.Context, params *protocol.ConfigurationParams) ([]any, error) {
	c.log.Debug("client: Configuration", slog.Any("params", params))
	return nil, nil
}

func (c *testModeClient) WorkspaceFolders(ctx context.Context) ([]protocol.WorkspaceFolder, error) {
	c.log.Debug("client: WorkspaceFolders")
	return nil, nil
}

func initializeParams(rootPath string) *protocol.InitializeParams {
	return &protocol.InitializeParams{
		Capabilities: protocol.ClientCapabilities{
			TextDocument: &protocol.TextDocumentClientCapabilities{
				Completion: &protocol.CompletionTextDocumentClientCapabilities{
					CompletionItem: &protocol.CompletionTextDocumentClientCapabilitiesItem{
						SnippetSupport: true,
					},
				},
			},
			Window: &protocol.WindowClientCapabilities{
				ShowDocument: &protocol.ShowDocumentClientCapabilities{
					Support: true,
				},
			},
			Workspace: &protocol.WorkspaceClientCapabilities{
				ApplyEdit:     true,
				Configuration: true,
				DidChangeConfiguration: &protocol.DidChangeConfigurationWorkspaceClientCapabilities{
					DynamicRegistration: true,
				},
				DidChangeWatchedFiles: &protocol.DidChangeWatchedFilesWorkspaceClientCapabilities{
					DynamicRegistration: true,
				},
				SemanticTokens: &protocol.SemanticTokensWorkspaceClientCapabilities{
					RefreshSupport: true,
				},
				Symbol: &protocol.WorkspaceSymbolClientCapabilities{
					DynamicRegistration: false,
					SymbolKind: &protocol.SymbolKindCapabilities{
						ValueSet: []protocol.SymbolKind{
							protocol.SymbolKindFile,
							protocol.SymbolKindModule,
							protocol.SymbolKindNamespace,
							protocol.SymbolKindPackage,
							protocol.SymbolKindClass,
							protocol.SymbolKindMethod,
							protocol.SymbolKindProperty,
							protocol.SymbolKindField,
							protocol.SymbolKindConstructor,
							protocol.SymbolKindEnum,
							protocol.SymbolKindInterface,
							protocol.SymbolKindFunction,
							protocol.SymbolKindVariable,
							protocol.SymbolKindConstant,
							protocol.SymbolKindString,
							protocol.SymbolKindNumber,
							protocol.SymbolKindBoolean,
							protocol.SymbolKindArray,
							protocol.SymbolKindObject,
							protocol.SymbolKindKey,
							protocol.SymbolKindNull,
							protocol.SymbolKindEnumMember,
							protocol.SymbolKindStruct,
							protocol.SymbolKindEvent,
							protocol.SymbolKindOperator,
							protocol.SymbolKindTypeParameter,
						},
					},
				},
				WorkspaceEdit: &protocol.WorkspaceClientCapabilitiesWorkspaceEdit{
					ResourceOperations:    []string{"rename", "create", "delete"},
					FailureHandling:       string(protocol.FailureHandlingKindAbort),
					NormalizesLineEndings: true,
					ChangeAnnotationSupport: &protocol.WorkspaceClientCapabilitiesWorkspaceEditChangeAnnotationSupport{
						GroupsOnLabel: true,
					},
				},
				WorkspaceFolders: true,
			},
			General: &protocol.GeneralClientCapabilities{
				RegularExpressions: &protocol.RegularExpressionsClientCapabilities{Engine: "ECMAScript", Version: "2020"},
				Markdown:           &protocol.MarkdownClientCapabilities{Parser: "gfm", Version: "0.29.0"},
			},
		},
		ClientInfo: &protocol.ClientInfo{
			Name:    "Templ LSP Client",
			Version: "1.0.0",
		},
		ProcessID: int32(os.Getpid()),
		RootPath:  rootPath,
		Trace:     protocol.TraceVerbose,
		RootURI:   protocol.DocumentURI("file://" + rootPath),
		WorkspaceFolders: []protocol.WorkspaceFolder{
			{
				Name: "Test Workspace",
				URI:  "file://" + rootPath,
			},
		},
	}
}
