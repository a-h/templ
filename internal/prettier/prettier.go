package prettier

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os/exec"
	"sync"
)


type Mode string

const (
	ModeJS  Mode = "js"
	ModeCSS Mode = "css"
)

type Prettier struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out io.ReadCloser
	mu  sync.Mutex
}

// New starts a background prettier process for the given mode (JS or CSS).
// Returns ok=false if prettier is not found or fails to start.
func New(mode Mode) (p *Prettier, ok bool) {
	path, err := exec.LookPath("prettier")
	if err != nil {
		return nil, false
	}
	var filename string
	switch mode {
	case ModeJS:
		filename = "file.js"
	case ModeCSS:
		filename = "file.css"
	default:
		return nil, false
	}
	cmd := exec.Command(path, "--stdin-filepath", filename)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, false
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		in.Close()
		return nil, false
	}
	if err := cmd.Start(); err != nil {
		in.Close()
		return nil, false
	}
	return &Prettier{
		cmd: cmd,
		in:  in,
		out: out,
	}, true
}

// Close shuts down the background prettier process.
func (p *Prettier) Close() error {
	if p == nil {
		return nil
	}
	p.in.Close()
	return p.cmd.Wait()
}

// Format formats code using prettier in the configured mode.
func (p *Prettier) Format(input string) (string, error) {
	if p == nil {
		return input, errors.New("prettier CLI not available")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return formatToProcess(p.in, p.out, input, p.cmd)
}

// formatToProcess writes input to the given stdin, reads output from stdout, and waits for the process to finish one format.
func formatToProcess(stdin io.WriteCloser, stdout io.ReadCloser, input string, cmd *exec.Cmd) (string, error) {
	_, err := io.WriteString(stdin, input)
	stdin.Close()
	if err != nil {
		return input, err
	}
	var buf bytes.Buffer
	s := bufio.NewScanner(stdout)
	for s.Scan() {
		buf.WriteString(s.Text())
		buf.WriteByte('\n')
	}
	cmd.Wait()
	return buf.String(), nil
}
