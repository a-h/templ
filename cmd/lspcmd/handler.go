package lspcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type Proxy struct {
	log            *zap.Logger
	gopls          *jsonrpc2.Conn
	client         *jsonrpc2.Conn
	fileCache      *fileCache
	sourceMapCache *sourceMapCache
	toClient       chan toClientRequest
	context        context.Context
}

// NewProxy returns a new proxy to send messages from the client to and from gopls,
// however, init needs to be called before it is usable.
func NewProxy(logger *zap.Logger) (p *Proxy) {
	return &Proxy{
		log:            logger,
		fileCache:      newFileCache(logger),
		sourceMapCache: newSourceMapCache(),
		// Prevent trying to send to the client when message handling is taking place.
		// The proxy can place up to 32 requests onto the toClient buffered channel
		// during handling. They're processed when the clientInUse mutex is released.
		toClient: make(chan toClientRequest, 32),
	}
}

// Init the proxy.
func (p *Proxy) Init(ctx context.Context, client, gopls *jsonrpc2.Conn) {
	p.context = ctx
	p.client = client
	p.gopls = gopls
	go func() {
		for r := range p.toClient {
			r := r
			p.sendToClient(r)
		}
	}()
}

type toClientRequest struct {
	Method string
	Notif  bool
	Params interface{}
}

// sendToClient should not be called directly. Instead, send a message to the non-blocking
// toClient channel.
func (p *Proxy) sendToClient(r toClientRequest) {
	p.log.Info("sendToClient: starting", zap.String("method", r.Method))
	if r.Notif {
		err := p.client.Notify(p.context, r.Method, r.Params)
		if err != nil {
			p.log.Error("sendToClient: error", zap.String("type", "notification"), zap.String("method", r.Method), zap.Error(err))
			return
		}
	} else {
		var result map[string]interface{}
		err := p.client.Call(p.context, r.Method, r.Params, &result)
		if err != nil {
			p.log.Error("sendToClient: error", zap.String("type", "call"), zap.String("method", r.Method), zap.Error(err))
			return
		}
	}
	p.log.Info("sendToClient: success", zap.String("method", r.Method))
}

func (p *Proxy) proxyFromGoplsToClient(ctx context.Context, conn *jsonrpc2.Conn, r *jsonrpc2.Request) {
	p.log.Info("gopls -> client", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	if r.Notif {
		p.log.Info("gopls -> client: notification", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
		err := p.client.Notify(ctx, r.Method, r.Params)
		if err != nil {
			p.log.Error("gopls to client: notification: send error", zap.Error(err))
		}
		p.log.Info("gopls -> client: notification: complete")
	} else {
		p.log.Info("gopls -> client: call", zap.String("method", r.Method), zap.Bool("notif", r.Notif), zap.Any("params", r.Params))
		var result map[string]interface{}
		err := p.client.Call(ctx, r.Method, &r.Params, &result)
		if err != nil {
			p.log.Error("gopls -> client: call: error", zap.Error(err))
		}
		p.log.Info("gopls -> client -> gopls", zap.String("method", r.Method), zap.Any("reply", result))
		err = conn.Reply(ctx, r.ID, result)
		if err != nil {
			p.log.Error("gopls -> client -> gopls: call reply: error", zap.Error(err))
		}
		p.log.Info("gopls -> client: call: complete", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	}
	p.log.Info("gopls -> client: complete", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
}

// Handle implements jsonrpc2.Handler. This function receives from the text editor client, and calls the proxy function
// to determine how to play it back to the client.
func (p *Proxy) Handle(ctx context.Context, conn *jsonrpc2.Conn, r *jsonrpc2.Request) {
	p.log.Info("client -> gopls", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	if r.Notif {
		var err error
		switch r.Method {
		case "textDocument/didOpen":
			err = p.rewriteDidOpenRequest(r)
		case "textDocument/didChange":
			err = p.rewriteDidChangeRequest(ctx, r)
		case "textDocument/didSave":
			err = p.rewriteDidSaveRequest(r)
		case "textDocument/didClose":
			err = p.rewriteDidCloseRequest(r)
		}
		if err != nil {
			p.log.Error("client -> gopls: error rewriting notification", zap.Error(err))
			return
		}
		err = p.gopls.Notify(ctx, r.Method, &r.Params)
		if err != nil {
			p.log.Error("client -> gopls: error proxying notification to gopls", zap.Error(err))
			return
		}
		p.log.Info("client -> gopls: notification complete", zap.String("method", r.Method))
	} else {
		switch r.Method {
		case "textDocument/completion":
			p.proxyCompletion(ctx, conn, r)
			return
		default:
			p.proxyCall(ctx, conn, r)
			return
		}
	}
}

func (p *Proxy) proxyCall(ctx context.Context, conn *jsonrpc2.Conn, r *jsonrpc2.Request) {
	var resp interface{}
	err := p.gopls.Call(ctx, r.Method, &r.Params, &resp)
	p.log.Info("client -> gopls -> client: reply", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	err = conn.Reply(ctx, r.ID, resp)
	if err != nil {
		p.log.Info("client -> gopls -> client: error sending response", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	}
	p.log.Info("client -> gopls -> client: complete", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
}

func (p *Proxy) proxyCompletion(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	// Unmarshal the params.
	var params lsp.CompletionParams
	err := json.Unmarshal(*req.Params, &params)
	if err != nil {
		p.log.Error("proxyCompletion: failed to unmarshal request params", zap.Error(err))
	}
	// Rewrite the request.
	err = p.rewriteCompletionRequest(&params)
	if err != nil {
		p.log.Error("proxyCompletion: error rewriting request", zap.Error(err))
	}
	// Call gopls and get the response.
	var resp lsp.CompletionList
	err = p.gopls.Call(ctx, req.Method, &params, &resp)
	if err != nil {
		p.log.Error("proxyCompletion: client -> gopls: error sending request", zap.Error(err))
	}
	// Rewrite the response.
	err = p.rewriteCompletionResponse(string(params.TextDocument.URI), &resp)
	if err != nil {
		p.log.Error("proxyCompletion: error rewriting response", zap.Error(err))
	}
	// Reply to the client.
	err = conn.Reply(ctx, req.ID, &resp)
	if err != nil {
		p.log.Error("proxyCompletion: error sending response", zap.Error(err))
	}
	p.log.Info("proxyCompletion: client -> gopls -> client: complete", zap.Any("resp", resp))
}

func (p *Proxy) rewriteCompletionResponse(uri string, resp *lsp.CompletionList) (err error) {
	// Get the sourcemap from the cache.
	uri = strings.TrimSuffix(uri, "_templ.go") + ".templ"
	sourceMap, ok := p.sourceMapCache.Get(uri)
	if !ok {
		return fmt.Errorf("unable to complete because the sourcemap doesn't exist in the cache, has the didOpen notification been sent yet?")
	}
	// Rewrite the positions.
	for i := 0; i < len(resp.Items); i++ {
		item := resp.Items[i]
		if item.TextEdit != nil {
			start, _, ok := sourceMap.SourcePositionFromTarget(item.TextEdit.Range.Start.Line+1, item.TextEdit.Range.Start.Character)
			if ok {
				p.log.Info("rewriteCompletionResponse: found new start position", zap.Any("from", item.TextEdit.Range.Start), zap.Any("start", start))
				item.TextEdit.Range.Start.Line = start.Line - 1
				item.TextEdit.Range.Start.Character = start.Col + 1
			}
			end, _, ok := sourceMap.SourcePositionFromTarget(item.TextEdit.Range.End.Line+1, item.TextEdit.Range.End.Character)
			if ok {
				p.log.Info("rewriteCompletionResponse: found new end position", zap.Any("from", item.TextEdit.Range.End), zap.Any("end", end))
				item.TextEdit.Range.End.Line = end.Line - 1
				item.TextEdit.Range.End.Character = end.Col + 1
			}
		}
		resp.Items[i] = item
	}
	return nil
}

func (p *Proxy) rewriteCompletionRequest(params *lsp.CompletionParams) (err error) {
	base, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}
	// Get the sourcemap from the cache.
	sourceMap, ok := p.sourceMapCache.Get(string(params.TextDocument.URI))
	if !ok {
		return fmt.Errorf("unable to complete because the sourcemap doesn't exist in the cache, has the didOpen notification been sent yet?")
	}
	// Map from the source position to target Go position.
	to, mapping, ok := sourceMap.TargetPositionFromSource(params.Position.Line+1, params.Position.Character)
	if ok {
		p.log.Info("rewriteCompletionRequest: found position", zap.Int("fromLine", params.Position.Line+1), zap.Int("fromCol", params.Position.Character), zap.Any("to", to), zap.Any("mapping", mapping))
		params.Position.Line = to.Line - 1
		params.Position.Character = to.Col - 1
		params.TextDocumentPositionParams.Position.Line = params.Position.Line
		params.TextDocumentPositionParams.Position.Character = params.Position.Character
	}
	// Update the URI to make gopls look at the Go code instead.
	params.TextDocument.URI = lsp.DocumentURI(base + (strings.TrimSuffix(fileName, ".templ") + "_templ.go"))
	// Done.
	return err
}

func (p *Proxy) rewriteDidOpenRequest(r *jsonrpc2.Request) (err error) {
	// Unmarshal the params.
	var params lsp.DidOpenTextDocumentParams
	if err = json.Unmarshal(*r.Params, &params); err != nil {
		return err
	}
	base, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}
	// Cache the template doc.
	p.fileCache.Set(string(params.TextDocument.URI), []byte(params.TextDocument.Text))
	// Parse the template.
	template, err := templ.ParseString(params.TextDocument.Text)
	if err != nil {
		return
	}
	// Generate the output code and cache the source map and Go contents to use during completion
	// requests.
	w := new(strings.Builder)
	sm, err := generator.Generate(template, w)
	if err != nil {
		return
	}
	p.sourceMapCache.Set(string(params.TextDocument.URI), sm)
	// Set the Go contents.
	params.TextDocument.Text = w.String()
	// Change the path.
	params.TextDocument.URI = lsp.DocumentURI(base + (strings.TrimSuffix(fileName, ".templ") + "_templ.go"))
	// Marshal the params back.
	jsonMessage, err := json.Marshal(params)
	if err != nil {
		return
	}
	err = r.Params.UnmarshalJSON(jsonMessage)
	// Done.
	return err
}

type applyWorkspaceEditParams struct {
	Label string            `json:"label"`
	Edit  lsp.WorkspaceEdit `json:"edit"`
}

func (p *Proxy) rewriteDidChangeRequest(ctx context.Context, r *jsonrpc2.Request) (err error) {
	// Unmarshal the params.
	var params lsp.DidChangeTextDocumentParams
	if err = json.Unmarshal(*r.Params, &params); err != nil {
		return
	}
	base, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}
	// Apply content changes to the cached template.
	templateText, requestsToClient, err := p.fileCache.Apply(string(params.TextDocument.URI), params.ContentChanges)
	if err != nil {
		return
	}
	// Apply changes to the client.
	for i := 0; i < len(requestsToClient); i++ {
		p.toClient <- requestsToClient[i]
	}
	// Update the Go code.
	template, err := templ.ParseString(string(templateText))
	if err != nil {
		return
	}
	w := new(strings.Builder)
	sm, err := generator.Generate(template, w)
	if err != nil {
		return
	}
	// Cache the sourcemap.
	p.sourceMapCache.Set(string(params.TextDocument.URI), sm)
	// Overwrite all the Go contents.
	params.ContentChanges = []lsp.TextDocumentContentChangeEvent{{
		Range:       nil,
		RangeLength: 0,
		Text:        w.String(),
	}}
	// Change the path.
	params.TextDocument.URI = lsp.DocumentURI(base + (strings.TrimSuffix(fileName, ".templ") + "_templ.go"))
	// Marshal the params back.
	jsonMessage, err := json.Marshal(params)
	if err != nil {
		return
	}
	err = r.Params.UnmarshalJSON(jsonMessage)
	// Done.
	return
}

func (p *Proxy) rewriteDidSaveRequest(r *jsonrpc2.Request) (err error) {
	// Unmarshal the params.
	var params lsp.DidSaveTextDocumentParams
	if err = json.Unmarshal(*r.Params, &params); err != nil {
		return err
	}
	base, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}
	// Update the path.
	params.TextDocument.URI = lsp.DocumentURI(base + (strings.TrimSuffix(fileName, ".templ") + "_templ.go"))
	// Marshal the params back.
	jsonMessage, err := json.Marshal(params)
	if err != nil {
		return
	}
	err = r.Params.UnmarshalJSON(jsonMessage)
	// Done.
	return err
}

func (p *Proxy) rewriteDidCloseRequest(r *jsonrpc2.Request) (err error) {
	// Unmarshal the params.
	var params lsp.DidCloseTextDocumentParams
	if err = json.Unmarshal(*r.Params, &params); err != nil {
		return err
	}
	base, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}
	// Delete the template and sourcemaps from caches.
	p.fileCache.Delete(string(params.TextDocument.URI))
	p.sourceMapCache.Delete(string(params.TextDocument.URI))
	// Get gopls to delete the Go file from its cache.
	params.TextDocument.URI = lsp.DocumentURI(base + (strings.TrimSuffix(fileName, ".templ") + "_templ.go"))
	// Marshal the params back.
	jsonMessage, err := json.Marshal(params)
	if err != nil {
		return
	}
	err = r.Params.UnmarshalJSON(jsonMessage)
	// Done.
	return err
}

// Cache of .templ file URIs to the source map.
func newSourceMapCache() *sourceMapCache {
	return &sourceMapCache{
		m:              new(sync.Mutex),
		uriToSourceMap: make(map[string]*templ.SourceMap),
	}
}

type sourceMapCache struct {
	m              *sync.Mutex
	uriToSourceMap map[string]*templ.SourceMap
}

func (fc *sourceMapCache) Set(uri string, m *templ.SourceMap) {
	fc.m.Lock()
	defer fc.m.Unlock()
	fc.uriToSourceMap[uri] = m
}

func (fc *sourceMapCache) Get(uri string) (m *templ.SourceMap, ok bool) {
	fc.m.Lock()
	defer fc.m.Unlock()
	m, ok = fc.uriToSourceMap[uri]
	return
}

func (fc *sourceMapCache) Delete(uri string) {
	fc.m.Lock()
	defer fc.m.Unlock()
	delete(fc.uriToSourceMap, uri)
}

// Cache of files to their contents.
func newFileCache(logger *zap.Logger) *fileCache {
	return &fileCache{
		m:             new(sync.Mutex),
		uriToContents: make(map[string][]byte),
		log:           logger,
	}
}

type fileCache struct {
	m             *sync.Mutex
	uriToContents map[string][]byte
	log           *zap.Logger
}

func (fc *fileCache) Set(uri string, contents []byte) {
	fc.m.Lock()
	defer fc.m.Unlock()
	fc.uriToContents[uri] = contents
}

func (fc *fileCache) Get(uri string) (contents []byte, ok bool) {
	fc.m.Lock()
	defer fc.m.Unlock()
	contents, ok = fc.uriToContents[uri]
	return
}

func (fc *fileCache) Delete(uri string) {
	fc.m.Lock()
	defer fc.m.Unlock()
	delete(fc.uriToContents, uri)
}

func (fc *fileCache) Apply(uri string, changes []lsp.TextDocumentContentChangeEvent) (updated []byte, requestsToClient []toClientRequest, err error) {
	fc.m.Lock()
	defer fc.m.Unlock()
	contents, ok := fc.uriToContents[uri]
	if !ok {
		err = fmt.Errorf("document not found")
		return
	}
	updated, insertCloses, err := applyContentChanges(fc.log, lsp.DocumentURI(uri), contents, changes)
	if err != nil {
		return
	}
	requestsToClient = make([]toClientRequest, len(insertCloses))
	for i := 0; i < len(insertCloses); i++ {
		requestsToClient[i] = createWorkspaceApplyEditInsert(uri, " %}\n", insertCloses[i], insertAfter)
	}
	fc.uriToContents[uri] = updated
	return
}

type insertPosition int

const (
	insertBefore insertPosition = iota
	insertAfter
)

func createWorkspaceApplyEditInsert(documentURI, text string, at lsp.Position, position insertPosition) toClientRequest {
	textRange := lsp.Range{
		Start: at,
		End:   at,
	}
	if position == insertAfter {
		textRange.Start.Character++
		textRange.End.Character += len(text) + 1
	}
	return toClientRequest{
		Method: "workspace/applyEdit",
		Notif:  false,
		Params: applyWorkspaceEditParams{
			Label: "templ close tag",
			Edit: lsp.WorkspaceEdit{
				Changes: map[string][]lsp.TextEdit{
					documentURI: {
						{
							Range:   textRange,
							NewText: text,
						},
					},
				},
			},
		},
	}
}

// Check the last couple of bytes for "{%= " and "{% ".
func shouldInsertCloseTag(prefix, change string) (shouldInsert bool) {
	last := 4
	if last > len(prefix) {
		last = len(prefix)
	}
	upToCaret := prefix[len(prefix)-last:] + change
	return strings.HasSuffix(upToCaret, "{% ") || strings.HasSuffix(upToCaret, "{%= ")
}

// Contents below adapted from https://github.com/sourcegraph/go-langserver/blob/4b49d01c8a692968252730d45980091dcec7752e/langserver/fs.go#L141
// It implements the ability to react to changes on document edits.
// MIT licensed.
// applyContentChanges updates `contents` based on `changes`
func applyContentChanges(log *zap.Logger, uri lsp.DocumentURI, contents []byte, changes []lsp.TextDocumentContentChangeEvent) (c []byte, insertCloses []lsp.Position, err error) {
	for _, change := range changes {
		if change.Range == nil && change.RangeLength == 0 {
			contents = []byte(change.Text) // new full content
			continue
		}
		start, ok, why := offsetForPosition(contents, change.Range.Start)
		if !ok {
			return nil, insertCloses, fmt.Errorf("received textDocument/didChange for invalid position %q on %q: %s", change.Range.Start, uri, why)
		}
		var end int
		if change.RangeLength != 0 {
			end = start + int(change.RangeLength)
		} else {
			// RangeLength not specified, work it out from Range.End
			end, ok, why = offsetForPosition(contents, change.Range.End)
			if !ok {
				return nil, insertCloses, fmt.Errorf("received textDocument/didChange for invalid position %q on %q: %s", change.Range.Start, uri, why)
			}
		}
		if start < 0 || end > len(contents) || end < start {
			return nil, insertCloses, fmt.Errorf("received textDocument/didChange for out of range position %q on %q", change.Range, uri)
		}
		// Custom code to check for autocomplete.
		if len(change.Text) > 0 && shouldInsertCloseTag(string(contents[:start]), change.Text) {
			end := lsp.Position{
				Line:      change.Range.End.Line,
				Character: change.Range.End.Character,
			}
			insertCloses = append(insertCloses, end)
		}
		// End of custom code.
		// Try avoid doing too many allocations, so use bytes.Buffer
		b := &bytes.Buffer{}
		b.Grow(start + len(change.Text) + len(contents) - end)
		b.Write(contents[:start])
		b.WriteString(change.Text)
		b.Write(contents[end:])
		contents = b.Bytes()
	}
	return contents, insertCloses, nil
}

func offsetForPosition(contents []byte, p lsp.Position) (offset int, valid bool, whyInvalid string) {
	line := 0
	col := 0
	// TODO(sqs): count chars, not bytes, per LSP. does that mean we
	// need to maintain 2 separate counters since we still need to
	// return the offset as bytes?
	for _, b := range contents {
		if line == p.Line && col == p.Character {
			return offset, true, ""
		}
		if (line == p.Line && col > p.Character) || line > p.Line {
			return 0, false, fmt.Sprintf("character %d (zero-based) is beyond line %d boundary (zero-based)", p.Character, p.Line)
		}
		offset++
		if b == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	if line == p.Line && col == p.Character {
		return offset, true, ""
	}
	if line == 0 {
		return 0, false, fmt.Sprintf("character %d (zero-based) is beyond first line boundary", p.Character)
	}
	return 0, false, fmt.Sprintf("file only has %d lines", line+1)
}
