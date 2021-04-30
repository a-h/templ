package lsp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type Handler struct {
	jsonrpc2.Handler
	log *zap.Logger
}

func NewHandler(logger *zap.Logger) Handler {
	return Handler{
		log: logger,
	}
}

// Handle implements jsonrpc2.Handler
func (h Handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	h.log.Info("request", zap.Any("req", req))
	resp, err := h.internal(ctx, conn, req)
	if err != nil {
		h.log.Error("error creating response", zap.Error(err))
		return
	}
	h.log.Info("response", zap.Any("resp", resp))
	err = conn.Reply(ctx, req.ID, resp)
	if err != nil {
		h.log.Error("error sending response", zap.Error(err))
	}
}

func (h Handler) internal(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	//TODO: Prevent any uncaught panics from taking the entire server down.
	switch req.Method {
	case "initialize":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		kind := lsp.TDSKIncremental
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
					Kind: &kind,
				},
				CompletionProvider:     &lsp.CompletionOptions{ResolveProvider: true, TriggerCharacters: []string{"(", "."}},
				DefinitionProvider:     true,
				TypeDefinitionProvider: true,
				DocumentSymbolProvider: true,
				HoverProvider:          true,
				ReferencesProvider:     true,
				ImplementationProvider: true,
				SignatureHelpProvider:  &lsp.SignatureHelpOptions{TriggerCharacters: []string{"(", ","}},
			},
		}, nil

	case "initialized":
		// A notification that the client is ready to receive requests. Ignore
		return nil, nil

	case "shutdown":
		return nil, nil

	case "exit":
		conn.Close()
		return nil, nil

	case "$/cancelRequest":
		// notification, don't send back results/errors
		if req.Params == nil {
			return nil, nil
		}
		var params lsp.CancelParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, nil
		}
		return nil, nil

	case "textDocument/didOpen":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.DidOpenTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

	case "textDocument/didClose":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.DidCloseTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleDidClose(ctx, conn, req, params)

	case "textDocument/hover":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleTextDocumentHover(ctx, conn, req, params)

	case "textDocument/completion":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.CompletionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleTextDocumentCompletion(ctx, conn, req, params)

	case "textDocument/signatureHelp":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleTextDocumentSignatureHelp(ctx, conn, req, params)
	}
	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (h Handler) handleDidOpen(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.DidOpenTextDocumentParams) (*lsp.TextDocumentSyncOptions, error) {
	// Get the client to always send the whole document.
	return &lsp.TextDocumentSyncOptions{
		OpenClose: true,
		Change:    lsp.TDSKFull,
	}, nil
}

func (h Handler) handleDidClose(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.DidCloseTextDocumentParams) (*lsp.TextDocumentSyncOptions, error) {
	// Get the client to always send the whole document.
	return &lsp.TextDocumentSyncOptions{
		OpenClose: true,
		Change:    lsp.TDSKFull,
	}, nil
}

func (h Handler) handleTextDocumentSignatureHelp(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.SignatureHelp, error) {
	return &lsp.SignatureHelp{
		Signatures: []lsp.SignatureInformation{
			{
				Label:         "signature",
				Documentation: "doc",
				Parameters: []lsp.ParameterInformation{
					{
						Label:         "param1",
						Documentation: "param1Doc",
					},
				},
			},
		},
	}, nil
}

func (h Handler) handleTextDocumentHover(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	return &lsp.Hover{
		Contents: []lsp.MarkedString{
			{
				Language: "go",
				Value:    "Hover handler",
			},
		},
		Range: &lsp.Range{
			Start: lsp.Position{Line: params.Position.Line, Character: params.Position.Character},
			End:   lsp.Position{Line: params.Position.Line, Character: params.Position.Character},
		},
	}, nil
}

func (h Handler) handleTextDocumentCompletion(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.CompletionParams) (*lsp.CompletionList, error) {
	citems := []lsp.CompletionItem{
		{
			Kind:             lsp.CIKConstant,
			Label:            "A constant",
			Detail:           "Its type",
			InsertTextFormat: lsp.ITFPlainText,
			InsertText:       "New Text",
			TextEdit: &lsp.TextEdit{
				Range: lsp.Range{
					Start: lsp.Position{Line: params.Position.Line, Character: params.Position.Character - len("New Text")},
					End:   lsp.Position{Line: params.Position.Line, Character: params.Position.Character},
				},
				NewText: "New Text",
			},
		},
	}
	return &lsp.CompletionList{
		IsIncomplete: false,
		Items:        citems,
	}, nil
}
