package proxy

import (
	"context"
	"fmt"
	"strings"

	lsp "github.com/a-h/protocol"
	"go.uber.org/zap"
)

// Client is responsible for rewriting messages that are
// originated from gopls, and are sent to the client.
//
// Since `gopls` is working on Go files, and this is the `templ` LSP,
// the job of this code is to rewrite incoming requests to adjust the
// file name from `*_templ.go` to `*.templ`, and to remap the char
// positions where required.
type Client struct {
	Log             *zap.Logger
	Target          lsp.Client
	SourceMapCache  *SourceMapCache
	DiagnosticCache *DiagnosticCache
}

func NewClient(log *zap.Logger, cache *SourceMapCache, diagnosticCache *DiagnosticCache) (c *Client, init func(lsp.Client)) {
	c = &Client{
		Log:             log,
		SourceMapCache:  cache,
		DiagnosticCache: diagnosticCache,
	}
	return c, func(target lsp.Client) {
		c.Target = target
	}
}

func (p Client) Progress(ctx context.Context, params *lsp.ProgressParams) (err error) {
	p.Log.Info("client <- server: Progress")
	return p.Target.Progress(ctx, params)
}
func (p Client) WorkDoneProgressCreate(ctx context.Context, params *lsp.WorkDoneProgressCreateParams) (err error) {
	p.Log.Info("client <- server: WorkDoneProgressCreate")
	return p.Target.WorkDoneProgressCreate(ctx, params)
}

func (p Client) LogMessage(ctx context.Context, params *lsp.LogMessageParams) (err error) {
	p.Log.Info("client <- server: LogMessage", zap.String("message", params.Message))
	return p.Target.LogMessage(ctx, params)
}

func (p Client) PublishDiagnostics(ctx context.Context, params *lsp.PublishDiagnosticsParams) (err error) {
	p.Log.Info("client <- server: PublishDiagnostics")
	if strings.HasSuffix(string(params.URI), "go.mod") {
		p.Log.Debug("client <- server: PublishDiagnostics: skipping go.mod diagnostics")
		return nil
	}
	// Log diagnostics.
	for i, diagnostic := range params.Diagnostics {
		p.Log.Info(fmt.Sprintf("client <- server: PublishDiagnostics: [%d]", i), zap.Any("diagnostic", diagnostic))
	}
	// Get the sourcemap from the cache.
	uri := strings.TrimSuffix(string(params.URI), "_templ.go") + ".templ"
	sourceMap, ok := p.SourceMapCache.Get(uri)
	if !ok {
		p.Log.Error("unable to complete because the sourcemap for the URI doesn't exist in the cache", zap.String("uri", uri))
		return fmt.Errorf("unable to complete because the sourcemap for %q doesn't exist in the cache, has the didOpen notification been sent yet?", uri)
	}
	params.URI = lsp.DocumentURI(uri)
	// Rewrite the positions.
	for i := 0; i < len(params.Diagnostics); i++ {
		item := params.Diagnostics[i]
		start, ok := sourceMap.SourcePositionFromTarget(item.Range.Start.Line, item.Range.Start.Character)
		if !ok {
			continue
		}
		if item.Range.Start.Line == item.Range.End.Line {
			length := item.Range.End.Character - item.Range.Start.Character
			item.Range.Start.Line = start.Line
			item.Range.Start.Character = start.Col
			item.Range.End.Line = start.Line
			item.Range.End.Character = start.Col + length
			params.Diagnostics[i] = item
			p.Log.Info(fmt.Sprintf("diagnostic [%d] rewritten", i), zap.Any("diagnostic", item))
			continue
		}
		end, ok := sourceMap.SourcePositionFromTarget(item.Range.End.Line, item.Range.End.Character)
		if !ok {
			continue
		}
		item.Range.Start.Line = start.Line
		item.Range.Start.Character = start.Col
		item.Range.End.Line = end.Line
		item.Range.End.Character = end.Col
		params.Diagnostics[i] = item
		p.Log.Info(fmt.Sprintf("diagnostic [%d] rewritten", i), zap.Any("diagnostic", item))
	}
	params.Diagnostics = p.DiagnosticCache.AddTemplDiagnostics(uri, params.Diagnostics)
	err = p.Target.PublishDiagnostics(ctx, params)
	return err
}

func (p Client) ShowMessage(ctx context.Context, params *lsp.ShowMessageParams) (err error) {
	p.Log.Info("client <- server: ShowMessage", zap.String("message", params.Message))
	if strings.HasPrefix(params.Message, "Do not edit this file!") {
		return
	}
	return p.Target.ShowMessage(ctx, params)
}

func (p Client) ShowMessageRequest(ctx context.Context, params *lsp.ShowMessageRequestParams) (result *lsp.MessageActionItem, err error) {
	p.Log.Info("client <- server: ShowMessageRequest", zap.String("message", params.Message))
	return p.Target.ShowMessageRequest(ctx, params)
}

func (p Client) Telemetry(ctx context.Context, params interface{}) (err error) {
	p.Log.Info("client <- server: Telemetry")
	return p.Target.Telemetry(ctx, params)
}

func (p Client) RegisterCapability(ctx context.Context, params *lsp.RegistrationParams) (err error) {
	p.Log.Info("client <- server: RegisterCapability")
	return p.Target.RegisterCapability(ctx, params)
}

func (p Client) UnregisterCapability(ctx context.Context, params *lsp.UnregistrationParams) (err error) {
	p.Log.Info("client <- server: UnregisterCapability")
	return p.Target.UnregisterCapability(ctx, params)
}

func (p Client) ApplyEdit(ctx context.Context, params *lsp.ApplyWorkspaceEditParams) (result *lsp.ApplyWorkspaceEditResponse, err error) {
	p.Log.Info("client <- server: ApplyEdit")
	return p.Target.ApplyEdit(ctx, params)
}

func (p Client) Configuration(ctx context.Context, params *lsp.ConfigurationParams) (result []interface{}, err error) {
	p.Log.Info("client <- server: Configuration")
	return p.Target.Configuration(ctx, params)
}

func (p Client) WorkspaceFolders(ctx context.Context) (result []lsp.WorkspaceFolder, err error) {
	p.Log.Info("client <- server: WorkspaceFolders")
	return p.Target.WorkspaceFolders(ctx)
}
