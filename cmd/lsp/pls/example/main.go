package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/sourcegraph/jsonrpc2"
)

func main() {
	fmt.Println(runRPC())
}

type RPC struct {
	Method  string
	Payload string
	Notify  bool
}

var rpcInitialize = RPC{
	Method:  "initialize",
	Payload: `{"processId":71156,"rootPath":"/Users/adrian/github.com/a-h/templ","rootUri":"file:///Users/adrian/github.com/a-h/templ","capabilities":{"workspace":{"applyEdit":true,"workspaceEdit":{"documentChanges":true,"resourceOperations":["create","rename","delete"],"failureHandling":"textOnlyTransactional"},"didChangeConfiguration":{"dynamicRegistration":true},"didChangeWatchedFiles":{"dynamicRegistration":true},"symbol":{"dynamicRegistration":true,"symbolKind":{"valueSet":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26]},"tagSupport":{"valueSet":[1]}},"executeCommand":{"dynamicRegistration":true},"configuration":true,"workspaceFolders":true},"textDocument":{"publishDiagnostics":{"relatedInformation":true,"versionSupport":false,"tagSupport":{"valueSet":[1,2]}},"synchronization":{"dynamicRegistration":true,"willSave":true,"willSaveWaitUntil":true,"didSave":true},"completion":{"dynamicRegistration":true,"contextSupport":true,"completionItem":{"snippetSupport":true,"commitCharactersSupport":true,"documentationFormat":["markdown","plaintext"],"deprecatedSupport":true,"preselectSupport":true,"tagSupport":{"valueSet":[1]}},"completionItemKind":{"valueSet":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25]}},"hover":{"dynamicRegistration":true,"contentFormat":["markdown","plaintext"]},"signatureHelp":{"dynamicRegistration":true,"contextSupport":true,"signatureInformation":{"documentationFormat":["markdown","plaintext"],"activeParameterSupport":true,"parameterInformation":{"labelOffsetSupport":true}}},"definition":{"dynamicRegistration":true},"references":{"dynamicRegistration":true},"documentHighlight":{"dynamicRegistration":true},"documentSymbol":{"dynamicRegistration":true,"symbolKind":{"valueSet":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26]},"hierarchicalDocumentSymbolSupport":true,"tagSupport":{"valueSet":[1]}},"codeAction":{"dynamicRegistration":true,"isPreferredSupport":true,"codeActionLiteralSupport":{"codeActionKind":{"valueSet":["","quickfix","refactor","refactor.extract","refactor.inline","refactor.rewrite","source","source.organizeImports"]}}},"codeLens":{"dynamicRegistration":true},"formatting":{"dynamicRegistration":true},"rangeFormatting":{"dynamicRegistration":true},"onTypeFormatting":{"dynamicRegistration":true},"rename":{"dynamicRegistration":true,"prepareSupport":true},"documentLink":{"dynamicRegistration":true,"tooltipSupport":true},"typeDefinition":{"dynamicRegistration":true},"implementation":{"dynamicRegistration":true},"declaration":{"dynamicRegistration":true},"colorProvider":{"dynamicRegistration":true},"foldingRange":{"dynamicRegistration":true,"rangeLimit":5000,"lineFoldingOnly":true},"selectionRange":{"dynamicRegistration":true}},"window":{"workDoneProgress":true}},"initializationOptions":{},"trace":"off","workspaceFolders":[{"uri":"file:///Users/adrian/github.com/a-h/templ","name":"templ"}],"clientInfo":{"name":"coc.nvim","version":"0.0.80"},"workDoneToken":"3ae4d8b2-f46a-46cc-886e-ab2ebafaa5e2"}`,
}

var rpcInitialized = RPC{
	Method:  "initialized",
	Payload: `{}`,
	Notify:  true,
}

var rpcDidOpen = RPC{
	Method:  "textDocument/didOpen",
	Payload: `{"textDocument":{"uri":"file:///Users/adrian/github.com/a-h/templ/cmd/example/address.go","languageId":"go","version":1,"text":"package templ\n\ntype Address struct {\n\tAddress1 string\n\tAddress2 string\n\tAddress3 string\n\tAddress4 string\n}\n\ntype Person struct {\n\t// Addresses of the person.\n\tAddresses []Address\n\t// Type of person.\n\tType string\n\t// URL of person.\n\tURL string\n}\n\n// ID of the person.\nfunc (p Person) ID() string {\n\treturn \"123\"\n}\n\n// Name of the person.\nfunc (p Person) Name() string {\n\treturn \"Person Name\"\n}\n\nfunc init() {\n\tp := Person{}\n\tp.ID()\n}\n"}}`,
	Notify:  true,
}

var rpcCompletion = RPC{
	Method:  "textDocument/completion",
	Payload: `{"textDocument":{"uri":"file:///Users/adrian/github.com/a-h/templ/cmd/example/address.go"},"position":{"line":29,"character":3},"context":{"triggerKind":2,"triggerCharacter":"."}}`,
}

var payloads []RPC = []RPC{
	rpcInitialize,
	rpcInitialized,
	rpcDidOpen,
	rpcCompletion,
}

type RPCHandler struct{}

func (h RPCHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, r *jsonrpc2.Request) {
	fmt.Println("<-- From server method:", r.Method)
	jr, err := r.Params.MarshalJSON()
	if err != nil {
		fmt.Println("RPCHandler.Handle: failed to marshal JSON", err)
	}
	fmt.Println("<-- From server request:", string(jr))
	// Return nil by default.
	err = conn.Reply(ctx, r.ID, nil)
	if err != nil {
		fmt.Println("RPCHandler.Handle: reply failed:", err)
	}
}

func runRPC() (s string, err error) {
	cmd := exec.Command("templ", "lsp")
	rwc, err := NewProcessReadWriteCloser(cmd)
	if err != nil {
		return
	}
	stream := jsonrpc2.NewBufferedStream(rwc, jsonrpc2.VSCodeObjectCodec{})
	conn := jsonrpc2.NewConn(context.Background(), stream, RPCHandler{})

	for i := 0; i < len(payloads); i++ {
		pl := payloads[i]
		fmt.Println("--> Method:", pl.Method)
		var input map[string]interface{}
		err = json.Unmarshal([]byte(pl.Payload), &input)
		if err != nil {
			err = fmt.Errorf("error unmarshalling JSON request: %w", err)
			return
		}
		fmt.Println("--> Payload:", pl.Payload)
		if pl.Notify {
			err = conn.Notify(context.Background(), pl.Method, &input)
			if err != nil {
				err = fmt.Errorf("error making %q notification: %w", pl.Method, err)
				return
			}
			continue
		}
		var output map[string]interface{}
		err = conn.Call(context.Background(), pl.Method, &input, &output)
		if err != nil {
			err = fmt.Errorf("error making %q call: %w", pl.Method, err)
			continue
		}
		jd, err := json.MarshalIndent(output, "", " ")
		if err != nil {
			err = fmt.Errorf("error unmarshalling JSON response: %w", err)
			continue
		}
		fmt.Println("<-- Response:", string(jd))
	}
	return "", rwc.Error
}

func New() (conn *jsonrpc2.Conn, err error) {
	cmd := exec.Command("gopls", "-logfile", "/Users/adrian/github.com/a-h/templ/cmd/lsp/proxy/main.txt", "-rpc.trace")
	rwc, err := NewProcessReadWriteCloser(cmd)
	if err != nil {
		return
	}
	stream := jsonrpc2.NewBufferedStream(rwc, jsonrpc2.VSCodeObjectCodec{})
	conn = jsonrpc2.NewConn(context.Background(), stream, RPCHandler{})
	return conn, err
}

func NewProcessReadWriteCloser(cmd *exec.Cmd) (rwc ProcessReadWriteCloser, err error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	rwc = ProcessReadWriteCloser{
		in:  stdin,
		out: stdout,
	}
	go func() {
		rwc.Error = cmd.Run()
	}()
	return
}

type ProcessReadWriteCloser struct {
	in    io.WriteCloser
	out   io.ReadCloser
	Error error
}

func (prwc ProcessReadWriteCloser) Read(p []byte) (n int, err error) {
	return prwc.out.Read(p)
}

func (prwc ProcessReadWriteCloser) Write(p []byte) (n int, err error) {
	return prwc.in.Write(p)
}

func (prwc ProcessReadWriteCloser) Close() error {
	errInClose := prwc.in.Close()
	errOutClose := prwc.out.Close()
	if errInClose != nil || errOutClose != nil {
		return fmt.Errorf("error closing process - in: %v, out: %v", errInClose, errOutClose)
	}
	return nil
}
