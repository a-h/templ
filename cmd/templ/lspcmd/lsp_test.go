package lspcmd

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/a-h/protocol"
	"github.com/a-h/templ/cmd/templ/generatecmd/modcheck"
	"go.lsp.dev/jsonrpc2"
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

func replaceInFile(name, src, tgt string) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	updated := strings.Replace(string(data), src, tgt, -1)
	return os.WriteFile(name, []byte(updated), 0660)
}

func TestLSP(t *testing.T) {
	if testing.Short() {
		return
	}
	ctx := context.Background()

	_, server, teardown, err := Setup(ctx)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)

	initializeResult, err := server.Initialize(ctx, &protocol.InitializeParams{})
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	if initializeResult.Capabilities.HoverProvider == nil {
		t.Errorf("expected hover capabilities, got nil")
	}
}

type TestClient struct {
}

func (tc TestClient) Progress(ctx context.Context, params *protocol.ProgressParams) (err error) {
	return nil
}

func (tc TestClient) WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) (err error) {
	return nil
}

func (tc TestClient) LogMessage(ctx context.Context, params *protocol.LogMessageParams) (err error) {
	return nil
}

func (tc TestClient) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) (err error) {
	return nil
}

func (tc TestClient) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) (err error) {
	return nil
}

func (tc TestClient) ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (result *protocol.MessageActionItem, err error) {
	return nil, nil
}

func (tc TestClient) Telemetry(ctx context.Context, params interface{}) (err error) {
	return nil
}

func (tc TestClient) RegisterCapability(ctx context.Context, params *protocol.RegistrationParams) (err error) {
	return nil
}

func (tc TestClient) UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) (err error) {
	return nil
}

func (tc TestClient) ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (result *protocol.ApplyWorkspaceEditResponse, err error) {
	return nil, nil
}

func (tc TestClient) Configuration(ctx context.Context, params *protocol.ConfigurationParams) (result []interface{}, err error) {
	return nil, nil
}

func (tc TestClient) WorkspaceFolders(ctx context.Context) (result []protocol.WorkspaceFolder, err error) {
	return nil, nil
}

func Setup(ctx context.Context) (client protocol.Client, server protocol.Server, teardown func(t *testing.T), err error) {
	wd, err := os.Getwd()
	if err != nil {
		return client, server, teardown, fmt.Errorf("could not find working dir: %w", err)
	}
	moduleRoot, err := modcheck.WalkUp(wd)
	if err != nil {
		return client, server, teardown, fmt.Errorf("could not find local templ go.mod file: %v", err)
	}

	appDir, err := createTestProject(moduleRoot)
	if err != nil {
		return client, server, teardown, fmt.Errorf("failed to create test project: %v", err)
	}

	var wg sync.WaitGroup
	var cmdErr error

	log, _ := zap.NewProduction()

	// Copy from the LSP to the CLient, and vice versa.
	toLSP := new(bytes.Buffer)
	fromLSP := new(bytes.Buffer)
	serverStream := jsonrpc2.NewStream(newStdRwc(log, "templStream", fromLSP, toLSP))

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("Running")
		cmdErr = run(ctx, log, serverStream, Arguments{})
		log.Info("Stopped")
	}()

	_, _, templServer := protocol.NewClient(context.Background(), client, serverStream, log)

	_, err = templServer.Initialize(context.Background(), &protocol.InitializeParams{})
	if err != nil {
		log.Error("Failed to init", zap.Error(err))
	}
	log.Info("Initialized...")

	// Wait for exit.
	teardown = func(t *testing.T) {
		wg.Wait()
		if cmdErr != nil {
			t.Errorf("failed to run lsp cmd: %v", err)
		}

		if err = os.RemoveAll(appDir); err != nil {
			t.Errorf("failed to remove test dir %q: %v", appDir, err)
		}
	}
	return client, server, teardown, err
}
