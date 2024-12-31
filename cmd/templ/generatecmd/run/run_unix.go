//go:build unix

package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	m       = &sync.Mutex{}
	running = map[string]*exec.Cmd{}
)

func KillAll() (err error) {
	m.Lock()
	defer m.Unlock()
	var errs []error
	for _, cmd := range running {
		if err := kill(cmd); err != nil {
			errs = append(errs, fmt.Errorf("failed to kill process %d: %w", cmd.Process.Pid, err))
		}
	}
	running = map[string]*exec.Cmd{}
	return errors.Join(errs...)
}

func kill(cmd *exec.Cmd) (err error) {
	errs := make([]error, 4)
	errs[0] = ignoreExited(cmd.Process.Signal(syscall.SIGINT))
	errs[1] = ignoreExited(cmd.Process.Signal(syscall.SIGTERM))
	errs[2] = ignoreExited(cmd.Wait())
	errs[3] = ignoreExited(syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL))
	return errors.Join(errs...)
}

func ignoreExited(err error) error {
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	// Ignore *exec.ExitError
	if _, ok := err.(*exec.ExitError); ok {
		return nil
	}
	return err
}

func Run(ctx context.Context, workingDir string, input string) (cmd *exec.Cmd, err error) {
	m.Lock()
	defer m.Unlock()
	cmd, ok := running[input]
	if ok {
		if err := kill(cmd); err != nil {
			return cmd, fmt.Errorf("failed to kill process %d: %w", cmd.Process.Pid, err)
		}

		delete(running, input)
	}
	parts := strings.Fields(input)
	executable := parts[0]
	args := []string{}
	if len(parts) > 1 {
		args = append(args, parts[1:]...)
	}

	cmd = exec.CommandContext(ctx, executable, args...)
	// Wait for the process to finish gracefully before termination.
	cmd.WaitDelay = time.Second * 3
	cmd.Env = os.Environ()
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	running[input] = cmd
	err = cmd.Start()
	return
}
