package run

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var m = &sync.Mutex{}
var running = map[string]*exec.Cmd{}

func Run(ctx context.Context, workingDir, input string) (cmd *exec.Cmd, err error) {
	m.Lock()
	defer m.Unlock()
	cmd, ok := running[input]
	if ok {
		if err = cmd.Process.Kill(); err != nil {
			return nil, fmt.Errorf("failed to kill existing process: %w", err)
		}
		delete(running, input)
	}

	parts := strings.SplitN(input, " ", 2)
	executable := parts[0]
	args := []string{}
	if len(parts) > 1 {
		args = append(args, parts[1])
	}

	cmd = exec.CommandContext(ctx, executable, args...)
	cmd.Env = os.Environ()
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	running[input] = cmd
	err = cmd.Start()
	return
}
