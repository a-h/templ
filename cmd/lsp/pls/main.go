package pls

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type RPCHandler struct {
	log             *zap.Logger
	onServerRequest func(ctx context.Context, conn *jsonrpc2.Conn, r *jsonrpc2.Request)
}

func (h RPCHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, r *jsonrpc2.Request) {
	h.onServerRequest(ctx, conn, r)
}

func NewGopls(zapLogger *zap.Logger, onServerRequest func(ctx context.Context, conn *jsonrpc2.Conn, r *jsonrpc2.Request)) (conn *jsonrpc2.Conn, err error) {
	//TODO: Configure the log location.
	cmd := exec.Command("gopls", "-logfile", "/Users/adrian/github.com/a-h/templ/cmd/lsp/gopls-log.txt", "-rpc.trace")
	rwc, err := NewProcessReadWriteCloser(zapLogger, cmd)
	if err != nil {
		return
	}
	stream := jsonrpc2.NewBufferedStream(rwc, jsonrpc2.VSCodeObjectCodec{})
	handler := RPCHandler{log: zapLogger, onServerRequest: onServerRequest}
	conn = jsonrpc2.NewConn(context.Background(), stream, handler)
	return
}

func NewProcessReadWriteCloser(zapLogger *zap.Logger, cmd *exec.Cmd) (rwc ProcessReadWriteCloser, err error) {
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
		if err := cmd.Run(); err != nil {
			zapLogger.Error("gopls command error", zap.Error(err))
		}
	}()
	return
}

type ProcessReadWriteCloser struct {
	in  io.WriteCloser
	out io.ReadCloser
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
