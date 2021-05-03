package lsp

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
	log       *zap.Logger
	gopls     *jsonrpc2.Conn
	client    *jsonrpc2.Conn
	fileCache *FileCache
}

// NewProxy returns a new proxy to send messages from the client to and from gopls,
// however, init needs to be called before it is usable.
func NewProxy(logger *zap.Logger) (p *Proxy) {
	return &Proxy{
		log:       logger,
		fileCache: NewFileCache(),
	}
}

// Init the proxy.
func (p *Proxy) Init(client, gopls *jsonrpc2.Conn) {
	p.client = client
	p.gopls = gopls
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
		// Reply to gopls.
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
		//TODO: Log potential errors here.
		switch r.Method {
		case "textDocument/didOpen":
			p.rewriteDidOpenRequest(r)
		case "textDocument/didChange":
			p.rewriteDidChangeRequest(r)
		case "textDocument/didSave":
			//TODO: Probably just needs the filename adjusting.
		case "textDocument/didClose":
			p.rewriteDidCloseRequest(r)
		}
		err := p.gopls.Notify(ctx, r.Method, &r.Params)
		if err != nil {
			p.log.Error("client -> gopls: error proxying to gopls", zap.Error(err))
			return
		}
		p.log.Info("client -> gopls: notification: complete", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	} else {
		//TODO: Log potential errors here.
		switch r.Method {
		case "textDocument/completion":
			p.rewriteCompletionRequest(r)
		}
		var resp interface{}
		err := p.gopls.Call(ctx, r.Method, &r.Params, &resp)
		p.log.Info("client -> gopls -> client: reply", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
		err = conn.Reply(ctx, r.ID, resp)
		if err != nil {
			p.log.Info("client -> gopls -> client: error sending response", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
		}
		p.log.Info("client -> gopls -> client: complete", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	}
}

func (p *Proxy) rewriteCompletionRequest(r *jsonrpc2.Request) (err error) {
	// Unmarshal the params.
	var params lsp.CompletionParams
	if err = json.Unmarshal(*r.Params, &params); err != nil {
		return err
	}
	base, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}
	// Parse the template from the cache.
	templateText, ok := p.fileCache.Get(string(params.TextDocument.URI))
	if !ok {
		return fmt.Errorf("unable to complete because the document doesn't exist in the cache, has the didOpen notification been sent yet?")
	}
	template, err := templ.ParseString(string(templateText))
	if err != nil {
		p.log.Warn("rewriteCompletionRequest: failed to parse document", zap.Error(err))
		return
	}
	w := new(strings.Builder)
	sourceMap, err := generator.Generate(template, w)
	if err != nil {
		p.log.Warn("rewriteCompletionRequest: failed to generate Go code", zap.Error(err))
		return
	}
	// Map from the source position to target go position.
	from := templ.Position{
		Line: params.Position.Line + 1,
		Col:  params.Position.Character,
	}
	to, mapping, ok := sourceMap.TargetPositionFromSource(from)
	if ok {
		sourceLine := getLines(from.Line, from.Line, string(templateText))
		targetLine := getLines(to.Line, to.Line, w.String())
		p.log.Info("rewriteCompletionRequest: found position", zap.Any("from", from), zap.Any("to", to), zap.Any("mapping", mapping), zap.String("sourceLine", sourceLine), zap.String("targetLine", targetLine))
		params.Position.Line = to.Line - 1
		params.Position.Character = to.Col - 1
		params.TextDocumentPositionParams.Position.Line = params.Position.Line
		params.TextDocumentPositionParams.Position.Character = params.Position.Character
	}

	// Update the URI to make gopls look at the Go code instead.
	params.TextDocument.URI = lsp.DocumentURI(base + (strings.TrimSuffix(fileName, ".templ") + "_templ.go"))

	// Marshal the params back.
	jsonMessage, err := json.Marshal(params)
	if err != nil {
		p.log.Warn("rewriteCompletionRequest: failed to marshal param", zap.Error(err))
		return
	}
	err = r.Params.UnmarshalJSON(jsonMessage)

	// Done.
	return err
}

func getLines(from, to int, text string) string {
	lines := strings.Split(text, "\n")
	return strings.Join(lines[from-1:to], "\n")
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
	w := new(strings.Builder)
	_, err = generator.Generate(template, w)
	if err != nil {
		return
	}
	// Set the go contents.
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

func (p *Proxy) rewriteDidChangeRequest(r *jsonrpc2.Request) (err error) {
	// Unmarshal the params.
	var params lsp.DidChangeTextDocumentParams
	if err = json.Unmarshal(*r.Params, &params); err != nil {
		return err
	}
	base, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}

	// Apply content changes to the cached template.
	templateText, err := p.fileCache.Apply(string(params.TextDocument.URI), params.ContentChanges)
	if err != nil {
		return
	}

	// Update the go code.
	template, err := templ.ParseString(string(templateText))
	if err != nil {
		return
	}
	w := new(strings.Builder)
	_, err = generator.Generate(template, w)
	if err != nil {
		return
	}
	// Overwrite all the go contents.
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
	// Delete the template from the cache.
	p.fileCache.Delete(string(params.TextDocument.URI))

	// Get gopls to delete the go file from its cache.
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

func NewFileCache() *FileCache {
	return &FileCache{
		m:             new(sync.Mutex),
		uriToContents: make(map[string][]byte),
	}
}

type FileCache struct {
	m             *sync.Mutex
	uriToContents map[string][]byte
}

func (fc *FileCache) Set(uri string, contents []byte) {
	fc.m.Lock()
	defer fc.m.Unlock()
	fc.uriToContents[uri] = contents
}

func (fc *FileCache) Get(uri string) (contents []byte, ok bool) {
	fc.m.Lock()
	defer fc.m.Unlock()
	contents, ok = fc.uriToContents[uri]
	return
}

func (fc *FileCache) Delete(uri string) {
	fc.m.Lock()
	defer fc.m.Unlock()
	delete(fc.uriToContents, uri)
}

func (fc *FileCache) Apply(uri string, changes []lsp.TextDocumentContentChangeEvent) (updated []byte, err error) {
	fc.m.Lock()
	defer fc.m.Unlock()
	contents, ok := fc.uriToContents[uri]
	if !ok {
		err = fmt.Errorf("document not found")
		return
	}
	updated, err = applyContentChanges(lsp.DocumentURI(uri), contents, changes)
	if err != nil {
		return
	}
	fc.uriToContents[uri] = updated
	return
}

// Contents below adapted from https://github.com/sourcegraph/go-langserver/blob/4b49d01c8a692968252730d45980091dcec7752e/langserver/fs.go#L141
// It implements the ability to react to changes on document edits.
// MIT licensed.

// applyContentChanges updates `contents` based on `changes`
func applyContentChanges(uri lsp.DocumentURI, contents []byte, changes []lsp.TextDocumentContentChangeEvent) ([]byte, error) {
	for _, change := range changes {
		if change.Range == nil && change.RangeLength == 0 {
			contents = []byte(change.Text) // new full content
			continue
		}
		start, ok, why := offsetForPosition(contents, change.Range.Start)
		if !ok {
			return nil, fmt.Errorf("received textDocument/didChange for invalid position %q on %q: %s", change.Range.Start, uri, why)
		}
		var end int
		if change.RangeLength != 0 {
			end = start + int(change.RangeLength)
		} else {
			// RangeLength not specified, work it out from Range.End
			end, ok, why = offsetForPosition(contents, change.Range.End)
			if !ok {
				return nil, fmt.Errorf("received textDocument/didChange for invalid position %q on %q: %s", change.Range.Start, uri, why)
			}
		}
		if start < 0 || end > len(contents) || end < start {
			return nil, fmt.Errorf("received textDocument/didChange for out of range position %q on %q", change.Range, uri)
		}
		// Try avoid doing too many allocations, so use bytes.Buffer
		b := &bytes.Buffer{}
		b.Grow(start + len(change.Text) + len(contents) - end)
		b.Write(contents[:start])
		b.WriteString(change.Text)
		b.Write(contents[end:])
		contents = b.Bytes()
	}
	return contents, nil
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
