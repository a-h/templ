package lsp

import (
	"context"

	"github.com/a-h/templ/cmd/lsp/pls"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type Proxy struct {
	log    *zap.Logger
	gopls  *jsonrpc2.Conn
	client *jsonrpc2.Conn
}

func NewProxy(logger *zap.Logger) (h *Proxy, err error) {
	h = &Proxy{
		log: logger,
	}
	h.gopls, err = pls.NewGopls(logger, h.proxyFromGoplsToClient)
	return h, err
}

func (h *Proxy) proxyFromGoplsToClient(ctx context.Context, conn *jsonrpc2.Conn, r *jsonrpc2.Request) {
	h.log.Info("gopls -> client", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	if r.Notif {
		h.log.Info("gopls -> client: notification", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
		err := h.client.Notify(ctx, r.Method, r.Params)
		if err != nil {
			h.log.Error("gopls to client: notification: send error", zap.Error(err))
		}
		h.log.Info("gopls -> client: notification: complete")
	} else {
		h.log.Info("gopls -> client: call", zap.String("method", r.Method), zap.Bool("notif", r.Notif), zap.Any("params", r.Params))
		var result map[string]interface{}
		err := h.client.Call(ctx, r.Method, &r.Params, &result)
		if err != nil {
			h.log.Error("gopls -> client: call: error", zap.Error(err))
		}
		h.log.Info("gopls -> client -> gopls", zap.String("method", r.Method), zap.Any("reply", result))
		// Reply to gopls.
		err = conn.Reply(ctx, r.ID, result)
		if err != nil {
			h.log.Error("gopls -> client -> gopls: call reply: error", zap.Error(err))
		}
		h.log.Info("gopls -> client: call: complete", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	}
	h.log.Info("gopls -> client: complete", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
}

// Handle implements jsonrpc2.Handler. This function receives from the text editor client, and calls the proxy function
// to determine how to play it back to the client.
func (h *Proxy) Handle(ctx context.Context, conn *jsonrpc2.Conn, r *jsonrpc2.Request) {
	h.log.Info("client -> gopls", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	if r.Notif {
		err := h.gopls.Notify(ctx, r.Method, &r.Params)
		if err != nil {
			h.log.Error("client -> gopls: error proxying to gopls", zap.Error(err))
			return
		}
		h.log.Info("client -> gopls: notification: complete", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	} else {
		var resp interface{}
		err := h.gopls.Call(ctx, r.Method, &r.Params, &resp)
		h.log.Info("client -> gopls -> client: reply", zap.String("method", r.Method), zap.Bool("notif", r.Notif), zap.Any("resp", resp))
		err = conn.Reply(ctx, r.ID, resp)
		if err != nil {
			h.log.Info("client -> gopls -> client: error sending response", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
		}
		h.log.Info("client -> gopls -> client: complete", zap.String("method", r.Method), zap.Bool("notif", r.Notif))
	}
}

//func (h *Proxy) proxy(ctx context.Context, r *jsonrpc2.Request) (result interface{}, err error) {
////TODO: Prevent any uncaught panics from taking the entire server down.
//switch r.Method {
//case "textDocument/didOpen":
////TODO: change the file name and file position
//if r.Params == nil {
//return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
//}
//var params lsp.DidOpenTextDocumentParams
//if err := json.Unmarshal(*r.Params, &params); err != nil {
//return nil, err
//}
//return h.handleDidOpen(ctx, r, params)

//case "textDocument/didClose":
////TODO: change the file name and file position
//if r.Params == nil {
//return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
//}
//var params lsp.DidCloseTextDocumentParams
//if err := json.Unmarshal(*r.Params, &params); err != nil {
//return nil, err
//}
//return h.handleDidClose(ctx, r, params)

//case "textDocument/hover":
////TODO: change the file name and file position
//if r.Params == nil {
//return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
//}
//var params lsp.TextDocumentPositionParams
//if err := json.Unmarshal(*r.Params, &params); err != nil {
//return nil, err
//}
//return h.handleTextDocumentHover(ctx, r, params)

//case "textDocument/completion":
////TODO: mutate the cols/rows
//if r.Params == nil {
//return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
//}
//var params lsp.CompletionParams
//if err := json.Unmarshal(*r.Params, &params); err != nil {
//return nil, err
//}
//return h.handleTextDocumentCompletion(ctx, r, params)

//case "textDocument/signatureHelp":
////TODO: change the file name and file position
//if r.Params == nil {
//return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
//}
//var params lsp.TextDocumentPositionParams
//if err := json.Unmarshal(*r.Params, &params); err != nil {
//return nil, err
//}
//return h.handleTextDocumentSignatureHelp(ctx, r, params)
//}

//// Proxy to gopls without any changes by default.
//if r.Notif {
//err = h.gopls.Notify(ctx, r.Method, &r.Params)
//} else {
//err = h.gopls.Call(ctx, r.Method, &r.Params, &result)
//}
//return
//}

//func (h Proxy) handleDidOpen(ctx context.Context, req *jsonrpc2.Request, params lsp.DidOpenTextDocumentParams) (result *lsp.TextDocumentSyncOptions, err error) {
//// Get the client to always send the whole document?
//err = h.gopls.Notify(ctx, req.Method, params)
//return
//}

//func (h Proxy) handleDidClose(ctx context.Context, req *jsonrpc2.Request, params lsp.DidCloseTextDocumentParams) (*lsp.TextDocumentSyncOptions, error) {
//// Get the client to always send the whole document.
//return &lsp.TextDocumentSyncOptions{
//OpenClose: true,
//Change:    lsp.TDSKFull,
//}, nil
//}

//func (h Proxy) handleTextDocumentSignatureHelp(ctx context.Context, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.SignatureHelp, error) {
//return &lsp.SignatureHelp{
//Signatures: []lsp.SignatureInformation{
//{
//Label:         "signature",
//Documentation: "doc",
//Parameters: []lsp.ParameterInformation{
//{
//Label:         "param1",
//Documentation: "param1Doc",
//},
//},
//},
//},
//}, nil
//}

//func (h Proxy) handleTextDocumentHover(ctx context.Context, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
//return &lsp.Hover{
//Contents: []lsp.MarkedString{
//{
//Language: "go",
//Value:    "Hover handler",
//},
//},
//Range: &lsp.Range{
//Start: lsp.Position{Line: params.Position.Line, Character: params.Position.Character},
//End:   lsp.Position{Line: params.Position.Line, Character: params.Position.Character},
//},
//}, nil
//}

//func (h Proxy) handleTextDocumentCompletion(ctx context.Context, req *jsonrpc2.Request, params lsp.CompletionParams) (*lsp.CompletionList, error) {
//citems := []lsp.CompletionItem{
//{
//Kind:             lsp.CIKConstant,
//Label:            "A constant",
//Detail:           "Its type",
//InsertTextFormat: lsp.ITFPlainText,
//InsertText:       "New Text",
//TextEdit: &lsp.TextEdit{
//Range: lsp.Range{
//Start: lsp.Position{Line: params.Position.Line, Character: params.Position.Character - len("New Text")},
//End:   lsp.Position{Line: params.Position.Line, Character: params.Position.Character},
//},
//NewText: "New Text",
//},
//},
//}
//return &lsp.CompletionList{
//IsIncomplete: false,
//Items:        citems,
//}, nil
//}
