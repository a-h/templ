package pls

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"go.uber.org/zap"
)

// Options for the gopls client.
type Options struct {
	Log      string
	RPCTrace bool
}

// AsArguments converts the options into command line arguments for gopls.
func (opts Options) AsArguments() []string {
	var args []string
	if opts.Log != "" {
		args = append(args, "-logfile", opts.Log)
	}
	if opts.RPCTrace {
		args = append(args, "-rpc.trace")
	}
	return args
}

// NewGopls starts gopls and opens up a jsonrpc2 connection to it.
func NewGopls(ctx context.Context, log *zap.Logger, opts Options) (rwc io.ReadWriteCloser, err error) {
	_, err = exec.LookPath("gopls")
	if errors.Is(err, exec.ErrNotFound) {
		err = fmt.Errorf("cannot find gopls on the path (%q), you can install it with `go install golang.org/x/tools/gopls@latest`", os.Getenv("PATH"))
		return
	}
	if err != nil {
		return
	}
	cmd := exec.Command("gopls", opts.AsArguments()...)
	return newProcessReadWriteCloser(log, cmd)
}

// newProcessReadWriteCloser creates a processReadWriteCloser to allow stdin/stdout to be used as
// a JSON RPC 2.0 transport.
func newProcessReadWriteCloser(zapLogger *zap.Logger, cmd *exec.Cmd) (rwc processReadWriteCloser, err error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	rwc = processReadWriteCloser{
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

type processReadWriteCloser struct {
	in  io.WriteCloser
	out io.ReadCloser
}

func (prwc processReadWriteCloser) Read(p []byte) (n int, err error) {
	return prwc.out.Read(p)
}

func (prwc processReadWriteCloser) Write(p []byte) (n int, err error) {
	return prwc.in.Write(p)
}

func (prwc processReadWriteCloser) Close() error {
	errInClose := prwc.in.Close()
	errOutClose := prwc.out.Close()
	if errInClose != nil || errOutClose != nil {
		return fmt.Errorf("error closing process - in: %v, out: %v", errInClose, errOutClose)
	}
	return nil
}
