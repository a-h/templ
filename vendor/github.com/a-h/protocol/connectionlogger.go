package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"go.lsp.dev/jsonrpc2"
)

func NewConnectionLogger(w io.Writer, next jsonrpc2.Conn) *ConnectionLogger {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return &ConnectionLogger{
		w:    w,
		enc:  enc,
		next: next,
	}
}

type ConnectionLogger struct {
	w    io.Writer
	enc  *json.Encoder
	next jsonrpc2.Conn
}

func (cl *ConnectionLogger) Call(ctx context.Context, method string, params, result interface{}) (jsonrpc2.ID, error) {
	io.WriteString(cl.w, fmt.Sprintf("-> %s\n", method))
	cl.enc.Encode(params)
	var res json.RawMessage
	id, err := cl.next.Call(ctx, method, params, &res)
	if err != nil {
		io.WriteString(cl.w, fmt.Sprintf("<- %s %v\n", method, err))
		return id, err
	}
	if res != nil {
		io.WriteString(cl.w, fmt.Sprintf("<- %s\n", method))
		cl.enc.Encode(res)
	}
	err = json.Unmarshal(res, &result)
	return id, err
}

func (cl *ConnectionLogger) Notify(ctx context.Context, method string, params interface{}) error {
	io.WriteString(cl.w, fmt.Sprintf("-> %s\n", method))
	cl.enc.Encode(params)
	return cl.next.Notify(ctx, method, params)
}

func (cl *ConnectionLogger) Go(ctx context.Context, handler jsonrpc2.Handler) {
	cl.next.Go(ctx, handler)
}

func (cl *ConnectionLogger) Close() error {
	return cl.next.Close()
}

func (cl *ConnectionLogger) Done() <-chan struct{} {
	return cl.next.Done()
}

func (cl *ConnectionLogger) Err() error {
	return cl.next.Err()
}
