package pls

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

// Options for the gopls client.
type Options struct {
	Log      string
	RPCTrace bool
	Remote   string
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
	if opts.Remote != "" {
		args = append(args, "-remote", opts.Remote)
	}
	return args
}

func FindGopls() (location string, err error) {
	executableName := "gopls"
	if runtime.GOOS == "windows" {
		executableName = "gopls.exe"
	}

	pathLocation, err := exec.LookPath(executableName)
	if err == nil {
		// Found on the path.
		return pathLocation, nil
	}
	// Unexpected error.
	if !errors.Is(err, exec.ErrNotFound) {
		return "", fmt.Errorf("unexpected error looking for gopls: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unexpected error looking for gopls: %w", err)
	}

	// Probe standard locations.
	locations := []string{
		path.Join(home, "go", "bin", executableName),
		path.Join(home, ".local", "bin", executableName),
	}
	for _, location := range locations {
		_, err = os.Stat(location)
		if err != nil {
			continue
		}
		// Found in a standard location.
		return location, nil
	}

	return "", fmt.Errorf("cannot find gopls on the path (%q), in $HOME/go/bin or $HOME/.local/bin/gopls. You can install gopls with `go install golang.org/x/tools/gopls@latest`", os.Getenv("PATH"))
}

// GoplsVersion runs "gopls version" and returns the version string (e.g. "v0.21.1").
func GoplsVersion(location string) (version string, err error) {
	out, err := exec.Command(location, "version").Output()
	if err != nil {
		return "", fmt.Errorf("failed to run gopls version: %w", err)
	}
	// Output is like: "golang.org/x/tools/gopls v0.21.1"
	for _, line := range strings.Split(string(out), "\n") {
		_, v, ok := strings.Cut(line, "golang.org/x/tools/gopls ")
		if ok {
			return strings.TrimSpace(v), nil
		}
	}
	return "", fmt.Errorf("could not parse gopls version from output: %s", string(out))
}

// NewGopls starts gopls and opens up a jsonrpc2 connection to it.
func NewGopls(ctx context.Context, log *slog.Logger, opts Options) (location string, rwc io.ReadWriteCloser, err error) {
	location, err = FindGopls()
	if err != nil {
		return "", nil, err
	}
	cmd := exec.Command(location, opts.AsArguments()...)
	rwc, err = newProcessReadWriteCloser(log, cmd)
	return location, rwc, err
}

// newProcessReadWriteCloser creates a processReadWriteCloser to allow stdin/stdout to be used as
// a JSON RPC 2.0 transport.
func newProcessReadWriteCloser(logger *slog.Logger, cmd *exec.Cmd) (rwc processReadWriteCloser, err error) {
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
			logger.Error("gopls command error", slog.Any("error", err))
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
